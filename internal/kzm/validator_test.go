package kzm

import (
	"testing"
)

func TestValidator(t *testing.T) {
	// 1. Unknown target with suggestion
	prog := &Program{
		Line: 1,
		Triggers: []*Trigger{
			{Source: "github", Repo: "foo/bar"},
		},
		Deploy: &DeployTarget{
			Type: "lambbd", // Typo
			Name: "test-lambda",
			Line: 10,
		},
	}

	res := Validate(prog)
	if len(res.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(res.Errors))
	}
	err := res.Errors[0]
	if err.Line != 10 || !stringsContains(err.Message, `Unknown deploy target type "lambbd". Did you mean "lambda"?`) {
		t.Errorf("unexpected error message: %q", err.Message)
	}

	// 2. Lambda missing region
	prog2 := &Program{
		Line: 1,
		Triggers: []*Trigger{
			{Source: "github", Repo: "foo/bar"},
		},
		Deploy: &DeployTarget{
			Type: "lambda",
			Name: "test-lambda",
			Line: 5,
		},
	}
	res2 := Validate(prog2)
	if len(res2.Errors) != 2 { // Missing region, missing runtime
		t.Fatalf("expected 2 errors, got %d", len(res2.Errors))
	}

	// 3. Lambda valid region
	prog3 := &Program{
		Line: 1,
		Triggers: []*Trigger{
			{Source: "github", Repo: "foo/bar"},
		},
		Deploy: &DeployTarget{
			Type: "lambda",
			Name: "test-lambda",
			Options: map[string]string{
				"region":  "us-east-1",
				"runtime": "nodejs20.x",
			},
			Line: 5,
		},
	}
	res3 := Validate(prog3)
	if len(res3.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d: %+v", len(res3.Errors), res3.Errors)
	}

	// 4. Invalid region format
	prog4 := &Program{
		Line: 1,
		Triggers: []*Trigger{
			{Source: "github", Repo: "foo/bar"},
		},
		Deploy: &DeployTarget{
			Type: "lambda",
			Name: "test-lambda",
			Options: map[string]string{
				"region":  "invalid-region-format-123",
				"runtime": "nodejs20.x",
			},
			Line: 5,
		},
	}
	res4 := Validate(prog4)
	if len(res4.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(res4.Errors))
	}
}

func stringsContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || stringsContainsWord(s, sub))
}

func stringsContainsWord(s, sub string) bool {
	// basic substring checker
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
