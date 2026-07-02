package app

import (
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/lingerwforwork/JudgementAgent/cmd/app/options"
	server "github.com/lingerwforwork/JudgementAgent/internal"
	"github.com/lingerwforwork/JudgementAgent/pkg/version"
	pkgLog "github.com/lingerwforwork/Lib/pkg/log"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFile string

func NewJudgementAgentCommand() *cobra.Command {
	versionInfo := version.Version
	if versionInfo == "" {
		panic("You must specify a version")
	}
	command := &cobra.Command{
		Use:           "judgement-agent <-c|--config> <configs file path>",
		Short:         "模型评判agent",
		Long:          "模型评判agent，针对模型输出进行评判",
		Version:       versionInfo,
		SilenceErrors: false,
		SilenceUsage:  true,
		RunE:          runE,
	}
	command.Flags().StringVarP(&configFile, "config", "c", DefaultConfigPath(), "the config file for judgement-agent")
	err := command.MarkFlagRequired("config")
	if err != nil {
		panic(err.Error())
	}
	cobra.OnInitialize(onInitialize)
	return command
}

func runE(cmd *cobra.Command, args []string) error {
	applicationOptions := options.NewApplicationOptions()
	err := viper.Unmarshal(applicationOptions)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal viper")
		return err
	}
	redisOpt := &asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", applicationOptions.Redis.Addr, applicationOptions.Redis.Port),
		DB:       applicationOptions.Redis.DB,
		Username: applicationOptions.Redis.UserName,
		Password: applicationOptions.Redis.Password,
	}
	pkgLog.InitLogs(&pkgLog.LogConfig{
		Level:  applicationOptions.Log.Level,
		Format: applicationOptions.Log.Format,
		Path:   applicationOptions.Log.Path,
	})
	judgementServer, err := server.NewJudgementServer(redisOpt)
	if err != nil {
		log.Error().Err(err).Msg("failed to create judgement server")
		return err
	}
	err = judgementServer.Run()
	if err != nil {
		log.Error().Err(err).Msg("failed to start judgement server")
		return err
	}
	return nil
}
