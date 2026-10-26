package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/atotto/clipboard"
)

// writeClipboard keeps external clipboard utilities within the app lifecycle.
// On Windows the library uses native APIs with a bounded clipboard-open retry.
func writeClipboard(ctx context.Context, text string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		return clipboard.WriteAll(text)
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var candidates [][]string
	if runtime.GOOS == "darwin" {
		candidates = [][]string{{"pbcopy"}}
	} else {
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			candidates = append(candidates, []string{"wl-copy"})
		}
		candidates = append(candidates, []string{"xclip", "-in", "-selection", "clipboard"}, []string{"xsel", "--input", "--clipboard"}, []string{"termux-clipboard-set"}, []string{"clip.exe"})
	}
	for _, args := range candidates {
		path, err := exec.LookPath(args[0])
		if err != nil {
			continue
		}
		cmd := exec.CommandContext(ctx, path, args[1:]...)
		cmd.Stdin = strings.NewReader(text)
		cmd.WaitDelay = 250 * time.Millisecond
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", args[0], err)
		}
		return nil
	}
	return fmt.Errorf("no clipboard utility available (install xclip, xsel, or wl-copy)")
}
