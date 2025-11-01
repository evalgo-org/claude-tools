package cp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	eve "eve.evalgo.org/common"
	"github.com/spf13/cobra"
)

// Options holds cp configuration
type Options struct {
	Recursive bool
	Preserve  bool
	Verbose   bool
	Force     bool
}

// Command returns the cp command
func Command() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "cp [flags] source... destination",
		Short: "Copy files and directories",
		Long: `Copy source files/directories to destination.

If the last argument names an existing directory, cp copies each source
into that directory. Otherwise, if only two files are given, it copies
the first onto the second.`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sources := args[:len(args)-1]
			dest := args[len(args)-1]

			return copyFiles(sources, dest, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Recursive, "recursive", "r", false, "Copy directories recursively")
	cmd.Flags().BoolVarP(&opts.Preserve, "preserve", "p", false, "Preserve file attributes (mode, timestamps)")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Explain what is being done")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Overwrite existing files without prompting")

	return cmd
}

// copyFiles copies source files to destination
func copyFiles(sources []string, dest string, opts *Options) error {
	// Check if destination is a directory
	destInfo, destErr := os.Stat(dest)
	isDestDir := destErr == nil && destInfo.IsDir()

	// If multiple sources, destination must be a directory
	if len(sources) > 1 && !isDestDir {
		return fmt.Errorf("target '%s' is not a directory", dest)
	}

	for _, src := range sources {
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

		if srcInfo.IsDir() {
			if !opts.Recursive {
				return fmt.Errorf("'%s' is a directory (use -r to copy directories)", src)
			}

			if err := copyDir(src, targetPath, opts); err != nil {
				return err
			}
		} else {
			if err := copyFile(src, targetPath, opts); err != nil {
				return err
			}
		}

		if opts.Verbose {
			fmt.Printf("'%s' -> '%s'\n", src, targetPath)
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dest string, opts *Options) error {
	// Check if destination exists
	if _, err := os.Stat(dest); err == nil && !opts.Force {
		return fmt.Errorf("'%s' already exists (use -f to overwrite)", dest)
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source '%s': %w", src, err)
	}
	defer srcFile.Close()

	// Get source file info
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	// Create destination file
	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination '%s': %w", dest, err)
	}
	defer destFile.Close()

	// Copy contents
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy contents: %w", err)
	}

	// Preserve timestamps if requested
	if opts.Preserve {
		if err := os.Chtimes(dest, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
			return fmt.Errorf("failed to preserve timestamps: %w", err)
		}
	}

	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dest string, opts *Options) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

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

		if entry.IsDir() {
			if err := copyDir(srcPath, destPath, opts); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, destPath, opts); err != nil {
				return err
			}
		}
	}

	// Preserve directory timestamps if requested
	if opts.Preserve {
		if err := os.Chtimes(dest, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
			return fmt.Errorf("failed to preserve directory timestamps: %w", err)
		}
	}

	return nil
}
