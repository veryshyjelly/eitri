package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The test binary supplies deterministic, portable compiler and runner fixtures.
// Helpers exit directly to avoid testing package output in captured stdout.
func TestProcessHelper(t *testing.T) {
	index := slices.Index(os.Args, "--eitri-helper")
	if index < 0 {
		return
	}
	mode := os.Args[index+1]
	switch mode {
	case "run-source":
		data, err := os.ReadFile("source.txt")
		if err != nil {
			os.Exit(1)
		}
		fmt.Fprint(os.Stdout, string(data))
	case "run-artifact":
		data, err := os.ReadFile("submission.txt")
		if err != nil {
			os.Exit(1)
		}
		fmt.Fprint(os.Stdout, strings.TrimPrefix(string(data), "built:"))
	case "echo":
		_, _ = io.Copy(os.Stdout, os.Stdin)
	case "stderr":
		fmt.Fprint(os.Stdout, "out")
		fmt.Fprint(os.Stderr, "diagnostic")
		os.Exit(7)
	case "sleep":
		time.Sleep(time.Minute)
	case "flood":
		for {
			fmt.Fprint(os.Stdout, strings.Repeat("x", 8192))
		}
	case "compile":
		source, err := os.ReadFile("source.txt")
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(2)
		}
		if err := os.WriteFile("submission.txt", append([]byte("built:"), source...), 0644); err != nil {
			os.Exit(3)
		}
	case "compile-fail":
		fmt.Fprint(os.Stderr, "compiler diagnostic")
		os.Exit(1)
	case "directory-artifact":
		if err := os.Mkdir("submission.txt", 0755); err != nil {
			os.Exit(1)
		}
	case "large-artifact":
		if err := os.WriteFile("submission.txt", []byte(strings.Repeat("x", maxOutput+1)), 0644); err != nil {
			os.Exit(1)
		}
	case "empty-artifact":
		if err := os.WriteFile("submission.txt", nil, 0644); err != nil {
			os.Exit(1)
		}
	case "record":
		dir, _ := os.Getwd()
		_ = os.WriteFile(os.Args[index+2], []byte(dir), 0644)
		if os.Args[index+3] == "fail" {
			os.Exit(1)
		}
	}
	os.Exit(0)
}
func helper(t *testing.T, mode string) (string, []string) {
	t.Helper()
	t.Setenv("GORACE", "atexit_sleep_ms=0")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return exe, []string{"-test.run=^TestProcessHelper$", "--", "--eitri-helper", mode}
}
func runner(t *testing.T, mode string) App {
	t.Helper()
	app := fixtureApp(t)
	writeSource(t, app, "source")
	exe, args := helper(t, mode)
	app.Config.Run = strconv.Quote(exe) + " " + strings.Join(args, " ")
	app.Tests = []Testcase{{Input: "1\n", Answer: "1"}, {Input: "2\n", Answer: "2\n"}}
	return app
}

func TestClassify(t *testing.T) {
	for _, tt := range []struct {
		output, answer string
		err            error
		timeout        bool
		want           Status
	}{
		{" \n42\n ", "42", nil, false, Accepted},
		{"1  2", "1 2", nil, false, WrongAnswer},
		{"1\n2", "1 2", nil, false, WrongAnswer},
		{"", "", nil, false, Accepted},
		{"", "42", nil, false, WrongAnswer},
		{"42", "42", errors.New("exit"), false, RuntimeError},
		{"42", "42", errors.New("exit"), true, TimeLimitExceeded},
	} {
		if got := classify(tt.output, tt.answer, tt.err, tt.timeout); got != tt.want {
			t.Fatalf("got %s, want %s", got, tt.want)
		}
	}
	if got := GetFinalStatus(nil); got != NotAvailable {
		t.Fatalf("empty samples: %s", got)
	}
}

func TestRun(t *testing.T) {
	for _, tt := range []struct {
		mode string
		want Status
	}{{"echo", Accepted}, {"stderr", RuntimeError}, {"sleep", TimeLimitExceeded}, {"flood", RuntimeError}} {
		t.Run(tt.mode, func(t *testing.T) {
			app := runner(t, tt.mode)
			if tt.mode == "sleep" {
				app.TimeLimit = 80
			}
			original := slices.Clone(app.Tests)
			got, err := app.Run(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 2 {
				t.Fatalf("got %d samples", len(got))
			}
			for i, test := range got {
				if test.Input != original[i].Input || test.Status != tt.want {
					t.Fatalf("sample %d: %+v", i, test)
				}
				if len(test.Output) > maxOutput {
					t.Fatal("unbounded output")
				}
				if tt.mode == "stderr" && !strings.Contains(test.Error, "diagnostic") {
					t.Fatalf("stderr: %q", test.Error)
				}
			}
			if !slices.Equal(app.Tests, original) {
				t.Fatal("run mutated input tests")
			}
		})
	}
}

func TestCompileAndSubmission(t *testing.T) {
	for _, tt := range []struct {
		mode, file, want string
		fail             bool
	}{
		{"compile", "submission.txt", "built:source", false},
		{"empty-artifact", "submission.txt", "", false},
		{"compile", "missing.txt", "", true},
		{"compile-fail", "submission.txt", "", true},
		{"directory-artifact", "submission.txt", "", true},
		{"large-artifact", "submission.txt", "", true},
	} {
		t.Run(tt.mode+tt.file, func(t *testing.T) {
			app := runner(t, "echo")
			app.Compile, app.Compiler.Args = helper(t, tt.mode)
			app.CompiledFile = tt.file
			got, err := app.Submission(context.Background())
			if (err != nil) != tt.fail || (!tt.fail && got != tt.want) {
				t.Fatalf("got %q, %v", got, err)
			}
		})
	}
	app := runner(t, "echo")
	app.Compile, app.Compiler.Args = helper(t, "compile-fail")
	tests, err := app.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "compiler diagnostic") || GetFinalStatus(tests) != CompilationError {
		t.Fatalf("got %+v, %v", tests, err)
	}
}

func TestCancellationAndCleanup(t *testing.T) {
	for _, stage := range []string{"compile", "run"} {
		t.Run(stage, func(t *testing.T) {
			app := runner(t, "sleep")
			if stage == "compile" {
				app.Compile, app.Compiler.Args = helper(t, "sleep")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			start := time.Now()
			_, err := app.Run(ctx)
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("got %v", err)
			}
			if time.Since(start) > 3*time.Second {
				t.Fatal("cancellation was not prompt")
			}
		})
	}
	for _, mode := range []string{"fail", "success"} {
		t.Run(mode, func(t *testing.T) {
			app := runner(t, "echo")
			record := filepath.Join(app.Root, "workspace")
			app.Compile, app.Compiler.Args = helper(t, "record")
			app.Compiler.Args = append(app.Compiler.Args, record, mode)
			_, _ = app.Run(context.Background())
			dir, err := os.ReadFile(record)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(string(dir)); !os.IsNotExist(err) {
				t.Fatalf("workspace leaked %s: %v", dir, err)
			}
		})
	}
}

func TestMissingCommandsAndSource(t *testing.T) {
	app := runner(t, "echo")
	app.Config.Run = "eitri-nonexistent-runner"
	tests, err := app.Run(context.Background())
	if err != nil || GetFinalStatus(tests) != RuntimeError {
		t.Fatalf("%v, %v", tests, err)
	}
	app.Compile = "eitri-nonexistent-compiler"
	tests, err = app.Run(context.Background())
	if err == nil || GetFinalStatus(tests) != CompilationError {
		t.Fatalf("%v, %v", tests, err)
	}
	app = fixtureApp(t)
	if _, err := app.Solution(); err == nil {
		t.Fatal("missing source accepted")
	}
}

func TestLanguagePipelines(t *testing.T) {
	for _, compiled := range []bool{false, true} {
		t.Run(strconv.FormatBool(compiled), func(t *testing.T) {
			mode := "run-source"
			if compiled {
				mode = "run-artifact"
			}
			app := runner(t, mode)
			writeSource(t, app, "42")
			if compiled {
				app.Compile, app.Args = helper(t, "compile")
				app.CompiledFile = "submission.txt"
			}
			app.Tests = []Testcase{{Answer: "42"}}
			tests, err := app.Run(context.Background())
			if err != nil || GetFinalStatus(tests) != Accepted {
				t.Fatalf("%+v, %v", tests, err)
			}
			artifact, err := app.Submission(context.Background())
			want := "42"
			if compiled {
				want = "built:42"
			}
			if err != nil || artifact != want {
				t.Fatalf("%q, %v", artifact, err)
			}
		})
	}
}
