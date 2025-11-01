package mkdir

import (
	"fmt"
	"os"
	"path/filepath"

	eve "eve.evalgo.org/common"
	"github.com/spf13/cobra"
)

// Options holds mkdir configuration
type Options struct {
	Parents bool
	Mode    os.FileMode
	Verbose bool
}

// Command returns the mkdir command
func Command() *cobra.Command {
	opts := &Options{
		Mode: 0755, // Default permissions: rwxr-xr-x
	}

	cmd := &cobra.Command{
		Use:   "mkdir [flags] directories...",
		Short: "Create directories",
		Long: `Create the DIRECTORY(ies), if they do not already exist.

Creates directories with the specified names. By default, intermediate
directories must already exist. Use -p to create parent directories as needed.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, dir := range args {
				if err := createDirectory(dir, opts); err != nil {
					eve.Logger.Error("Failed to create directory", dir, ":", err)
					return err
				}

				if opts.Verbose {
					fmt.Printf("created directory '%s'\n", dir)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.Parents, "parents", "p", false, "Create parent directories as needed, no error if existing")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Print a message for each created directory")
	cmd.Flags().Uint32VarP((*uint32)(&opts.Mode), "mode", "m", 0755, "Set file mode (as in chmod), default 0755")

	return cmd
}

// createDirectory creates a directory with the specified options
func createDirectory(path string, opts *Options) error {
	// Clean the path to normalize it
	path = filepath.Clean(path)

	// Check if directory already exists
	info, err := os.Stat(path)
	if err == nil {
		// Path exists
		if !info.IsDir() {
			return fmt.Errorf("'%s' exists but is not a directory", path)
		}

		// Directory already exists
		if opts.Parents {
			// With -p flag, existing directories are not an error
			return nil
		}
		return fmt.Errorf("cannot create directory '%s': File exists", path)
	}

	// If error is not "not exists", return it
	if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat '%s': %w", path, err)
	}

	// Create the directory
	if opts.Parents {
		// MkdirAll creates parent directories as needed
		if err := os.MkdirAll(path, opts.Mode); err != nil {
			return fmt.Errorf("failed to create directory '%s': %w", path, err)
		}
	} else {
		// Mkdir only creates the final directory
		if err := os.Mkdir(path, opts.Mode); err != nil {
			return fmt.Errorf("failed to create directory '%s': %w", path, err)
		}
	}

	return nil
}
