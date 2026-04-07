package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/techinpark/ga-cli/internal/auth"
	"github.com/techinpark/ga-cli/internal/client"
	"github.com/techinpark/ga-cli/internal/config"
	"github.com/techinpark/ga-cli/internal/formatter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

var (
	cfgFile     string
	credentials string
	format      string
	account     string
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
	cmd.PersistentFlags().StringVar(&account, "account", "", "사용할 계정 (기본: 활성 계정)")

	return cmd
}

// Execute is the main entry point for the CLI.
func Execute(version string) error {
	cobra.OnInitialize(func() {
		initConfig()
	})

	rootCmd := newRootCmd(version)

	var deps *Dependencies
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// auth 서브커맨드는 API 클라이언트 불필요
		if isAuthCommand(cmd) {
			return nil
		}

		cfg, err := config.Load()
		if err != nil {
			cfg = &config.Config{}
		}

		f, _ := cmd.Flags().GetString("format")
		ft, err := formatter.ParseFormat(f)
		if err != nil {
			return err
		}

		// 인증 옵션 결정 (우선순위: Service Account > OAuth2 > ADC)
		var apiOpts []option.ClientOption

		credPath := resolveCredentials(cmd, cfg)
		if credPath != "" {
			apiOpts = append(apiOpts, option.WithCredentialsFile(credPath))
		} else {
			authMgr, authErr := auth.NewManager()
			if authErr == nil {
				_ = authMgr.MigrateTokenIfNeeded()

				accountName := account // --account 글로벌 플래그
				if accountName == "" {
					accountName = authMgr.ActiveAccount()
				}

				if authMgr.IsLoggedInAs(accountName) {
					tokenSource, tsErr := authMgr.GetTokenSourceForAccount(cmd.Context(), accountName)
					if tsErr == nil {
						apiOpts = append(apiOpts, option.WithTokenSource(tokenSource))
					}
				}
			}
			// apiOpts가 비어있으면 ADC fallback
		}

		adminClient, err := client.NewAdminClient(apiOpts, cfg.Aliases)
		if err != nil {
			return fmt.Errorf("failed to create admin client: %w", err)
		}

		dataClient, err := client.NewDataClient(apiOpts)
		if err != nil {
			return fmt.Errorf("failed to create data client: %w", err)
		}

		deps = &Dependencies{
			Admin:     adminClient,
			Data:      dataClient,
			Resolver:  client.NewPropertyResolver(cfg.Aliases),
			Formatter: formatter.New(ft),
			Config:    cfg,
		}
		return nil
	}

	getDeps := func() *Dependencies {
		return deps
	}

	rootCmd.AddCommand(newAuthCmd())
	rootCmd.AddCommand(newPropertiesCmd(getDeps))
	rootCmd.AddCommand(newDAUCmd(getDeps))
	rootCmd.AddCommand(newEventsCmd(getDeps))
	rootCmd.AddCommand(newCountriesCmd(getDeps))
	rootCmd.AddCommand(newPlatformsCmd(getDeps))
	rootCmd.AddCommand(newRealtimeCmd(getDeps))

	return rootCmd.Execute()
}

// resolveCredentials determines the Service Account credentials file path
// from flag, config, or env. Returns empty string if none is configured,
// allowing OAuth2 or ADC fallback.
func resolveCredentials(cmd *cobra.Command, cfg *config.Config) string {
	if f := cmd.Flag("credentials"); f != nil && f.Changed {
		return f.Value.String()
	}

	if cfg.Credentials != "" {
		return cfg.Credentials
	}

	if envPath := os.Getenv("GA_SERVICE_ACCOUNT_KEY"); envPath != "" {
		return envPath
	}

	return ""
}

// isAuthCommand checks if the command is part of the auth subcommand tree.
func isAuthCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "auth" {
			return true
		}
	}
	return false
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
