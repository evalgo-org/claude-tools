package mv

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	eve "eve.evalgo.org/common"
	"github.com/spf13/cobra"
)

// Options holds mv configuration
type Options struct {
	Force       bool
	NoClobber   bool
	Verbose     bool
	Interactive bool
}

// Command returns the mv command
func Command() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "mv [flags] source... destination",
		Short: "Move (rename) files and directories",
		Long: `Move source files/directories to destination.

If the last argument names an existing directory, mv moves each source
into that directory. Otherwise, if only two files are given, it renames
the first to the second.`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sources := args[:len(args)-1]
			dest := args[len(args)-1]

			return moveFiles(sources, dest, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Overwrite existing files without prompting")
	cmd.Flags().BoolVarP(&opts.NoClobber, "no-clobber", "n", false, "Do not overwrite existing files")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Explain what is being done")

	return cmd
}

// moveFiles moves source files to destination
func moveFiles(sources []string, dest string, opts *Options) error {
	// Check if -f and -n are both set
	if opts.Force && opts.NoClobber {
		return fmt.Errorf("cannot specify both -f and -n")
	}

	// Check if destination is a directory
	destInfo, destErr := os.Stat(dest)
	isDestDir := destErr == nil && destInfo.IsDir()

	// If multiple sources, destination must be a directory
	if len(sources) > 1 && !isDestDir {
		return fmt.Errorf("target '%s' is not a directory", dest)
	}

	for _, src := range sources {
		// Check if source exists
		srcInfo, err := os.Stat(src)
		if err != nil {
			eve.Logger.Error("Failed to stat", src, ":", err)
			return err
		}

		var targetPath string
		if isDestDir {
			targetPath = filepath.Join(dest, filepath.Base(src))
		} else {
			targetPath = dest
		}

		// Check if destination exists
		if _, err := os.Stat(targetPath); err == nil {
			if opts.NoClobber {
				if opts.Verbose {
					eve.Logger.Info("Skipping", src, "(destination exists)")
				}
				continue
			}
			if !opts.Force {
				return fmt.Errorf("'%s' already exists (use -f to overwrite)", targetPath)
			}
		}

		// Attempt to move using os.Rename (fast for same filesystem)
		err = os.Rename(src, targetPath)
		if err != nil {
			// If rename fails (likely cross-filesystem), fall back to copy+delete
			if linkErr, ok := err.(*os.LinkError); ok {
				eve.Logger.Debug("Rename failed, using copy+delete:", linkErr)
				if err := copyAndDelete(src, targetPath, srcInfo); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("failed to move '%s' to '%s': %w", src, targetPath, err)
			}
		}

		if opts.Verbose {
			fmt.Printf("'%s' -> '%s'\n", src, targetPath)
		}
	}

	return nil
}

// copyAndDelete copies a file/directory and then deletes the source
func copyAndDelete(src, dest string, srcInfo os.FileInfo) error {
	if srcInfo.IsDir() {
		// Recursively copy directory
		if err := copyDir(src, dest, srcInfo); err != nil {
			return fmt.Errorf("failed to copy directory: %w", err)
		}
		// Remove source directory
		if err := os.RemoveAll(src); err != nil {
			return fmt.Errorf("failed to remove source directory: %w", err)
		}
	} else {
		// Copy file
		if err := copyFile(src, dest, srcInfo); err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}
		// Remove source file
		if err := os.Remove(src); err != nil {
			return fmt.Errorf("failed to remove source file: %w", err)
		}
	}
	return nil
}

// copyFile copies a single file with permissions
func copyFile(src, dest string, srcInfo os.FileInfo) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy contents: %w", err)
	}

	// Preserve timestamps
	if err := os.Chtimes(dest, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
		return fmt.Errorf("failed to preserve timestamps: %w", err)
	}

	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dest string, srcInfo os.FileInfo) error {
	// Create destination directory
	if err := os.MkdirAll(dest, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("failed to get entry info: %w", err)
		}

		if entry.IsDir() {
			if err := copyDir(srcPath, destPath, info); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, destPath, info); err != nil {
				return err
			}
		}
	}

	// Preserve directory timestamps
	if err := os.Chtimes(dest, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
		return fmt.Errorf("failed to preserve directory timestamps: %w", err)
	}

	return nil
}
