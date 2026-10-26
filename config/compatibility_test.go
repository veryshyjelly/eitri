package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/veryshyjelly/eitri/templates"
)

func TestEmbeddedConfigurations(t *testing.T) {
	entries, err := templates.FS.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		t.Run(entry.Name(), func(t *testing.T) {
			data, err := templates.FS.ReadFile(entry.Name())
			if err != nil {
				t.Fatal(err)
			}
			var cfg Config
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			path := ResolvePath(root, cfg.Template.Source)
			mustMkdirAll(t, filepath.Dir(path))
			mustWriteFile(t, path, "source")
			for _, path := range cfg.Lib.Include {
				path = ResolvePath(root, path)
				if filepath.Ext(path) == "" {
					mustMkdirAll(t, path)
				} else {
					mustMkdirAll(t, filepath.Dir(path))
					mustWriteFile(t, path, "library")
				}
			}
			if err := Validate(cfg, root, entry.Name()); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestOptionalSettingsAndValidation(t *testing.T) {
	for _, tt := range []struct {
		name   string
		change func(*Config)
		want   string
	}{
		{"interpreted", func(c *Config) { c.Compile = ""; c.Args = nil }, ""},
		{"no argument build", func(c *Config) { c.Args = nil }, ""},
		{"no boilerplate", func(c *Config) { c.CreateFile = false; c.Template = Template{} }, ""},
		{"absolute template", func(c *Config) {}, ""},
		{"invalid run quoting", func(c *Config) { c.Run = `"unterminated` }, "compiler.run"},
		{"empty argv", func(c *Config) { c.Run = `# comment` }, "compiler.run"},
		{"source escape", func(c *Config) { c.Compiler.Source = "../escape.rs" }, "compiler.source"},
		{"artifact escape", func(c *Config) { c.CompiledFile = "../escape.rs" }, "compiled_file"},
		{"regex execution", func(c *Config) { c.Lib.Regex = `[{{.Name}}` }, "lib.regex"},
		{"bad template", func(c *Config) { c.Code.Modifier = "{{" }, "code.modifier"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := setupValidRoot(t)
			cfg := validConfig()
			tt.change(&cfg)
			if tt.name == "absolute template" {
				cfg.Template.Source = filepath.Join(root, "src/main.rs")
			}
			err := Validate(cfg, root, "eitri.yml")
			if tt.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v, want %s", err, tt.want)
			}
		})
	}
}
func TestReadConfigAndPathResolution(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "eitri.yml")
	if _, err := ReadConfig(path); !os.IsNotExist(err) {
		t.Fatalf("missing: %v", err)
	}
	mustWriteFile(t, path, "compiler: [invalid")
	if _, err := ReadConfig(path); err == nil {
		t.Fatal("bad YAML accepted")
	}
	mustWriteFile(t, path, "compiler:\n  source: main.py\n  run: python3 main.py\nrun_on_save: true\n")
	cfg, err := ReadConfig(path)
	if err != nil || cfg.Compiler.Source != "main.py" || !cfg.RunOnSave {
		t.Fatalf("%+v, %v", cfg, err)
	}
	absolute := filepath.Join(root, "absolute")
	if ResolvePath(root, absolute) != absolute || ResolvePath(root, "relative") != filepath.Join(root, "relative") {
		t.Fatal("path base incorrect")
	}
}
