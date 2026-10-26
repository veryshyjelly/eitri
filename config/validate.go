package config

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"

	"github.com/google/shlex"
	"github.com/veryshyjelly/eitri/tmpl"
)

const configDocsURL = "https://github.com/veryshyjelly/eitri#-configuration-eitriyml"

// Issue describes a single configuration problem.
type Issue struct {
	Path    string
	Message string
	Hint    string
}

// ValidationError collects configuration issues found during startup validation.
type ValidationError struct {
	ConfigPath string
	Issues     []Issue
}

func (e *ValidationError) Error() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Error: Invalid configuration file (%s)\n", e.ConfigPath))
	for _, issue := range e.Issues {
		b.WriteString("\n  ")
		b.WriteString(issue.Path)
		b.WriteString(": ")
		b.WriteString(issue.Message)
		if issue.Hint != "" {
			b.WriteString("\n    hint: ")
			b.WriteString(issue.Hint)
		}
	}
	b.WriteString("\n\nSee: ")
	b.WriteString(configDocsURL)
	b.WriteString("\n")
	return b.String()
}

// Validate checks that the configuration is complete and usable before startup.
func Validate(cfg Config, root string, configPath string) error {
	var issues []Issue
	if stat, err := os.Stat(root); err != nil {
		issues = append(issues, Issue{Path: "root", Message: err.Error()})
	} else if !stat.IsDir() {
		issues = append(issues, Issue{Path: "root", Message: "must be a directory"})
	}

	issues = append(issues, validateFilenameRules(cfg, root)...)
	issues = append(issues, validateTemplateSection(cfg, root)...)
	issues = append(issues, validateCompilerSection(cfg)...)
	issues = append(issues, validateLibSection(cfg, root)...)
	issues = append(issues, validateOptionalTemplates(cfg)...)

	if len(issues) == 0 {
		return nil
	}
	return &ValidationError{
		ConfigPath: configPath,
		Issues:     issues,
	}
}

func validateFilenameRules(cfg Config, root string) []Issue {
	var issues []Issue

	if len(cfg.Rules) == 0 {
		issues = append(issues, Issue{
			Path:    "filename.rules",
			Message: "must contain at least one rule",
			Hint:    "add a rule with site, regex, and template fields",
		})
		return issues
	}

	for i, rule := range cfg.Rules {
		path := rulePath(i, rule.Site)
		if strings.TrimSpace(rule.Site) == "" {
			issues = append(issues, Issue{
				Path:    path,
				Message: "missing required field 'site'",
				Hint:    "site: codeforces.com",
			})
		}

		if strings.TrimSpace(rule.Regex) == "" {
			issues = append(issues, Issue{
				Path:    path,
				Message: "missing required field 'regex'",
				Hint:    "regex: \"problem/([^/?]+)\"",
			})
			continue
		}

		compiled, err := regexp.Compile(rule.Regex)
		if err != nil {
			issues = append(issues, Issue{
				Path:    path,
				Message: fmt.Sprintf("invalid regex: %v", err),
			})
			continue
		}

		if strings.TrimSpace(rule.Template) == "" {
			issues = append(issues, Issue{
				Path:    path,
				Message: "missing required field 'template'",
				Hint:    "template: './src/bin/{{ index .Captures 1 }}.rs'",
			})
			continue
		}

		issues = append(issues, validateFilenameTemplate(path, rule.Template, compiled, root)...)
	}

	return issues
}

func validateFilenameTemplate(path string, templateStr string, regex *regexp.Regexp, root string) []Issue {
	var issues []Issue

	tmpl, err := template.New("filename").Funcs(tmpl.FuncMap).Parse(templateStr)
	if err != nil {
		issues = append(issues, Issue{
			Path:    path,
			Message: fmt.Sprintf("invalid template: %v", err),
		})
		return issues
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, map[string]any{
		"Captures": dummyCapturesForRegex(regex),
		"Title":    "Example Problem",
	})
	if err != nil {
		issues = append(issues, Issue{
			Path:    path,
			Message: fmt.Sprintf("template execution failed: %v", err),
		})
		return issues
	}

	rendered := strings.TrimSpace(buffer.String())
	if rendered == "" {
		issues = append(issues, Issue{
			Path:    path,
			Message: "template renders to an empty path",
			Hint:    "include a filename with extension, e.g. './src/bin/{{ index .Captures 1 }}.rs'",
		})
		return issues
	}

	if strings.HasSuffix(rendered, "/") || strings.HasSuffix(rendered, "\\") {
		issues = append(issues, Issue{
			Path:    path,
			Message: fmt.Sprintf("template renders to a directory path '%s'", rendered),
			Hint:    "remove trailing path separator and include a filename with extension",
		})
		return issues
	}

	base := filepath.Base(rendered)
	if base == "." || base == ".." {
		issues = append(issues, Issue{
			Path:    path,
			Message: fmt.Sprintf("template renders to invalid path '%s'", rendered),
			Hint:    "include a filename with extension",
		})
		return issues
	}

	if filepath.Ext(base) == "" {
		issues = append(issues, Issue{
			Path:    path,
			Message: fmt.Sprintf("template renders to path without file extension '%s'", rendered),
			Hint:    "add a file extension, e.g. '.rs' or '.cpp'",
		})
	}

	fullPath := ResolvePath(root, rendered)
	stat, err := os.Stat(fullPath)
	if err == nil && stat.IsDir() {
		issues = append(issues, Issue{
			Path:    path,
			Message: fmt.Sprintf("template renders to existing directory '%s'", rendered),
			Hint:    "use a file path instead of an existing directory",
		})
	}

	return issues
}

func validateTemplateSection(cfg Config, root string) []Issue {
	var issues []Issue

	if strings.TrimSpace(cfg.Template.Source) == "" {
		if !cfg.CreateFile {
			return validateTemplateString("template.modifier", cfg.Template.Modifier)
		}
		issues = append(issues, Issue{
			Path:    "template.source",
			Message: "missing required field 'source'",
			Hint:    "source: ./src/main.rs",
		})
		return issues
	}

	sourcePath := ResolvePath(root, cfg.Template.Source)
	stat, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			issues = append(issues, Issue{
				Path:    "template.source",
				Message: fmt.Sprintf("file does not exist: %s", cfg.Template.Source),
				Hint:    "create the template file or update the path",
			})
		} else {
			issues = append(issues, Issue{
				Path:    "template.source",
				Message: fmt.Sprintf("cannot access file: %v", err),
			})
		}
	} else if stat.IsDir() {
		issues = append(issues, Issue{
			Path:    "template.source",
			Message: fmt.Sprintf("path is a directory, not a file: %s", cfg.Template.Source),
			Hint:    "point to a template source file",
		})
	}

	if strings.TrimSpace(cfg.Template.Modifier) != "" {
		issues = append(issues, validateTemplateString("template.modifier", cfg.Template.Modifier)...)
	}

	return issues
}

func validateCompilerSection(cfg Config) []Issue {
	var issues []Issue

	if strings.TrimSpace(cfg.Compiler.Source) == "" {
		issues = append(issues, Issue{
			Path:    "compiler.source",
			Message: "missing required field 'source'",
			Hint:    "source: main.rs",
		})
	}

	if strings.TrimSpace(cfg.Compiler.Run) == "" {
		issues = append(issues, Issue{
			Path:    "compiler.run",
			Message: "missing required field 'run'",
			Hint:    "run: ./main",
		})
	} else if args, err := shlex.Split(cfg.Compiler.Run); err != nil || len(args) == 0 || args[0] == "" {
		issues = append(issues, Issue{
			Path:    "compiler.run",
			Message: fmt.Sprintf("invalid or empty run command: %q (%v)", cfg.Compiler.Run, err),
		})
	}

	if cfg.Compiler.Source != "" && !LocalFile(cfg.Compiler.Source) {
		issues = append(issues, Issue{Path: "compiler.source", Message: "must name a relative file inside the temporary workspace"})
	}
	if cfg.CompiledFile != "" && !LocalFile(cfg.CompiledFile) {
		issues = append(issues, Issue{Path: "compiled_file", Message: "must name a relative file inside the temporary workspace"})
	}

	return issues
}

func validateLibSection(cfg Config, root string) []Issue {
	var issues []Issue
	if len(cfg.Lib.Include) > 0 && strings.TrimSpace(cfg.Lib.Regex) == "" {
		issues = append(issues, Issue{Path: "lib.regex", Message: "required when lib.include is set"})
	}

	if strings.TrimSpace(cfg.Lib.Regex) != "" {
		issues = append(issues, validateTemplateString("lib.regex", cfg.Lib.Regex)...)
	}

	for _, name := range slices.Sorted(maps.Keys(cfg.Lib.Include)) {
		location := cfg.Lib.Include[name]
		includePath := ResolvePath(root, location)
		stat, err := os.Stat(includePath)
		if err != nil {
			if os.IsNotExist(err) {
				issues = append(issues, Issue{
					Path:    fmt.Sprintf("lib.include.%s", name),
					Message: fmt.Sprintf("path does not exist: %s", location),
					Hint:    "create the library file or directory, or update the path",
				})
			} else {
				issues = append(issues, Issue{
					Path:    fmt.Sprintf("lib.include.%s", name),
					Message: fmt.Sprintf("cannot access path: %v", err),
				})
			}
			continue
		}

		if !stat.IsDir() && !stat.Mode().IsRegular() {
			issues = append(issues, Issue{
				Path:    fmt.Sprintf("lib.include.%s", name),
				Message: fmt.Sprintf("path is not a regular file or directory: %s", location),
			})
		}
	}

	if len(issues) == 0 && len(cfg.Lib.Include) > 0 {
		paths, err := cfg.LibraryPaths(root)
		if err != nil {
			return append(issues, Issue{Path: "lib.include", Message: err.Error()})
		}
		expression, err := template.New("lib.regex").Funcs(tmpl.FuncMap).Option("missingkey=error").Parse(cfg.Lib.Regex)
		if err != nil {
			return append(issues, Issue{Path: "lib.regex", Message: err.Error()})
		}
		names := slices.Sorted(maps.Keys(paths))
		if len(names) == 0 {
			names = []string{"example"}
		}
		for _, name := range names {
			var b bytes.Buffer
			if err := expression.Execute(&b, map[string]any{"Name": name}); err != nil {
				issues = append(issues, Issue{Path: "lib.regex", Message: err.Error()})
				continue
			}
			if _, err := regexp.Compile(b.String()); err != nil {
				issues = append(issues, Issue{Path: "lib.regex", Message: fmt.Sprintf("invalid regex for %s: %v", name, err)})
			}
		}
	}

	return issues
}

func validateOptionalTemplates(cfg Config) []Issue {
	var issues []Issue

	if strings.TrimSpace(cfg.Editor) != "" {
		t, err := template.New("editor").Funcs(tmpl.FuncMap).Option("missingkey=error").Parse(cfg.Editor)
		if err == nil {
			var b bytes.Buffer
			err = t.Execute(&b, map[string]any{"Filename": "/example/source"})
			if err == nil {
				var args []string
				args, err = shlex.Split(b.String())
				if err == nil && (len(args) == 0 || args[0] == "") {
					err = fmt.Errorf("editor command is empty")
				}
			}
		}
		if err != nil {
			issues = append(issues, Issue{Path: "editor", Message: err.Error()})
		}

	}

	if strings.TrimSpace(cfg.Code.Modifier) != "" {
		issues = append(issues, validateTemplateString("code.modifier", cfg.Code.Modifier)...)
	}

	return issues
}

func validateTemplateString(path string, value string) []Issue {
	_, err := template.New(path).Funcs(tmpl.FuncMap).Parse(value)
	if err != nil {
		return []Issue{{
			Path:    path,
			Message: fmt.Sprintf("invalid template: %v", err),
		}}
	}
	return nil
}

func rulePath(index int, site string) string {
	if strings.TrimSpace(site) == "" {
		return fmt.Sprintf("filename.rules[%d]", index)
	}
	return fmt.Sprintf("filename.rules[%d] (site: %s)", index, site)
}

func dummyCapturesForRegex(regex *regexp.Regexp) []string {
	names := regex.SubexpNames()
	captures := make([]string, len(names))
	for i := range captures {
		if i == 0 {
			captures[i] = "https://example.com/problem/123"
			continue
		}
		captures[i] = fmt.Sprintf("capture%d", i)
	}
	return captures
}
