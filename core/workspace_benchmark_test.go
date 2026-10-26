package core

import (
	"fmt"
	"github.com/veryshyjelly/eitri/config"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkSolutionWorkspace(b *testing.B) {

	root := b.TempDir()
	files := map[string]string{}
	for i := range 100 {
		name := fmt.Sprintf("lib%03d", i)
		path := filepath.Join(root, name+".txt")
		if err := os.WriteFile(path, []byte(fmt.Sprintf("use lib%03d", i+1)), 0644); err != nil {
			b.Fatal(err)
		}
		files[name] = name + ".txt"
	}
	if err := os.WriteFile(filepath.Join(root, "main.txt"), []byte("use lib000"), 0644); err != nil {
		b.Fatal(err)
	}
	app := App{Root: root, Problem: Problem{Url: "https://example.com/a"}, Config: config.Config{
		Filename: config.Filename{Rules: []config.Rule{{Site: "example.com", Regex: ".*", Template: "main.txt"}}},
		Lib:      config.Lib{Include: files, Regex: `use {{.Name}}\b`},
		Code:     config.Code{Modifier: `{{.Header}}{{range .LibFiles}}{{.Content}}{{end}}{{.Code}}`},
	}}
	for b.Loop() {
		if _, err := app.Solution(); err != nil {
			b.Fatal(err)
		}
	}
}
