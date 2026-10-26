package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/veryshyjelly/eitri/config"
)

func TestLinkFiles(t *testing.T) {
	for _, tt := range []struct {
		name, source string
		files        map[string]string
		want         []string
		err          string
	}{
		{"diamond and repeated roots", "use a; use b; use a", map[string]string{"a": "use c", "b": "use c", "c": "leaf", "unused": "use unused"}, []string{"c", "a", "b"}, ""},
		{"transitive", "use a", map[string]string{"a": "use b", "b": "use c", "c": "leaf"}, []string{"c", "b", "a"}, ""},
		{"unreferenced cycle", "nothing", map[string]string{"a": "use b", "b": "use a"}, nil, ""},
		{"cycle", "use a", map[string]string{"a": "use b", "b": "use c", "c": "use a"}, nil, "a -> b -> c -> a"},
		{"self cycle", "use a", map[string]string{"a": "use a"}, nil, "a -> a"},
		{"empty libraries", "anything", nil, nil, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for range 30 {
				got, err := linkFiles(tt.source, tt.files, `use {{.Name}}\b`)
				if tt.err != "" {
					if err == nil || !strings.Contains(err.Error(), tt.err) {
						t.Fatalf("got %v, want %s", err, tt.err)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				var names []string
				for _, lib := range got {
					names = append(names, lib.Name)
				}
				if !reflect.DeepEqual(names, tt.want) {
					t.Fatalf("got %v, want %v", names, tt.want)
				}
			}
		})
	}
	for _, pattern := range []string{`[{{.Name}}`, `{{ .Missing }}`, `{{`} {
		if _, err := linkFiles("a", map[string]string{"a": "x"}, pattern); err == nil {
			t.Fatalf("accepted pattern %s", pattern)
		}
	}
}

func TestExtractBlock(t *testing.T) {
	for _, tt := range []struct {
		name, source, want string
		fail               bool
	}{
		{"absent", " full\n source ", " full\n source ", false},
		{"marked", "outside\n// @code begin\n value\n// @code end\noutside", "value", false},
		{"empty", "// @code begin\n// @code end", "", false},
		{"adjacent comment", "//@code begin\nvalue\n//@code end", "value", false},
		{"invalid action", "// @code beginning", "", true},
		{"incomplete marker", "// @code", "", true},
		{"missing end", "// @code begin\nvalue", "", true},
		{"missing begin", "value\n// @code end", "", true},
		{"nested", "// @code begin\n// @code begin\n// @code end", "", true},
		{"repeated", "// @code begin\n// @code end\n// @code begin\n// @code end", "", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractBlock(tt.source, "code", tt.source)
			if (err != nil) != tt.fail || (!tt.fail && got != tt.want) {
				t.Fatalf("got %q, %v", got, err)
			}
		})
	}
}

func fixtureApp(t *testing.T) App {
	t.Helper()
	root := t.TempDir()
	return App{Root: root, Config: config.Config{
		Filename: config.Filename{Rules: []config.Rule{{Site: "example.com", Regex: `problem/(\w+)`, Template: `nested/{{index .Captures 1}}.txt`}}},
		Compiler: config.Compiler{Source: "source.txt"},
	}, Problem: Problem{Title: "Example", Url: "https://example.com/problem/one", TimeLimit: 2000}}
}
func writeSource(t *testing.T, app App, text string) {
	t.Helper()
	name, err := app.FileName()
	if err != nil {
		t.Fatal(err)
	}
	path := config.ResolvePath(app.Root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestSolution(t *testing.T) {
	app := fixtureApp(t)
	writeSource(t, app, "// @head begin\nheader\n// @head end\n// @code begin\nmain uses a\n// @code end")
	lib := filepath.Join(app.Root, "a.lib")
	if err := os.WriteFile(lib, []byte("// @head begin\nheader\nother\n// @head end\n// @code begin\nlibrary\n// @code end"), 0644); err != nil {
		t.Fatal(err)
	}
	app.Lib = config.Lib{Regex: `uses {{.Name}}\b`, Include: map[string]string{"a": lib}}
	app.Code.Modifier = `{{.Header}}|{{range .LibFiles}}{{.Name}}={{.Content}}{{end}}|{{.Code}}`
	for range 20 {
		got, err := app.Solution()
		if err != nil || got != "header\nother|a=library|main uses a" {
			t.Fatalf("got %q, %v", got, err)
		}
	}
	app.Lib.Include["missing"] = "missing.lib"
	if _, err := app.Solution(); err == nil {
		t.Fatal("missing library accepted")
	}
}

func TestCreateFileAndRouting(t *testing.T) {
	app := fixtureApp(t)
	app.Template.Modifier = `{{.Title}} {{.Filename}} {{.Code}}`
	if err := app.CreateFile(); err != nil {
		t.Fatal(err)
	}
	if got, err := app.Solution(); err != nil || got != "Example one " {
		t.Fatalf("got %q, %v", got, err)
	}
	writeSource(t, app, "edited")
	if err := app.CreateFile(); err != nil {
		t.Fatal(err)
	}
	if got, _ := app.Solution(); got != "edited" {
		t.Fatal("overwrote source")
	}
	app.Rules = append([]config.Rule{{Site: "example.com", Regex: "unmatched", Template: "wrong"}}, app.Rules...)
	if name, err := app.FileName(); err != nil || name != filepath.Join("nested", "one.txt") {
		t.Fatalf("%s, %v", name, err)
	}
	app.Url = "https://example.com/other"
	if _, err := app.FileName(); err == nil {
		t.Fatal("unmatched route accepted")
	}
}

func TestLibraryDuplicateNames(t *testing.T) {
	app := fixtureApp(t)
	writeSource(t, app, "source")
	for _, file := range []string{"a.txt", "a.rs"} {
		if err := os.WriteFile(filepath.Join(app.Root, file), nil, 0644); err != nil {
			t.Fatal(err)
		}
	}
	app.Lib.Include = map[string]string{"_": "."}
	if _, err := app.libraries(); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("got %v", err)
	}
}

func BenchmarkLinkFiles(b *testing.B) {
	files := map[string]string{}
	for i := range 100 {
		files[fmt.Sprintf("lib%03d", i)] = fmt.Sprintf("use lib%03d", i+1)
	}
	for b.Loop() {
		if _, err := linkFiles("use lib000", files, `use {{.Name}}\b`); err != nil {
			b.Fatal(err)
		}
	}
}

func TestEditorPathWithSpaces(t *testing.T) {
	app := fixtureApp(t)
	app.Root = filepath.Join(app.Root, "space in workspace")
	app.Editor = `editor --wait "{{ .Filename }}"`
	command, err := app.EditorCommand(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(app.Root, "nested", "one.txt")
	if len(command.Args) != 3 || command.Args[2] != want || command.Dir != app.Root {
		t.Fatalf("args %q, dir %q", command.Args, command.Dir)
	}
	app.Editor = `{{ .Missing }}`
	if _, err := app.EditorCommand(context.Background()); err == nil {
		t.Fatal("invalid editor template accepted")
	}
}

func TestTemplateFailureLeavesNoSource(t *testing.T) {
	app := fixtureApp(t)
	app.Template.Modifier = `{{ .Missing }}`
	if err := app.CreateFile(); err == nil {
		t.Fatal("invalid boilerplate template accepted")
	}
	name, _ := app.FileName()
	if _, err := os.Stat(config.ResolvePath(app.Root, name)); !os.IsNotExist(err) {
		t.Fatal("failed render left a source file")
	}
}
