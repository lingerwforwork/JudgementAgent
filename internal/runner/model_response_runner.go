package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"text/template"

	"github.com/lingerwforwork/Lib/pkg/container"
	libContainer "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/client"
)

type JudgeModelInput struct {
	Response string `json:"response"`
}

type ScoreDetail struct {
	Score  float32 `json:"score"`
	Reason string  `json:"reason"`
}

type Input struct {
	Id       int    `json:"id"`
	Response string `json:"response"`
}

type JudgeModelData struct {
	ScoreDetail
	Category string `json:"category"`
}

type JudgeModelOutput struct {
	Response []JudgeModelData `json:"response"`
	Error    string           `json:"error,omitempty"`
}

type Output struct {
	ScoreSet
	Error     string `json:"error,omitempty"`
	IsCorrect bool   `json:"isCorrect"`
	Judgement *bool  `json:"judgement,omitempty"`
}

type ScoreSet struct {
	Actionability  *ScoreDetail `gorm:"embedded;embeddedPrefix:actionability_" json:"actionability,omitempty"`
	Bias           *ScoreDetail `gorm:"embedded;embeddedPrefix:bias_" json:"bias,omitempty"`
	Coherence      *ScoreDetail `gorm:"embedded;embeddedPrefix:coherence_" json:"coherence,omitempty"`
	Compliance     *ScoreDetail `gorm:"embedded;embeddedPrefix:compliance_" json:"compliance,omitempty"`
	Helpfulness    *ScoreDetail `gorm:"embedded;embeddedPrefix:helpfulness_" json:"helpfulness,omitempty"`
	Meaningfulness *ScoreDetail `gorm:"embedded;embeddedPrefix:meaningfulness_" json:"meaningfulness,omitempty"`
	Politeness     *ScoreDetail `gorm:"embedded;embeddedPrefix:politeness_" json:"politeness,omitempty"`
	Safety         *ScoreDetail `gorm:"embedded;embeddedPrefix:safety_" json:"safety,omitempty"`
}

type ModelResponseRunner interface {
	GetMemoryRequired() int
	Run(gpuIds []uint32, stdout, stderr io.Writer, inputs []JudgeModelInput) ([]Output, error)
	Clear() error
}

type ModelResponseRunnerImpl struct {
	cli         *client.Client
	template    string
	WorkDirTmpl string
	tmpDir      string
}

func NewModelResponseRunnerImpl(template string, cli *client.Client) *ModelResponseRunnerImpl {
	tmpl := template
	if template == "" {
		tmpl = promptTemplate
	}
	return &ModelResponseRunnerImpl{
		cli:         cli,
		template:    tmpl,
		WorkDirTmpl: WORK_DIR_TMPL,
	}
}

func (runner *ModelResponseRunnerImpl) GetMemoryRequired() int {
	return JUDGE_REQUIRED_MEMORY_MB
}

func (runner *ModelResponseRunnerImpl) Run(gpuIds []uint32, stdout, stderr io.Writer, judgeModelInputs []JudgeModelInput) ([]Output, error) {
	tmpl, err := template.New("").Parse(runner.template)
	if err != nil {
		return nil, err
	}
	inputs := make([]Input, len(judgeModelInputs))
	for index, judgeModelInput := range judgeModelInputs {
		var prompt bytes.Buffer
		_ = tmpl.Execute(&prompt, judgeModelInput)
		inputs[index] = Input{
			Id:       index,
			Response: prompt.String(),
		}
	}
	tmpDir, err := os.MkdirTemp(runner.WorkDirTmpl, "judgement-*")
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(tmpDir, os.ModePerm); err != nil {
		return nil, err
	}
	runner.tmpDir = tmpDir
	inputFile := filepath.Join(tmpDir, "input")
	outputFile := filepath.Join(tmpDir, "output")

	out, err := json.Marshal(inputs)
	if err != nil {
		return nil, err
	}

	if err = os.WriteFile(inputFile, out, fs.ModePerm); err != nil {
		return nil, err
	}

	config := &libContainer.Config{
		Image:      JUDGE_DOCKER_IMAGE,
		WorkingDir: "/app",
		Cmd: []string{
			"python3",
			"run_model.py",
			"--model_path=/model",
			"--use_fast=False",
			fmt.Sprintf("--input_path=%s", inputFile),
			fmt.Sprintf("--output_path=%s", outputFile),
			"--bits=4_BIT",
		},
		AttachStdout: true,
		AttachStderr: true,
	}
	hostConfig := &libContainer.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   PYFILE_PATH_TEMPLATE,
				Target:   "/app/run_model.py",
				ReadOnly: true,
			},
			{
				Type:     mount.TypeBind,
				Source:   MODEL_PATH,
				Target:   "/model",
				ReadOnly: true,
			},
			{
				Type:     mount.TypeBind,
				Source:   tmpDir,
				Target:   tmpDir,
				ReadOnly: false,
			},
		},
		SecurityOpt: []string{
			"seccomp=unconfined",
		},
		//自动删除
		AutoRemove: true,
	}

	containerInstance := container.NewDockerContainer(runner.cli, config, hostConfig)
	waitResult, err := containerInstance.Run(gpuIds, stdout, stderr)
	if err != nil {
		return nil, err
	}
	select {
	case err := <-waitResult.Error:
		return nil, err
	case <-waitResult.Result:
		raw, err := os.ReadFile(outputFile)
		if err != nil {
			return nil, err
		}
		var output []JudgeModelOutput
		if err := json.Unmarshal(raw, &output); err != nil {
			return nil, err
		}
		results := make([]Output, len(output))
		for i, o := range output {
			results[i].Error = o.Error
			if o.Response == nil {
				continue
			}
			for _, e := range o.Response {
				detail := e.ScoreDetail
				switch e.Category {
				case "actionability":
					results[i].Actionability = &ScoreDetail{
						Score:  float32(math.Round(float64(detail.Score)*10)) / 10,
						Reason: detail.Reason,
					}
				//case domain.JUDGEMENT_SCORE_CATEGORY_BIAS:
				//	results[i].Bias = &detail
				//case domain.JUDGEMENT_SCORE_CATEGORY_COHERENCE:
				//	results[i].Coherence = &detail
				//case domain.JUDGEMENT_SCORE_CATEGORY_COMPLIANCE:
				//	results[i].Compliance = &detail
				//case domain.JUDGEMENT_SCORE_CATEGORY_HELPFULNESS:
				//	results[i].Helpfulness = &detail
				//case domain.JUDGEMENT_SCORE_CATEGORY_MEANINGFULNESS:
				//	results[i].Meaningfulness = &detail
				//case domain.JUDGEMENT_SCORE_CATEGORY_POLITENESS:
				//	results[i].Politeness = &detail
				case "severity":
					results[i].Safety = &ScoreDetail{
						Score:  float32(math.Round(float64(1-detail.Score)*10)) / 10,
						Reason: detail.Reason,
					}
				}
			}
		}
		return results, nil
	}

}

func (runner *ModelResponseRunnerImpl) Clear() error {
	if runner.tmpDir != "" {
		return os.RemoveAll(runner.tmpDir)
	}
	return nil
}
