package tmpl

import (
	"strings"
	"text/template"

	"github.com/ettle/strcase"
)

// FuncMap is the template function map used for config rendering.
var FuncMap = template.FuncMap{
	"toKebabCase":  strcase.ToKebab,
	"toCamelCase":  strcase.ToCamel,
	"toSnakeCase":  strcase.ToSnake,
	"toPascalCase": strcase.ToPascal,
	"toKEBABCase":  strcase.ToKEBAB,
	"toSNAKECase":  strcase.ToSNAKE,
	"toLowerCase":  strings.ToLower,
	"toUpperCase":  strings.ToUpper,
	"stripPrefix":  stripPrefix,
	"removeLines":  removeLines,
}

// removeLines removes the first n lines from a given string.
func removeLines(n int, s string) string {
	if n <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	if n >= len(lines) {
		return ""
	}
	return strings.Join(lines[n:], "\n")
}

func stripPrefix(delim, title string) string {
	parts := strings.SplitN(title, delim, 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return title
}
