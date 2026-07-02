package main

import (
	"os"

	"github.com/lingerwforwork/JudgementAgent/cmd/app"
)

func main() {
	command := app.NewJudgementAgentCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
