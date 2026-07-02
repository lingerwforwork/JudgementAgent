package handler

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/lingerwforwork/Lib/pkg/device"
	"github.com/moby/moby/client"
)

func ModelResponseJudgeHandler(cli *asynq.Client, gpuClient device.GpuClient, docketCli *client.Client) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		info := string(task.Payload())
		fmt.Println(info)
		return nil
	}
}
