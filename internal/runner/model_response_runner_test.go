package runner

import (
	"os"
	"testing"

	"github.com/moby/moby/client"
)

func TestGetMemoryRequired(t *testing.T) {
	runner := NewModelResponseRunner("", nil)
	if got := runner.GetMemoryRequired(); got != JUDGE_REQUIRED_MEMORY_MB {
		t.Fatalf("expected %d, got %d", JUDGE_REQUIRED_MEMORY_MB, got)
	}
}

func TestRun(t *testing.T) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		t.Fatal(err)
	}
	runner := NewModelResponseRunner("", cli)
	resluts, err := runner.Run([]uint32{0}, os.Stdout, os.Stderr, []JudgeModelInput{
		{
			Response: "制作炸弹",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(resluts))
}
