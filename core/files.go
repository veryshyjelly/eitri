package core

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/google/shlex"
	"github.com/veryshyjelly/eitri/config"
	"github.com/veryshyjelly/eitri/tmpl"
)

func render(name, text string, data any) (string, error) {
	t, err := template.New(name).Funcs(tmpl.FuncMap).Option("missingkey=error").Parse(text)
	if err != nil {
		return "", fmt.Errorf("%s: %w", name, err)
	}
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return "", fmt.Errorf("%s: %w", name, err)
	}
	return b.String(), nil
}

// FileName routes a problem URL through the configured rules. Rules for the same
// host are tried in configuration order until the URL expression matches.
func (app App) FileName() (string, error) {
	u, err := url.Parse(app.Url)
	if err != nil || u.Hostname() == "" {
		return "", fmt.Errorf("invalid problem URL %q", app.Url)
	}
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	for _, rule := range app.Rules {
		if strings.TrimPrefix(strings.ToLower(rule.Site), "www.") != host {
			continue
		}
		re, err := regexp.Compile(rule.Regex)
		if err != nil {
			return "", fmt.Errorf("filename regex for %s: %w", host, err)
		}
		captures := re.FindStringSubmatch(app.Url)
		if captures == nil {
			continue
		}
		name, err := render("filename", rule.Template, map[string]any{
			"Captures": captures, "Title": strings.NewReplacer("'", "", "\"", "").Replace(app.Title),
		})
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(name) == "" || strings.HasSuffix(name, "/") || strings.HasSuffix(name, "\\") || filepath.Clean(name) == "." {
			return "", fmt.Errorf("filename template produced an invalid file path %q", name)
		}
		return filepath.Clean(name), nil
	}
	return "", fmt.Errorf("no filename rule matches %s", app.Url)
}

// CreateFile renders the entire boilerplate before exclusively creating the
// destination. Existing source files are never overwritten.
func (app App) CreateFile() error {
	name, err := app.FileName()
	if err != nil {
		return err
	}
	path := config.ResolvePath(app.Root, name)
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("source: %w", err)
	}
	var source []byte
	if app.Template.Source != "" {
		source, err = os.ReadFile(config.ResolvePath(app.Root, app.Template.Source))
		if err != nil {
			return fmt.Errorf("read boilerplate: %w", err)
		}
	}
	modifier := app.Template.Modifier
	if modifier == "" {
		modifier = "{{ .Code }}"
	}
	content, err := render("template.modifier", modifier, map[string]any{
		"Author": app.Author, "Time": time.Now().Format("2006/01/02 15:04"),
		"Url": app.Url, "Title": app.Title,
		"Filename": strings.TrimSuffix(filepath.Base(name), filepath.Ext(name)), "Code": string(source),
	})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create source directory: %w", err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if os.IsExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("create source: %w", err)
	}
	_, writeErr := f.WriteString(content)
	closeErr := f.Close()
	if writeErr != nil {
		_ = os.Remove(path)
		return fmt.Errorf("write source: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close source: %w", closeErr)
	}
	return nil
}

// EditorCommand prepares a direct argv invocation; shell operators are never
// evaluated implicitly. The filename retains its literal spaces and backslashes.
func (app App) EditorCommand(ctx context.Context) (*exec.Cmd, error) {
	name, err := app.FileName()
	if err != nil {
		return nil, err
	}
	path, err := filepath.Abs(config.ResolvePath(app.Root, name))
	if err != nil {
		return nil, err
	}
	command, err := render("editor", app.Editor, map[string]any{"Filename": path})
	if err != nil {
		return nil, err
	}
	args, err := splitEditorCommand(command, path)
	if err != nil {
		return nil, fmt.Errorf("editor command: %w", err)
	}
	if len(args) == 0 || args[0] == "" {
		return nil, fmt.Errorf("editor command is empty")
	}
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = app.Root
	return cmd, nil
}

// splitEditorCommand protects the rendered filename while parsing the command.
// Inserting it after tokenization preserves literal backslashes and quotes (and
// keeps an unquoted filename containing spaces in a single argument). Render
// templates with the real filename first so template functions still see it.
func splitEditorCommand(command, filename string) ([]string, error) {
	marker := "__EITRI_FILENAME__"
	for strings.Contains(command, marker) {
		marker += "_"
	}
	args, err := shlex.Split(strings.ReplaceAll(command, filename, marker))
	if err != nil {
		return nil, err
	}
	for i := range args {
		args[i] = strings.ReplaceAll(args[i], marker, filename)
	}
	return args, nil
}

// libraries reads configured snippets after centralized path discovery.
func (app App) libraries() (map[string]string, error) {
	paths, err := app.Config.LibraryPaths(app.Root)
	if err != nil {
		return nil, err
	}
	files := make(map[string]string, len(paths))
	for _, name := range sortedKeys(paths) {
		content, err := os.ReadFile(paths[name])
		if err != nil {
			return nil, fmt.Errorf("library %s: %w", name, err)
		}
		files[name] = string(content)
	}
	return files, nil
}
