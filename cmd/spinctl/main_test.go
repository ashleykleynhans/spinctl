package main

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommandHelp(t *testing.T) {
	cmd := rootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	if !containsStr(output, "spinctl") {
		t.Error("help should mention spinctl")
	}
}

func TestDeployCommand(t *testing.T) {
	cmd := rootCmd()
	deploy := findSubcommand(cmd, "deploy")
	if deploy == nil {
		t.Fatal("deploy subcommand not found")
	}
}

func TestImportCommand(t *testing.T) {
	cmd := rootCmd()
	imp := findSubcommand(cmd, "import")
	if imp == nil {
		t.Fatal("import subcommand not found")
	}
}

func TestShowCommand(t *testing.T) {
	cmd := rootCmd()
	show := findSubcommand(cmd, "show")
	if show == nil {
		t.Fatal("show subcommand not found")
	}
}

func findSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, c := range cmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
