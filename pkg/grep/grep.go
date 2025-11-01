package grep

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	eve "eve.evalgo.org/common"
	"github.com/spf13/cobra"
)

// Options holds grep configuration
type Options struct {
	CaseInsensitive bool
	Recursive       bool
	LineNumbers     bool
	ContextBefore   int
	ContextAfter    int
	Invert          bool
	FilesOnly       bool
	Count           bool
}

// Command returns the grep command
func Command() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "grep [flags] pattern [files...]",
		Short: "Search for patterns in files",
		Long:  `Search for patterns in files using regular expressions. Compatible with common grep flags.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pattern := args[0]
			files := args[1:]

			// If no files specified, read from stdin
			if len(files) == 0 {
				return grepReader(os.Stdin, pattern, opts, "<stdin>")
			}

			// If recursive, expand directories
			if opts.Recursive {
				expanded, err := expandDirs(files)
				if err != nil {
					return fmt.Errorf("failed to expand directories: %w", err)
				}
				files = expanded
			}

			// Process each file
			for _, file := range files {
				if err := grepFile(file, pattern, opts); err != nil {
					eve.Logger.Error("Failed to grep file", file, ":", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.CaseInsensitive, "ignore-case", "i", false, "Case insensitive search")
	cmd.Flags().BoolVarP(&opts.Recursive, "recursive", "r", false, "Search recursively in directories")
	cmd.Flags().BoolVarP(&opts.LineNumbers, "line-number", "n", false, "Show line numbers")
	cmd.Flags().IntVarP(&opts.ContextBefore, "before-context", "B", 0, "Show N lines before match")
	cmd.Flags().IntVarP(&opts.ContextAfter, "after-context", "A", 0, "Show N lines after match")
	cmd.Flags().IntVarP(&opts.ContextAfter, "context", "C", 0, "Show N lines before and after match")
	cmd.Flags().BoolVarP(&opts.Invert, "invert-match", "v", false, "Invert match (show non-matching lines)")
	cmd.Flags().BoolVarP(&opts.FilesOnly, "files-with-matches", "l", false, "Show only filenames with matches")
	cmd.Flags().BoolVarP(&opts.Count, "count", "c", false, "Show count of matching lines")

	return cmd
}

// grepFile searches for pattern in a file
func grepFile(filename, pattern string, opts *Options) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return grepReader(file, pattern, opts, filename)
}

// grepReader searches for pattern in a reader
func grepReader(reader *os.File, pattern string, opts *Options, filename string) error {
	// Compile regex
	flags := ""
	if opts.CaseInsensitive {
		flags = "(?i)"
	}
	re, err := regexp.Compile(flags + pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	scanner := bufio.NewScanner(reader)
	lineNum := 0
	matchCount := 0
	foundMatch := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		matches := re.MatchString(line)

		// Invert logic if requested
		if opts.Invert {
			matches = !matches
		}

		if matches {
			matchCount++
			foundMatch = true

			// Files-only mode: just record that we found a match
			if opts.FilesOnly {
				fmt.Println(filename)
				return nil
			}

			// Count mode: just count
			if opts.Count {
				continue
			}

			// Regular output
			prefix := ""
			if filename != "<stdin>" {
				prefix = filename + ":"
			}
			if opts.LineNumbers {
				prefix += fmt.Sprintf("%d:", lineNum)
			}

			fmt.Printf("%s%s\n", prefix, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Print count if requested
	if opts.Count && foundMatch {
		prefix := ""
		if filename != "<stdin>" {
			prefix = filename + ":"
		}
		fmt.Printf("%s%d\n", prefix, matchCount)
	}

	return nil
}

// expandDirs recursively expands directories to file list
func expandDirs(paths []string) ([]string, error) {
	var files []string

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("failed to stat %s: %w", path, err)
		}

		if info.IsDir() {
			err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					files = append(files, walkPath)
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("failed to walk directory %s: %w", path, err)
			}
		} else {
			files = append(files, path)
		}
	}

	return files, nil
}
