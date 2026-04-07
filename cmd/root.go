package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	credentials string
	format      string
)

func newRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ga",
		Short:   "Google Analytics CLI 도구",
		Long:    "Google Analytics 4 데이터를 터미널에서 빠르게 조회하는 CLI 도구입니다.",
		Version: version,
	}

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "설정 파일 경로 (기본값: ~/.ga-cli/config.yaml)")
	cmd.PersistentFlags().StringVarP(&credentials, "credentials", "c", "", "서비스 계정 키 경로")
	cmd.PersistentFlags().StringVarP(&format, "format", "f", "table", "출력 형식 (table, json, csv)")

	return cmd
}

func Execute(version string) error {
	rootCmd := newRootCmd(version)

	cobra.OnInitialize(func() {
		initConfig()
	})

	return rootCmd.Execute()
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to get home directory: %v\n", err)
			return
		}

		configDir := filepath.Join(home, ".ga-cli")
		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("GA_CLI")
	viper.AutomaticEnv()

	viper.SetDefault("defaults.days", 30)
	viper.SetDefault("defaults.top", 20)
	viper.SetDefault("defaults.output", "table")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "warning: failed to read config: %v\n", err)
		}
	}
}
