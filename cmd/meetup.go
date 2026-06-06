/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import "github.com/spf13/cobra"

var meetupCmd = &cobra.Command{
	Use:   "meetup",
	Short: "create a meetup",
	Long: `create a meetup to hacktrackmmu interactively`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Error: 'meetup' command requires a subcommand. Please choose one of the available subcommands listed below:")
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(meetupCmd)
}
