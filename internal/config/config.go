package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config represents the application configuration.
type Config struct {
	Credentials string            `mapstructure:"credentials"`
	Aliases     map[string]string `mapstructure:"aliases"`
	Defaults    DefaultsConfig    `mapstructure:"defaults"`
}

// DefaultsConfig holds default values for CLI flags.
type DefaultsConfig struct {
	Days   int    `mapstructure:"days"`
	Top    int    `mapstructure:"top"`
	Output string `mapstructure:"output"`
}

// Load reads configuration from viper and returns a Config.
func Load() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
