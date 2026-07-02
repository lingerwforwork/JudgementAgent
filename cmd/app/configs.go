package app

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultConfigPath = "judgement-agent"

	defaultConfigFileName = "configs.yaml"

	defaultConfigFileNameWithOutExtension = "configs"
)

// 使用viper读取全局配置
func onInitialize() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
		_ = viper.ReadInConfig()
		return
	}
	viper.SetConfigType("yaml")
	for _, dir := range searchDirs() {
		viper.AddConfigPath(dir)
	}
	viper.SetConfigName(defaultConfigFileNameWithOutExtension)
	_ = viper.ReadInConfig()
}
func searchDirs() []string {
	homeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)
	return []string{filepath.Join(homeDir, defaultConfigPath), "."}
}

func DefaultConfigPath() string {
	dir, err := os.UserHomeDir()
	cobra.CheckErr(err)
	return filepath.Join(dir, defaultConfigPath, defaultConfigFileName)
}
