package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/hackerspace/hacktrackmmu-cli/internal/version"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information of ht",
	Long:  `Print detailed version and build information for this CLI tool.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		info := version.GetInfo()

		if jsonOutput {
			bz, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal version info to JSON: %w", err)
			}
			fmt.Println(string(bz))
			return nil
		}

		fmt.Println(info.String())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output version information in JSON format")
}
