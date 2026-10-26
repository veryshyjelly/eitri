package core

import (
	"fmt"
	"strings"
)

// extractBlock accepts marker-free text, but rejects incomplete, nested,
// reversed, or repeated marker pairs. Empty marked blocks are intentionally empty.
func extractBlock(source, tag, fallback string) (string, error) {
	var content []string
	active, done := false, false
	for _, line := range strings.Split(source, "\n") {
		marker := ""
		if index := strings.Index(line, "@"+tag); index >= 0 {
			tail := line[index+len(tag)+1:]
			if tail == "" || tail[0] == ' ' || tail[0] == '\t' || tail[0] == '\r' {
				fields := strings.Fields(tail)
				if len(fields) == 0 || (fields[0] != "begin" && fields[0] != "end") {
					return "", fmt.Errorf("@%s: expected begin or end marker", tag)
				}
				marker = fields[0]
			}
		}

		switch marker {
		case "begin":
			if active || done {
				return "", fmt.Errorf("@%s: repeated begin marker", tag)
			}
			active = true
		case "end":
			if !active {
				return "", fmt.Errorf("@%s: end without begin", tag)
			}
			active, done = false, true
		default:
			if active {
				content = append(content, line)
			}
		}
	}
	if active {
		return "", fmt.Errorf("@%s: missing end marker", tag)
	}
	if !done {
		return fallback, nil
	}
	return strings.TrimSpace(strings.Join(content, "\n")), nil
}
