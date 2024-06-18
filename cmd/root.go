package cmd

import (
	"github.com/spf13/cobra"
)

var configFile string
var logLevel string
var logFormat string

func Run() error {
	rootCmd := &cobra.Command{
		Use:   "root",
		Short: "Root command",
	}
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Path to the YAML configuration file")
	err := rootCmd.MarkPersistentFlagRequired("config")
	if err != nil {
		return err
	}
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "v", "info", "Logger log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "Logger logs format (text, json)")

	serverCmd := buildServerCmd()
	rootCmd.AddCommand(serverCmd)
	return rootCmd.Execute()
}
