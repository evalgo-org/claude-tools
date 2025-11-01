package tail

import (
	"bufio"
	"fmt"
	"io"
	"os"

	eve "eve.evalgo.org/common"
	"github.com/spf13/cobra"
)

// Options holds tail configuration
type Options struct {
	Lines int
	Bytes int
	Quiet bool
}

// Command returns the tail command
func Command() *cobra.Command {
	opts := &Options{
		Lines: 10, // Default to 10 lines
	}

	cmd := &cobra.Command{
		Use:   "tail [flags] [files...]",
		Short: "Output the last part of files",
		Long:  `Print the last N lines (default 10) of each file to standard output. With no files, or when file is -, read standard input.`,
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			files := args

			// If no files specified, read from stdin
			if len(files) == 0 {
				return tailReader(os.Stdin, opts, "", len(files) > 1)
			}

			// Process each file
			for i, file := range files {
				if file == "-" {
					if err := tailReader(os.Stdin, opts, "standard input", len(files) > 1); err != nil {
						eve.Logger.Error("Failed to read stdin:", err)
					}
				} else {
					if err := tailFile(file, opts, len(files) > 1); err != nil {
						eve.Logger.Error("Failed to read file", file, ":", err)
					}
				}

				// Add blank line between files (except after last)
				if i < len(files)-1 && len(files) > 1 {
					fmt.Println()
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&opts.Lines, "lines", "n", 10, "Output the last N lines")
	cmd.Flags().IntVarP(&opts.Bytes, "bytes", "c", 0, "Output the last N bytes")
	cmd.Flags().BoolVarP(&opts.Quiet, "quiet", "q", false, "Never print headers giving file names")

	return cmd
}

// tailFile reads and displays the last part of a file
func tailFile(filename string, opts *Options, multipleFiles bool) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return tailReader(file, opts, filename, multipleFiles)
}

// tailReader reads and displays the last part from a reader
func tailReader(reader io.Reader, opts *Options, filename string, multipleFiles bool) error {
	// Print header if multiple files and not quiet
	if multipleFiles && !opts.Quiet && filename != "" {
		fmt.Printf("==> %s <==\n", filename)
	}

	// Handle byte mode
	if opts.Bytes > 0 {
		return tailBytes(reader, opts.Bytes)
	}

	// Handle line mode (default)
	// Read all lines into a circular buffer
	lines := make([]string, opts.Lines)
	scanner := bufio.NewScanner(reader)
	index := 0
	count := 0

	for scanner.Scan() {
		lines[index%opts.Lines] = scanner.Text()
		index++
		count++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	// Print the last N lines in correct order
	start := 0
	numLines := opts.Lines
	if count < opts.Lines {
		// We have fewer lines than requested
		numLines = count
	} else {
		// Start from the oldest line in the circular buffer
		start = index % opts.Lines
	}

	for i := 0; i < numLines; i++ {
		fmt.Println(lines[(start+i)%opts.Lines])
	}

	return nil
}

// tailBytes reads and displays the last N bytes
func tailBytes(reader io.Reader, n int) error {
	// Read all content
	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	// Calculate start position
	start := 0
	if len(content) > n {
		start = len(content) - n
	}

	// Write the last N bytes
	if _, err := os.Stdout.Write(content[start:]); err != nil {
		return fmt.Errorf("error writing output: %w", err)
	}

	return nil
}
