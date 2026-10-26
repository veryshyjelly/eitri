package core

import "testing"

// These cases characterize the original verdict priority and marker-free source
// behavior before changing execution and linking.
func TestVerdictContract(t *testing.T) {
	for _, tt := range []struct {
		name     string
		statuses []Status
		want     Status
	}{
		{"accepted", []Status{Accepted, Accepted}, Accepted},
		{"wrong answer", []Status{Accepted, WrongAnswer}, WrongAnswer},
		{"timeout before wrong answer", []Status{WrongAnswer, TimeLimitExceeded}, TimeLimitExceeded},
		{"runtime before timeout", []Status{TimeLimitExceeded, RuntimeError}, RuntimeError},
		{"compile before runtime", []Status{RuntimeError, CompilationError}, CompilationError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tests := make([]Testcase, len(tt.statuses))
			for i, status := range tt.statuses {
				tests[i].Status = status
			}
			if got := GetFinalStatus(tests); got != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestUnmarkedSourceContract(t *testing.T) {
	source := "\n  print(42)\n"
	if got, err := extractBlock(source, "code", source); err != nil || got != source {
		t.Fatalf("source changed: %q", got)
	}
	if got, err := extractBlock(source, "head", ""); err != nil || got != "" {
		t.Fatalf("unexpected header: %q", got)
	}
}
