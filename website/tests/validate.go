package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/veryshyjelly/eitri/config"
	"github.com/veryshyjelly/eitri/core"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type fixture struct{ Lang, Os, Ext, Config string }

func main() {
	data, err := os.ReadFile("website/dist/fixtures.json")
	must(err)
	var cases []fixture
	must(json.Unmarshal(data, &cases))
	for _, f := range cases {
		root, err := os.MkdirTemp("", "eitri-guide-")
		must(err)
		func() {
			defer os.RemoveAll(root)
			must(os.Mkdir(filepath.Join(root, "solutions"), 0755))
			source := filepath.Join(root, "base."+f.Ext)
			must(os.WriteFile(source, nil, 0644))
			path := filepath.Join(root, "eitri.yml")
			must(os.WriteFile(path, []byte(f.Config), 0644))
			cfg, err := config.ReadConfig(path)
			must(err)
			must(config.Validate(cfg, root, path))
			must(os.Remove(source))
			if config.Validate(cfg, root, path) == nil {
				panic("missing base file accepted")
			}
			must(os.WriteFile(source, nil, 0644))
			cfg.Lib.Include = map[string]string{"missing": "does-not-exist"}
			if config.Validate(cfg, root, path) == nil {
				panic("missing include accepted")
			}
		}()
	}
	fmt.Printf("Validated %d guide configurations; missing-file and include-path regressions rejected.\n", len(cases))
	validateRecipes()
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}

type recipe struct {
	ID, Config, Template, TemplateSource string
	Paths                                []string
}
type transformation struct{ ID, Name, LibraryName, Library, Source, Expected string }

func normalized(s string) string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if v := strings.TrimSpace(line); v != "" {
			lines = append(lines, v)
		}
	}
	return strings.Join(lines, "\n")
}
func validateRecipes() {
	data, err := os.ReadFile("website/tests/fixtures/setup-sources.json")
	must(err)
	var recipes []recipe
	must(json.Unmarshal(data, &recipes))
	byID := map[string]config.Config{}
	for _, r := range recipes {
		root, err := os.MkdirTemp("", "eitri-recipe-")
		must(err)
		func() {
			defer os.RemoveAll(root)
			for _, p := range r.Paths {
				path := filepath.Join(root, p)
				must(os.MkdirAll(filepath.Dir(path), 0755))
				must(os.WriteFile(path, nil, 0644))
			}
			must(os.WriteFile(filepath.Join(root, r.Template), []byte(r.TemplateSource), 0644))
			path := filepath.Join(root, "eitri.yml")
			must(os.WriteFile(path, []byte(r.Config), 0644))
			cfg, err := config.ReadConfig(path)
			must(err)
			must(config.Validate(cfg, root, path))
			byID[r.ID] = cfg
		}()
	}
	fmt.Printf("Validated %d recipe configurations and referenced-path manifests (library placeholders; no full-project compilation).\n", len(recipes))
	data, err = os.ReadFile("website/tests/fixtures/transformations.json")
	must(err)
	var fixtures []transformation
	must(json.Unmarshal(data, &fixtures))
	for _, f := range fixtures {
		root, err := os.MkdirTemp("", "eitri-transform-")
		must(err)
		func() {
			defer os.RemoveAll(root)
			cfg := byID[f.ID]
			cfg.Rules = []config.Rule{{Site: "example.com", Regex: "/problem", Template: "edited" + filepath.Ext(cfg.Compiler.Source)}}
			cfg.Template.Source = "base"
			cfg.Lib.Include = map[string]string{f.LibraryName: "snippet"}
			must(os.WriteFile(filepath.Join(root, "base"), nil, 0644))
			must(os.WriteFile(filepath.Join(root, "edited"+filepath.Ext(cfg.Compiler.Source)), []byte(f.Source), 0644))
			must(os.WriteFile(filepath.Join(root, "snippet"), []byte(f.Library), 0644))
			must(config.Validate(cfg, root, "fixture"))
			app := core.App{Root: root, Config: cfg, Problem: core.Problem{Url: "https://example.com/problem"}}
			actual, err := app.Solution()
			must(err)
			if normalized(actual) != normalized(f.Expected) {
				panic(fmt.Sprintf("%s transformation mismatch\nGOT:\n%s\nWANT:\n%s", f.ID, actual, f.Expected))
			}
			fmt.Println(f.Name + ": prepared output matches handbook.")
			if _, err := exec.LookPath(cfg.Compile); err != nil {
				fmt.Println(f.ID + ": compiler unavailable; execution unverified.")
				return
			}
			answer := "8"
			if f.ID == "llvm" {
				answer = "OK"
			}
			app.Tests = []core.Testcase{{Input: "", Answer: answer}}
			results, err := app.Run(context.Background())
			must(err)
			if len(results) != 1 || results[0].Status != core.Accepted {
				panic(fmt.Sprintf("%s generated-source run failed: %+v", f.ID, results))
			}
			fmt.Println(f.ID + ": generated-source compile/run passed.")
		}()
	}
}
