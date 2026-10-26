package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/shlex"
	"github.com/veryshyjelly/eitri/config"
	"github.com/veryshyjelly/eitri/memory"
)

const (
	compileTimeout   = 2 * time.Minute
	maxOutput        = 4 << 20 // per stream, per process
	defaultTimeLimit = 2 * time.Second
)

// limitedOutput continues draining until cancellation has stopped the process.
// Each buffer has exactly one writer (the corresponding os/exec copy goroutine).
type limitedOutput struct {
	data     []byte
	exceeded bool
	cancel   context.CancelFunc
}

func (b *limitedOutput) Write(p []byte) (int, error) {
	n := len(p)
	remaining := maxOutput - len(b.data)
	if len(p) > remaining {
		p = p[:remaining]
		b.exceeded = true
		b.cancel()
	}
	b.data = append(b.data, p...)
	return n, nil
}

type processResult struct {
	stdout, stderr string
	duration       float64
	memory         uint64
	err            error
}

func execute(ctx context.Context, dir, command string, args []string, input string) processResult {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stdout, stderr := &limitedOutput{cancel: cancel}, &limitedOutput{cancel: cancel}
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout, cmd.Stderr = stdout, stderr
	// Bound pipe draining even if a child retains stdout after its parent exits.
	cmd.WaitDelay = 250 * time.Millisecond
	start := time.Now()
	err := runProcess(ctx, cmd)
	if stdout.exceeded || stderr.exceeded {
		err = errors.New("output limit exceeded (4 MiB per stream)")
	}
	peak, _ := memory.PeakMemory(cmd)
	return processResult{string(stdout.data), string(stderr.data), time.Since(start).Seconds(), peak, err}
}

// compile transfers ownership of its temporary workspace only on success.
// A temporary working directory provides file isolation, not a security sandbox.
func (app App) compile(ctx context.Context) (dir string, err error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	solution, err := app.Solution()
	if err != nil {
		return "", err
	}
	if !config.LocalFile(app.Compiler.Source) {
		return "", fmt.Errorf("compiler.source must be a relative workspace file")
	}
	dir, err = os.MkdirTemp("", "eitri-")
	if err != nil {
		return "", fmt.Errorf("create workspace: %w", err)
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.RemoveAll(dir))
			dir = ""
		}
	}()
	path := filepath.Join(dir, app.Compiler.Source)
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return dir, fmt.Errorf("prepare source directory: %w", err)
	}
	if err = os.WriteFile(path, []byte(solution), 0644); err != nil {
		return dir, fmt.Errorf("prepare source: %w", err)
	}
	if app.Compile == "" {
		return dir, ctx.Err()
	}
	buildCtx, cancel := context.WithTimeout(ctx, compileTimeout)
	defer cancel()
	result := execute(buildCtx, dir, app.Compile, app.Compiler.Args, "")
	if buildCtx.Err() != nil {
		return dir, fmt.Errorf("compile: %w", buildCtx.Err())
	}
	if result.err != nil {
		return dir, fmt.Errorf("compile: %w\n%s%s", result.err, result.stdout, result.stderr)
	}
	return dir, nil
}

// Run judges samples sequentially in their original order. It never mutates the
// caller's test slice. Cancellation aborts the run rather than inventing a verdict.
func (app App) Run(ctx context.Context) (tests []Testcase, err error) {
	tests = slices.Clone(app.Tests)
	for i := range tests {
		tests[i].Status, tests[i].Output, tests[i].Error = NotAvailable, "", ""
		tests[i].Time, tests[i].Memory = 0, 0
	}
	dir, err := app.compile(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		for i := range tests {
			tests[i].Status, tests[i].Error = CompilationError, err.Error()
		}
		return tests, err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(dir)) }()
	args, err := shlex.Split(app.Config.Run)
	if err != nil || len(args) == 0 || args[0] == "" {
		return tests, fmt.Errorf("invalid compiler.run command %q", app.Config.Run)
	}
	timeout := defaultTimeLimit
	if app.TimeLimit > 0 {
		// Avoid uint64 -> Duration overflow from malformed intake/configuration.
		if app.TimeLimit > uint64((24*time.Hour)/time.Millisecond) {
			return tests, fmt.Errorf("time limit exceeds 24 hours")
		}
		timeout = time.Duration(app.TimeLimit) * time.Millisecond
	}
	for i := range tests {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		sampleCtx, cancel := context.WithTimeout(ctx, timeout)
		result := execute(sampleCtx, dir, args[0], args[1:], tests[i].Input)
		timedOut := errors.Is(sampleCtx.Err(), context.DeadlineExceeded)
		cancel()
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		tests[i].Output, tests[i].Error = result.stdout, result.stderr
		tests[i].Time, tests[i].Memory = result.duration, result.memory
		tests[i].Status = classify(result.stdout, tests[i].Answer, result.err, timedOut)
		if result.err != nil {
			tests[i].Error = strings.TrimSpace(result.stderr + "\n" + result.err.Error())
		}
	}
	return tests, nil
}

func classify(output, answer string, err error, timedOut bool) Status {
	switch {
	case timedOut:
		return TimeLimitExceeded
	case err != nil:
		return RuntimeError
	case strings.TrimSpace(output) == strings.TrimSpace(answer):
		return Accepted
	default:
		return WrongAnswer
	}
}

// Submission selects exactly the configured artifact, including an empty one.
// Artifact errors never fall back to submitting a different file.
func (app App) Submission(ctx context.Context) (content string, err error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if app.CompiledFile == "" {
		return app.Solution()
	}
	if !config.LocalFile(app.CompiledFile) {
		return "", fmt.Errorf("compiled_file must be a relative workspace file")
	}
	dir, err := app.compile(ctx)
	if err != nil {
		return "", err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(dir)) }()
	path := filepath.Join(dir, app.CompiledFile)
	// Resolve compiler-created symlinks before reading the requested artifact.
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("submission artifact: %w", err)
	}
	workspace, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", fmt.Errorf("resolve workspace: %w", err)
	}
	rel, err := filepath.Rel(workspace, resolved)
	if err != nil || !filepath.IsLocal(rel) {
		return "", fmt.Errorf("submission artifact escapes workspace")
	}
	stat, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("submission artifact: %w", err)
	}
	if !stat.Mode().IsRegular() {
		return "", fmt.Errorf("submission artifact must be a regular file")
	}
	f, err := os.Open(resolved)
	if err != nil {
		return "", fmt.Errorf("submission artifact: %w", err)
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxOutput+1))
	if err != nil {
		return "", fmt.Errorf("submission artifact: %w", err)
	}
	if len(data) > maxOutput {
		return "", fmt.Errorf("submission artifact exceeds 4 MiB")
	}
	return string(data), nil
}

// GetFinalStatus preserves the existing failure priority. Unrun/empty samples
// have no verdict; they cannot establish acceptance.
func GetFinalStatus(tests []Testcase) Status {
	for _, status := range []Status{CompilationError, RuntimeError, TimeLimitExceeded, WrongAnswer} {
		if slices.ContainsFunc(tests, func(t Testcase) bool { return t.Status == status }) {
			return status
		}
	}
	if len(tests) == 0 || slices.ContainsFunc(tests, func(t Testcase) bool { return t.Status != Accepted }) {
		return NotAvailable
	}
	return Accepted
}
