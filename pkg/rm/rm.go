package rm

import (
	"fmt"
	"os"
	"path/filepath"

	eve "eve.evalgo.org/common"
	"github.com/spf13/cobra"
)

// Options holds rm configuration
type Options struct {
	Recursive bool
	Force     bool
	Verbose   bool
}

// Command returns the rm command
func Command() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "rm [flags] files...",
		Short: "Remove files or directories",
		Long: `Remove (delete) files and directories.

By default, rm does not remove directories. Use -r to remove directories
and their contents recursively.

WARNING: Deleted files cannot be recovered. Use with caution.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, path := range args {
				if err := removePath(path, opts); err != nil {
					if !opts.Force {
						eve.Logger.Error("Failed to remove", path, ":", err)
						return err
					}
					// With -f, continue on errors
					if opts.Verbose {
						eve.Logger.Warn("Failed to remove", path, ":", err)
					}
				} else if opts.Verbose {
					fmt.Printf("removed '%s'\n", path)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.Recursive, "recursive", "r", false, "Remove directories and their contents recursively")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Ignore nonexistent files and never prompt")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Explain what is being done")

	return cmd
}

// removePath removes a file or directory
func removePath(path string, opts *Options) error {
	// Clean the path
	path = filepath.Clean(path)

	// Get file info
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) && opts.Force {
			// With -f, nonexistent files are not an error
			return nil
		}
		return fmt.Errorf("failed to stat '%s': %w", path, err)
	}

	// Check if it's a directory
	if info.IsDir() {
		if !opts.Recursive {
			return fmt.Errorf("cannot remove '%s': Is a directory (use -r to remove directories)", path)
		}

		// Remove directory recursively
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove directory '%s': %w", path, err)
		}
	} else {
		// Remove file
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove '%s': %w", path, err)
		}
	}

	return nil
}
