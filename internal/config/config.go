package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hackerspace/hacktrackmmu-cli/internal/version"
	"github.com/spf13/viper"
)

type Config struct {
	LogLevel string `mapstructure:"log_level"`
	APIUrl   string `mapstructure:"api_url"`
}

func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	v.SetDefault("log_level", "info")

	isDev := version.Version == "dev"

	if isDev {
		v.SetDefault("api_url", "http://localhost:3000/api/v1")
	} else {
		v.SetDefault("api_url", "https://hacktrackmmu.herokuapp.com/api/v1")
	}

	v.SetEnvPrefix("HACKTRACK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if isDev {
		if cfgFile != "" {
			v.SetConfigFile(cfgFile)
		} else {
			home, err := os.UserHomeDir()
			if err == nil {
				v.AddConfigPath(filepath.Join(home, ".config", "hacktrackmmu-cli"))
			}
			v.AddConfigPath(".")
			v.SetConfigName("config")
			v.SetConfigType("yaml")
		}

		if err := v.ReadInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !errors.As(err, &configFileNotFoundError) {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into config struct: %w", err)
	}

	return &cfg, nil
}
