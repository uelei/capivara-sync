/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "capivara-sync",
	Short: "capivara-sync is a backup tool for your files ",
	Long: `capivara-sync is a backup tool that use zts to reduce remote storage use,
	It can restore to a point in time.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
