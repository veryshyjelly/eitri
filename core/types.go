package core

import (
	"errors"

	"github.com/veryshyjelly/eitri/config"
)

var ErrNoProblem = errors.New("no problem is loaded")

type App struct {
	Root string
	config.Config
	Problem
	Status
	Tests []Testcase
}

type Problem struct {
	Title       string
	Url         string
	MemoryLimit uint64
	TimeLimit   uint64
}

type Testcase struct {
	Input  string
	Output string
	Error  string
	Answer string
	Status Status
	Time   float64
	Memory uint64 // in kbs
}

type Info struct {
	Name        string `json:"name"`
	Group       string `json:"group"`
	Url         string `json:"url"`
	Interactive bool   `json:"interactive"`
	MemoryLimit uint64 `json:"memoryLimit"`
	TimeLimit   uint64 `json:"timeLimit"`
	Tests       []Test `json:"tests"`
}

type Test struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type Status string

const (
	NotAvailable      Status = "NA"
	Accepted          Status = "AC"
	RuntimeError      Status = "RE"
	CompilationError  Status = "CE"
	TimeLimitExceeded Status = "TLE"
	WrongAnswer       Status = "WA"
	Running           Status = "WIP"
)
