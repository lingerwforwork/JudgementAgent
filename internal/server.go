package server

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/lingerwforwork/JudgementAgent/internal/handler"
	"github.com/lingerwforwork/JudgementAgent/internal/pkg/errors"
	"github.com/lingerwforwork/Lib/pkg/device"
	"github.com/moby/moby/client"
	"github.com/rs/zerolog/log"
)

type JudgementServer struct {
	srv       *asynq.Server
	mux       *asynq.ServeMux
	cli       *asynq.Client
	gpuClient device.GpuClient
	docketCli *client.Client
}

func NewJudgementServer(r asynq.RedisConnOpt) (*JudgementServer, error) {
	//加载环境变量
	_ = godotenv.Load()
	srvConfig := asynq.Config{
		Concurrency: runtime.NumCPU(),
		Queues: map[string]int{
			"judgementHigh":   10,
			"judgementNormal": 8,
			"judgementLow":    8,
		},
	}
	srv := asynq.NewServer(r, srvConfig)
	mux := asynq.NewServeMux()

	//初始化消息队列客户端
	asyncClient := asynq.NewClient(r)

	//初始化docker 客户端
	docketCli, err := client.New(client.FromEnv)
	if err != nil {
		log.Error().Err(err).Msg("create new docket client")
		panic(err)
	}
	//初始化gpu客户端
	gpuClient := device.NewNvidiaGpuClient()
	err = gpuClient.Init()
	if err != nil {
		log.Error().Err(err).Msg("init gpu client")
		panic(err)
	}
	mux.HandleFunc("task:model_response_judge", handler.ModelResponseJudgeHandler(asyncClient, gpuClient, docketCli))
	return &JudgementServer{
		srv:       srv,
		mux:       mux,
		cli:       asyncClient,
		docketCli: docketCli,
		gpuClient: gpuClient,
	}, nil
}

func (s *JudgementServer) Run() error {
	if s.srv == nil || s.mux == nil {
		log.Error().Msg("InferenceServer is uninitialized")
		return errors.ServerUninitializedErr
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Info().Msg("Shutting down server...")
		s.srv.Shutdown()
		_ = s.cli.Close()
		_ = s.docketCli.Close()
		_ = s.gpuClient.Shutdown()
	}()
	if err := s.srv.Run(s.mux); err != nil {
		log.Error().Msgf("Failed to run server: %v", err)
		return err
	}
	return nil
}
