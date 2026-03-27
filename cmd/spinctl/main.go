package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spinctl",
		Short: "Spinnaker configuration tool",
		Long:  "spinctl is a terminal UI tool for managing Spinnaker configuration and deployment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("spinctl TUI - coming soon")
			return nil
		},
	}
	return cmd
}
