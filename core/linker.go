package core

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/veryshyjelly/eitri/config"
)

// Library is a named snippet exposed to code.modifier templates.
type Library struct{ Name, Content string }

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

// linkFiles compiles each detection expression once, then visits only reachable
// snippets. Sorted roots and edges give reproducible dependency-first order.
func linkFiles(source string, files map[string]string, pattern string) ([]Library, error) {
	names := sortedKeys(files)
	expressions := make(map[string]*regexp.Regexp, len(names))
	for _, name := range names {
		text, err := render("lib.regex", pattern, map[string]any{"Name": name})
		if err != nil {
			return nil, err
		}
		re, err := regexp.Compile(text)
		if err != nil {
			return nil, fmt.Errorf("lib.regex for %q: %w", name, err)
		}
		expressions[name] = re
	}
	state := make(map[string]uint8, len(names))
	var order []Library
	var stack []string
	var visit func(string) error
	visit = func(name string) error {
		if state[name] == 2 {
			return nil
		}
		if state[name] == 1 {
			start := slices.Index(stack, name)
			chain := append(slices.Clone(stack[start:]), name)
			return fmt.Errorf("cyclic library dependency: %s", strings.Join(chain, " -> "))
		}
		state[name] = 1
		stack = append(stack, name)
		for _, dep := range names {
			if expressions[dep].MatchString(files[name]) {
				if err := visit(dep); err != nil {
					return err
				}
			}
		}
		stack = stack[:len(stack)-1]
		state[name] = 2
		order = append(order, Library{Name: name, Content: files[name]})
		return nil
	}
	for _, name := range names {
		if expressions[name].MatchString(source) {
			if err := visit(name); err != nil {
				return nil, err
			}
		}
	}
	return order, nil
}

// Solution reads the source once and renders the linked submission. Time remains
// available to custom templates; templates using it intentionally vary over time.
func (app App) Solution() (string, error) {
	name, err := app.FileName()
	if err != nil {
		return "", err
	}
	source, err := os.ReadFile(config.ResolvePath(app.Root, name))
	if err != nil {
		return "", fmt.Errorf("read source: %w", err)
	}
	files, err := app.libraries()
	if err != nil {
		return "", err
	}
	libraries, err := linkFiles(string(source), files, app.Lib.Regex)
	if err != nil {
		return "", err
	}
	var headers []string
	seen := make(map[string]bool)
	addHeaders := func(text string) error {
		header, err := extractBlock(text, "head", "")
		if err != nil {
			return err
		}
		for _, line := range strings.Split(header, "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !seen[line] {
				seen[line] = true
				headers = append(headers, line)
			}
		}
		return nil
	}
	for i := range libraries {
		if err := addHeaders(libraries[i].Content); err != nil {
			return "", fmt.Errorf("library %s: %w", libraries[i].Name, err)
		}
		content, err := extractBlock(libraries[i].Content, "code", libraries[i].Content)
		if err != nil {
			return "", fmt.Errorf("library %s: %w", libraries[i].Name, err)
		}
		libraries[i].Content = content
	}
	if err := addHeaders(string(source)); err != nil {
		return "", fmt.Errorf("source: %w", err)
	}
	code, err := extractBlock(string(source), "code", string(source))
	if err != nil {
		return "", fmt.Errorf("source: %w", err)
	}
	modifier := app.Code.Modifier
	if modifier == "" {
		modifier = "{{ .Code }}"
	}
	return render("code.modifier", modifier, map[string]any{
		"Author": app.Author, "Url": app.Url, "Time": time.Now().Format("2006/01/02 15:04"),
		"LibFiles": libraries, "Header": strings.Join(headers, "\n"), "Code": code,
	})
}
