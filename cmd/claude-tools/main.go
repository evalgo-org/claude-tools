package main

import (
	"os"

	"github.com/evalgo-org/claude-tools/pkg/cat"
	"github.com/evalgo-org/claude-tools/pkg/find"
	"github.com/evalgo-org/claude-tools/pkg/grep"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "claude-tools",
		Short: "Cross-platform CLI tools for development",
		Long: `claude-tools provides cross-platform implementations of common Linux/Unix tools.
Built in Go for consistent behavior across Windows, Linux, and macOS.`,
		Version: "0.1.0",
	}

	// Add subcommands
	rootCmd.AddCommand(grep.Command())
	rootCmd.AddCommand(find.Command())
	rootCmd.AddCommand(cat.Command())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
