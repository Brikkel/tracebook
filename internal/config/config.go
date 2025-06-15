package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	AnnotateDirectory bool     `mapstructure:"annotate_directory"`
	OutputDir         string   `mapstructure:"output_dir"`
	TempDir           string   `mapstructure:"temp_dir"`
	SkipCommands      []string `mapstructure:"skip_commands"`
}

// LoadConfig loads configuration from file and environment variables.
func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot find home directory: %w", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Join(home, ".config", "tracebook"))
	viper.AddConfigPath(".") // Optionally, also look in current directory

	// Set defaults
	viper.SetDefault("annotate_directory", true)
	viper.SetDefault("output_dir", filepath.Join(home, "tracebook_sessions"))
	viper.SetDefault("temp_dir", "/tmp/tracebook")

	viper.AutomaticEnv() // Override with environment variables if set

	if err := viper.ReadInConfig(); err != nil {
		// If config file not found, that's OK; use defaults and env
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}
	return &cfg, nil
}
