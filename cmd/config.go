package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func newConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "설정 관리",
	}

	configCmd.AddCommand(newConfigListCmd())
	configCmd.AddCommand(newConfigGetCmd())
	configCmd.AddCommand(newConfigSetCmd())
	configCmd.AddCommand(newConfigAliasCmd())
	configCmd.AddCommand(newConfigPathCmd())

	return configCmd
}

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "전체 설정 출력",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := viper.ConfigFileUsed()
			if configPath == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "No config file found.")
				return nil
			}

			data, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to read config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "# %s\n\n%s", configPath, string(data))
			return nil
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "설정 값 조회 (예: defaults.days, aliases.my-app)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			if !viper.IsSet(key) {
				return fmt.Errorf("key %q not found", key)
			}

			value := viper.Get(key)
			fmt.Fprintf(cmd.OutOrStdout(), "%v\n", value)
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "설정 값 변경 (예: defaults.days 14)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			cfg, err := loadRawConfig()
			if err != nil {
				cfg = make(map[string]interface{})
			}

			setNestedValue(cfg, key, value)

			if err := saveRawConfig(cfg); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Set %s = %s\n", key, value)
			return nil
		},
	}
}

func newConfigAliasCmd() *cobra.Command {
	var remove bool

	cmd := &cobra.Command{
		Use:   "alias <name> [property-id]",
		Short: "속성 별칭 등록/삭제",
		Args: func(cmd *cobra.Command, args []string) error {
			if remove {
				if len(args) < 1 {
					return fmt.Errorf("alias name is required")
				}
				return nil
			}
			if len(args) < 2 {
				return fmt.Errorf("usage: ga config alias <name> <property-id>")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadRawConfig()
			if err != nil {
				cfg = make(map[string]interface{})
			}

			aliases, _ := cfg["aliases"].(map[string]interface{})
			if aliases == nil {
				aliases = make(map[string]interface{})
			}

			name := args[0]

			if remove {
				if _, ok := aliases[name]; !ok {
					return fmt.Errorf("alias %q not found", name)
				}
				delete(aliases, name)
				cfg["aliases"] = aliases
				if err := saveRawConfig(cfg); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Removed alias %q\n", name)
				return nil
			}

			propertyID := args[1]
			aliases[name] = propertyID
			cfg["aliases"] = aliases

			if err := saveRawConfig(cfg); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Set alias %s = %s\n", name, propertyID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&remove, "delete", false, "별칭 삭제")
	return cmd
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "설정 파일 경로 출력",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := viper.ConfigFileUsed()
			if configPath == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				configPath = filepath.Join(home, ".ga-cli", "config.yaml")
			}
			fmt.Fprintln(cmd.OutOrStdout(), configPath)
			return nil
		},
	}
}

func configFilePath() (string, error) {
	configPath := viper.ConfigFileUsed()
	if configPath != "" {
		return configPath, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".ga-cli", "config.yaml"), nil
}

func loadRawConfig() (map[string]interface{}, error) {
	path, err := configFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg map[string]interface{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg == nil {
		cfg = make(map[string]interface{})
	}
	return cfg, nil
}

func saveRawConfig(cfg map[string]interface{}) error {
	path, err := configFilePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(path, out, 0600)
}

// setNestedValue sets a dotted key (e.g. "defaults.days") in a nested map.
func setNestedValue(m map[string]interface{}, key string, value string) {
	parts := strings.SplitN(key, ".", 2)
	if len(parts) == 1 {
		m[key] = value
		return
	}

	child, ok := m[parts[0]].(map[string]interface{})
	if !ok {
		child = make(map[string]interface{})
	}
	setNestedValue(child, parts[1], value)
	m[parts[0]] = child
}
