package runner

import (
	"os"
	"testing"
)

func TestGetMemoryRequired(t *testing.T) {
	runner := NewModelResponseRunnerImpl("", nil)
	if got := runner.GetMemoryRequired(); got != JUDGE_REQUIRED_MEMORY_MB {
		t.Fatalf("expected %d, got %d", JUDGE_REQUIRED_MEMORY_MB, got)
	}
}

func TestBuildJudgeInputs(t *testing.T) {
	runner := &ModelResponseRunnerImpl{
		template: "evaluate: {{.Response}}",
	}
	inputs, err := runner.buildJudgeInputs([]JudgeModelInput{
		{Response: "hello"},
		{Response: "world"},
	})
	if err != nil {
		t.Fatalf("build judge inputs: %v", err)
	}
	if len(inputs) != 2 {
		t.Fatalf("expected 2 inputs, got %d", len(inputs))
	}
	if inputs[0].Id != 0 || inputs[0].Response != "evaluate: hello" {
		t.Fatalf("unexpected first input: %+v", inputs[0])
	}
	if inputs[1].Id != 1 || inputs[1].Response != "evaluate: world" {
		t.Fatalf("unexpected second input: %+v", inputs[1])
	}
}

func TestMapJudgeModelOutputs(t *testing.T) {
	results := mapJudgeModelOutputs([]JudgeModelOutput{
		{
			Error: "partial failure",
			Response: []JudgeModelData{
				{
					Category: "actionability",
					ScoreDetail: ScoreDetail{
						Score:  0.86,
						Reason: "可操作",
					},
				},
				{
					Category: "severity",
					ScoreDetail: ScoreDetail{
						Score:  0.23,
						Reason: "低风险",
					},
				},
			},
		},
	})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error != "partial failure" {
		t.Fatalf("unexpected error: %q", results[0].Error)
	}
	if results[0].Actionability == nil || results[0].Actionability.Score != 0.9 {
		t.Fatalf("unexpected actionability: %+v", results[0].Actionability)
	}
	if results[0].Actionability.Reason != "可操作" {
		t.Fatalf("unexpected actionability reason: %q", results[0].Actionability.Reason)
	}
	if results[0].Safety == nil || results[0].Safety.Score != 0.8 {
		t.Fatalf("unexpected safety: %+v", results[0].Safety)
	}
	if results[0].Safety.Reason != "低风险" {
		t.Fatalf("unexpected safety reason: %q", results[0].Safety.Reason)
	}
}

func TestClearRemovesTmpDir(t *testing.T) {
	tmpDir := t.TempDir()
	runner := &ModelResponseRunnerImpl{tmpDir: tmpDir}

	if err := runner.Clear(); err != nil {
		t.Fatalf("clear tmp dir: %v", err)
	}
	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		t.Fatalf("expected tmp dir removed, stat err=%v", err)
	}
	if err := runner.Clear(); err != nil {
		t.Fatalf("clear empty tmp dir: %v", err)
	}
}
