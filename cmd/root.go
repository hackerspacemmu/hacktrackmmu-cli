package cmd

import (
	"fmt"
	"os"

	"github.com/hackerspace/hacktrackmmu-cli/internal/config"
	"github.com/hackerspace/hacktrackmmu-cli/internal/logger"
	"github.com/hackerspace/hacktrackmmu-cli/internal/session"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
	cfg     *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "hacktrackmmu-cli",
	Short: "Hacktrack MMU CLI Tool",
	Long: `A high-performance CLI utility for the MMU Hacktrack platform.
Configure it using environment variables or a YAML config file.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		loadedCfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		cfg = loadedCfg

		if verbose {
			cfg.LogLevel = "debug"
		}
		logger.Init(cfg.LogLevel)

		if cmd.Name() == "login" || cmd.Name() == "help" || cmd.Name() == "version" {
			return nil
		}

		s, err := session.Load()
		if err != nil || s == nil || !s.IsValid() {
			fmt.Println("WARNING: Session token is missing or has expired!")
			fmt.Println("Please run the login command to authenticate:")
			fmt.Println("  hacktrackmmu-cli login [password]")
			fmt.Println()
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml or $HOME/.config/hacktrackmmu-cli/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose/debug logging output")

	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func GetConfig() *config.Config {
	return cfg
}
