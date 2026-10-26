package tui

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"github.com/veryshyjelly/eitri/config"
	"github.com/veryshyjelly/eitri/core"
	"github.com/veryshyjelly/eitri/server"
	"github.com/veryshyjelly/eitri/watcher"
)

type Model struct {
	core.App
	fileChanged                 bool
	watch                       *watcher.Watcher
	submissions                 *server.Submissions
	tasks                       *tasks
	problemCtx                  context.Context
	problemCancel               context.CancelFunc
	runCancel                   context.CancelFunc
	copyCancel                  context.CancelFunc
	copyCtx                     context.Context
	writeClipboard              func(context.Context, string) error
	problemID, generation       uint64
	errorText                   string
	index, height, width        int
	leftPane, rightPane         Rect
	leftViewPort, rightViewPort viewport.Model
	mode                        Mode
	orientation                 Orientation
}

type runResult struct {
	generation uint64
	tests      []core.Testcase
	err        error
}
type prepared struct {
	problemID uint64
	path      string
	editor    *exec.Cmd
	err       error
}
type copyResult struct {
	generation uint64
	submission server.Submission
	err        error
}
type clipboardResult struct {
	generation uint64
	err        error
}
type operationError struct {
	problemID uint64
	err       error
}

type Mode int

const (
	InputAnswer Mode = iota
	InputOutput
	InputError
	AnswerOutput
)

type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

// NewModel creates the UI owner of problem state. Commands receive value
// snapshots; only Update installs results or publishes submission artifacts.
func NewModel(parent context.Context, root string, cfg config.Config, orientation Orientation, watch *watcher.Watcher, submissions *server.Submissions) Model {
	tasks := newTasks(parent)
	ctx, cancel := context.WithCancel(tasks.ctx)
	return Model{App: core.App{Root: root, Config: cfg, Status: core.NotAvailable}, watch: watch, submissions: submissions,
		tasks: tasks, problemCtx: ctx, problemCancel: cancel, orientation: orientation,
		leftViewPort: viewport.New(), rightViewPort: viewport.New(), writeClipboard: writeClipboard}
}

// Close cancels and joins owned background work after the tea program exits.
func (m Model) Close()        { m.tasks.close() }
func (m Model) Init() tea.Cmd { return nil }

func (m Model) prepare(create bool) tea.Cmd {
	app, id, ctx := m.App, m.problemID, m.problemCtx
	return func() tea.Msg {
		return m.tasks.do(ctx, func(ctx context.Context) tea.Msg {
			result := prepared{problemID: id}
			if create {
				result.err = app.CreateFile()
			}
			if result.err == nil {
				result.path, result.err = app.FileName()
			}
			if result.err == nil && app.Editor != "" {
				result.editor, result.err = app.EditorCommand(ctx)
			}
			return result
		})
	}
}

func (m Model) startRun() (Model, tea.Cmd) {
	if m.Url == "" {
		return m, nil
	}
	if m.runCancel != nil {
		m.runCancel()
	}
	m.generation++
	m.Status, m.errorText = core.Running, ""
	ctx, cancel := context.WithCancel(m.problemCtx)
	m.runCancel = cancel
	app, id := m.App, m.generation
	app.Tests = slices.Clone(app.Tests)
	return m, func() tea.Msg {
		return m.tasks.do(ctx, func(ctx context.Context) tea.Msg {
			tests, err := app.Run(ctx)
			return runResult{generation: id, tests: tests, err: err}
		})
	}
}

func (m Model) copySolution() (Model, tea.Cmd) {
	if m.copyCancel != nil {
		m.copyCancel()
	}
	m.copyCtx, m.copyCancel = context.WithCancel(m.problemCtx)
	app, id, ctx := m.App, m.generation, m.copyCtx
	return m, func() tea.Msg {
		return m.tasks.do(ctx, func(ctx context.Context) tea.Msg {
			source, err := app.Submission(ctx)
			name := ""
			if err == nil {
				name, err = app.FileName()
			}
			return copyResult{generation: id, err: err, submission: server.Submission{
				ProblemName: app.Title, Url: app.Url, SourceCode: source, FileName: name,
			}}
		})
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	previousError := m.errorText
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Quit):
			m.problemCancel()
			return m, tea.Quit
		case key.Matches(msg, DefaultKeyMap.Run):
			return m.startRun()
		case key.Matches(msg, DefaultKeyMap.CreateFile):
			return m, m.prepare(true)
		case key.Matches(msg, DefaultKeyMap.CopyFile):
			return m.copySolution()
		case key.Matches(msg, DefaultKeyMap.AddCase):
			form := NewAddTestCase(m)
			return form, form.Init()
		case key.Matches(msg, DefaultKeyMap.InputAnswer):
			m.mode = InputAnswer
		case key.Matches(msg, DefaultKeyMap.InputOutput):
			m.mode = InputOutput
		case key.Matches(msg, DefaultKeyMap.InputError):
			m.mode = InputError
		case key.Matches(msg, DefaultKeyMap.AnswerOutput):
			m.mode = AnswerOutput
		case key.Matches(msg, DefaultKeyMap.NextCase) && len(m.Tests) > 0:
			m.index = (m.index + 1) % len(m.Tests)
		case key.Matches(msg, DefaultKeyMap.PreviousCase) && len(m.Tests) > 0:
			m.index = (m.index - 1 + len(m.Tests)) % len(m.Tests)
		case key.Matches(msg, DefaultKeyMap.Help):
			return Help{parent: m, height: m.height, width: m.width}, nil
		case key.Matches(msg, DefaultKeyMap.HorizontalLayout):
			m.orientation = Horizontal
			m = m.setLayout()
		case key.Matches(msg, DefaultKeyMap.VerticalLayout):
			m.orientation = Vertical
			m = m.setLayout()
		}
	case tea.MouseMsg:
		if m.rightPane.Contains(msg.Mouse().X, msg.Mouse().Y) {
			m.rightViewPort, cmd = m.rightViewPort.Update(msg)
			return m, cmd
		} else if m.leftPane.Contains(msg.Mouse().X, msg.Mouse().Y) {
			m.leftViewPort, cmd = m.leftViewPort.Update(msg)
			return m, cmd
		} else if len(m.Tests) > 0 {
			switch msg.Mouse().Button {
			case tea.MouseWheelUp:
				m.index = (m.index + 1) % len(m.Tests)
			case tea.MouseWheelDown:
				m.index = (m.index - 1 + len(m.Tests)) % len(m.Tests)
			}
		}
	case tea.WindowSizeMsg:
		m.width, m.height = max(1, msg.Width-2), max(1, msg.Height-2)
		m = m.setLayout()
	case core.Info:
		m.problemCancel()
		m.problemCtx, m.problemCancel = context.WithCancel(m.tasks.ctx)
		m.problemID++
		m.generation++
		m.errorText = ""
		m.fileChanged = false
		m.submissions.Clear()
		m.watch.Select("")
		m = m.setProblem(msg).updatePanes()
		return m, m.prepare(m.Config.CreateFile)
	case prepared:
		if msg.problemID != m.problemID {
			return m, nil
		}
		if msg.err != nil {
			m.errorText = msg.err.Error()
			break
		}
		m.watch.Select(msg.path)
		if msg.editor != nil {
			id := m.problemID
			return m, tea.ExecProcess(msg.editor, func(err error) tea.Msg { return operationError{problemID: id, err: err} })
		}
	case runResult:
		if msg.generation != m.generation {
			return m, nil
		}
		m.Tests = msg.tests
		m.Status = core.GetFinalStatus(msg.tests)
		if msg.err != nil {
			m.errorText = msg.err.Error()
		}
	case copyResult:
		if msg.generation != m.generation {
			return m, nil
		}
		if msg.err != nil {
			m.errorText = msg.err.Error()
			break
		}
		m.submissions.Publish(msg.submission)
		m.fileChanged = false
		source, id, ctx := msg.submission.SourceCode, m.generation, m.copyCtx
		if ctx == nil {
			ctx = m.problemCtx
		}
		return m, func() tea.Msg {
			return m.tasks.do(ctx, func(context.Context) tea.Msg {
				return clipboardResult{generation: id, err: m.writeClipboard(ctx, source)}
			})
		}
	case clipboardResult:
		if msg.generation == m.generation && msg.err != nil {
			m.errorText = fmt.Sprintf("clipboard: %v", msg.err)
		}
	case operationError:
		if msg.problemID == m.problemID && msg.err != nil {
			m.errorText = msg.err.Error()
		}
	case watcher.Selected:
		return m, nil
	case watcher.Error:
		m.errorText = msg.Err.Error()
	case watcher.Changed:
		name, err := m.FileName()
		if err != nil {
			m.errorText = err.Error()
			break
		}
		path, _ := filepath.Abs(config.ResolvePath(m.Root, name))
		if path != msg.Path {
			return m, nil
		}
		m.fileChanged = true
		if m.copyCancel != nil {
			m.copyCancel()
		}
		m.submissions.Clear()
		if m.runCancel != nil {
			m.runCancel()
		}
		m.generation++ // invalidates results even when run_on_save is disabled
		m.Status = core.NotAvailable
		if m.Config.RunOnSave {
			return m.startRun()
		}
	}
	if m.errorText != "" && m.errorText != previousError {
		log.Error("operation failed", "err", m.errorText)
	}
	return m.updatePanes(), cmd
}

// updatePanes synchronizes the content of the left and right UI viewports with
// the currently selected test case and the active display mode (e.g., Input/Output,
// Input/Answer, Input/Error).
func (m Model) updatePanes() Model {
	if len(m.Tests) == 0 {
		m.index = 0
		m.leftViewPort.SetContent("")
		m.rightViewPort.SetContent("")
		return m
	}
	m.index = max(0, min(m.index, len(m.Tests)-1))

	width := max(1, m.leftPane.W)

	testCase := m.Tests[m.index]
	switch m.mode {
	case InputAnswer:
		m.leftViewPort.SetContent(wrapContent(testCase.Input, width))
		m.rightViewPort.SetContent(wrapContent(testCase.Answer, width))
	case InputOutput:
		m.leftViewPort.SetContent(wrapContent(testCase.Input, width))
		m.rightViewPort.SetContent(wrapContent(testCase.Output, width))
	case InputError:
		m.leftViewPort.SetContent(wrapContent(testCase.Input, width))
		m.rightViewPort.SetContent(wrapContent(testCase.Error, width))
	case AnswerOutput:
		m.leftViewPort.SetContent(wrapContent(testCase.Answer, width))
		m.rightViewPort.SetContent(wrapContent(testCase.Output, width))
	}

	return m
}

// setLayout recalculates and applies the dimensions, padding, and coordinates
// for the split-pane UI layout based on the current terminal window size.
func (m Model) setLayout() Model {
	if m.orientation == Vertical {
		m.leftPane = Rect{
			X: 0,
			Y: 5,
			W: max(1, m.width-4),
			H: max(2, (m.height-5)/2),
		}
		m.leftViewPort.SetHeight(m.leftPane.H - 1)
		m.leftViewPort.SetWidth(m.leftPane.W)

		m.rightPane = Rect{
			X: 0,
			Y: 5 + (m.height-5)/2 + 1,
			W: max(1, m.width-4),
			H: max(2, (m.height-5)/2),
		}
		m.rightViewPort.SetHeight(m.rightPane.H - 1)
		m.rightViewPort.SetWidth(m.rightPane.W)
	} else {
		m.leftPane = Rect{
			X: 0,
			Y: 5,
			W: max(1, m.width/2-4),
			H: max(2, m.height-5),
		}
		m.leftViewPort.SetHeight(m.leftPane.H - 1)
		m.leftViewPort.SetWidth(m.leftPane.W)

		m.rightPane = Rect{
			X: (m.width + 1) / 2,
			Y: 5,
			W: max(1, m.width/2-4),
			H: max(2, m.height-5),
		}
		m.rightViewPort.SetHeight(m.rightPane.H - 1)
		m.rightViewPort.SetWidth(m.rightPane.W)
	}
	return m
}

// setProblem updates the model's internal state with a new competitive programming problem.
// It initializes the test cases, resets the current test index to prevent out-of-bounds errors,
// and resets the overall execution status.
func (m Model) setProblem(info core.Info) Model {
	m.Status = core.NotAvailable
	m.index = 0
	// Fill problem and test case in model
	m.Problem = core.Problem{
		Title:       info.Name,
		Url:         info.Url,
		MemoryLimit: info.MemoryLimit,
		TimeLimit:   info.TimeLimit,
	}
	m.Tests = make([]core.Testcase, 0, len(info.Tests))
	for _, t := range info.Tests {
		m.Tests = append(m.Tests, core.Testcase{
			Input:  t.Input,
			Answer: t.Output,
			Status: core.NotAvailable,
		})
	}
	return m
}
