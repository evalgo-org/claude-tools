package find

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	eve "eve.evalgo.org/common"
	"github.com/spf13/cobra"
)

// Options holds find configuration
type Options struct {
	Name     string
	IName    string
	Type     string
	MaxDepth int
	MinDepth int
}

// Command returns the find command
func Command() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "find [path...] [flags]",
		Short: "Find files and directories",
		Long:  `Find files and directories by name, type, or other criteria.`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := args
			if len(paths) == 0 {
				paths = []string{"."}
			}

			for _, path := range paths {
				if err := findPath(path, opts, 0); err != nil {
					eve.Logger.Error("Failed to search path", path, ":", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Find by name pattern (case-sensitive)")
	cmd.Flags().StringVar(&opts.IName, "iname", "", "Find by name pattern (case-insensitive)")
	cmd.Flags().StringVarP(&opts.Type, "type", "t", "", "Find by type (f=file, d=directory, l=symlink)")
	cmd.Flags().IntVar(&opts.MaxDepth, "maxdepth", -1, "Maximum depth to search")
	cmd.Flags().IntVar(&opts.MinDepth, "mindepth", 0, "Minimum depth to search")

	return cmd
}

// findPath recursively searches a path
func findPath(root string, opts *Options, depth int) error {
	// Check depth constraints
	if opts.MaxDepth >= 0 && depth > opts.MaxDepth {
		return nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		fullPath := filepath.Join(root, entry.Name())

		// Check if this entry matches our criteria
		if shouldPrint(entry, fullPath, opts, depth) {
			fmt.Println(fullPath)
		}

		// Recurse into directories
		if entry.IsDir() {
			if err := findPath(fullPath, opts, depth+1); err != nil {
				eve.Logger.Error("Failed to search directory", fullPath, ":", err)
			}
		}
	}

	return nil
}

// shouldPrint determines if an entry should be printed
func shouldPrint(entry os.DirEntry, path string, opts *Options, depth int) bool {
	// Check minimum depth
	if depth < opts.MinDepth {
		return false
	}

	// Check type filter
	if opts.Type != "" {
		info, err := entry.Info()
		if err != nil {
			return false
		}

		switch opts.Type {
		case "f":
			if !info.Mode().IsRegular() {
				return false
			}
		case "d":
			if !info.IsDir() {
				return false
			}
		case "l":
			if info.Mode()&os.ModeSymlink == 0 {
				return false
			}
		}
	}

	// Check name filter (case-sensitive)
	if opts.Name != "" {
		matched, err := filepath.Match(opts.Name, entry.Name())
		if err != nil || !matched {
			return false
		}
	}

	// Check name filter (case-insensitive)
	if opts.IName != "" {
		pattern := strings.ToLower(opts.IName)
		name := strings.ToLower(entry.Name())
		matched, err := filepath.Match(pattern, name)
		if err != nil || !matched {
			return false
		}
	}

	return true
}
