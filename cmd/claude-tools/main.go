package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/evalgo-org/claude-tools/pkg/awk"
	"github.com/evalgo-org/claude-tools/pkg/cat"
	"github.com/evalgo-org/claude-tools/pkg/cp"
	"github.com/evalgo-org/claude-tools/pkg/db"
	"github.com/evalgo-org/claude-tools/pkg/find"
	"github.com/evalgo-org/claude-tools/pkg/grep"
	"github.com/evalgo-org/claude-tools/pkg/head"
	"github.com/evalgo-org/claude-tools/pkg/jq"
	"github.com/evalgo-org/claude-tools/pkg/ls"
	"github.com/evalgo-org/claude-tools/pkg/mkdir"
	"github.com/evalgo-org/claude-tools/pkg/mv"
	"github.com/evalgo-org/claude-tools/pkg/rm"
	"github.com/evalgo-org/claude-tools/pkg/sed"
	"github.com/evalgo-org/claude-tools/pkg/sort"
	"github.com/evalgo-org/claude-tools/pkg/tail"
	"github.com/evalgo-org/claude-tools/pkg/touch"
	"github.com/evalgo-org/claude-tools/pkg/tree"
	"github.com/evalgo-org/claude-tools/pkg/uniq"
	"github.com/evalgo-org/claude-tools/pkg/wc"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "claude-tools",
		Short: "Cross-platform CLI tools for development",
		Long: `claude-tools provides cross-platform implementations of common Linux/Unix tools.
Built in Go for consistent behavior across Windows, Linux, and macOS.`,
		Version: "0.5.1",
	}

	// Add subcommands - Phase 1
	rootCmd.AddCommand(grep.Command())
	rootCmd.AddCommand(find.Command())
	rootCmd.AddCommand(cat.Command())

	// Add subcommands - Phase 2
	rootCmd.AddCommand(head.Command())
	rootCmd.AddCommand(tail.Command())
	rootCmd.AddCommand(wc.Command())
	rootCmd.AddCommand(ls.Command())
	rootCmd.AddCommand(sort.Command())
	rootCmd.AddCommand(uniq.Command())

	// Add subcommands - Phase 3
	rootCmd.AddCommand(tree.Command())
	rootCmd.AddCommand(jq.Command())
	rootCmd.AddCommand(sed.Command())
	rootCmd.AddCommand(awk.Command())

	// Add subcommands - Phase 4
	rootCmd.AddCommand(db.Command())

	// Add subcommands - Phase 5
	rootCmd.AddCommand(mkdir.Command())

	// Add subcommands - Phase 6 (File operations)
	rootCmd.AddCommand(rm.Command())
	rootCmd.AddCommand(cp.Command())
	rootCmd.AddCommand(mv.Command())
	rootCmd.AddCommand(touch.Command())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
