package head

import (
	"bufio"
	"fmt"
	"io"
	"os"

	eve "eve.evalgo.org/common"
	"github.com/spf13/cobra"
)

// Options holds head configuration
type Options struct {
	Lines int
	Bytes int
	Quiet bool
}

// Command returns the head command
func Command() *cobra.Command {
	opts := &Options{
		Lines: 10, // Default to 10 lines
	}

	cmd := &cobra.Command{
		Use:   "head [flags] [files...]",
		Short: "Output the first part of files",
		Long:  `Print the first N lines (default 10) of each file to standard output. With no files, or when file is -, read standard input.`,
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			files := args

			// If no files specified, read from stdin
			if len(files) == 0 {
				return headReader(os.Stdin, opts, "", len(files) > 1)
			}

			// Process each file
			for i, file := range files {
				if file == "-" {
					if err := headReader(os.Stdin, opts, "standard input", len(files) > 1); err != nil {
						eve.Logger.Error("Failed to read stdin:", err)
					}
				} else {
					if err := headFile(file, opts, len(files) > 1); err != nil {
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

	cmd.Flags().IntVarP(&opts.Lines, "lines", "n", 10, "Print the first N lines")
	cmd.Flags().IntVarP(&opts.Bytes, "bytes", "c", 0, "Print the first N bytes")
	cmd.Flags().BoolVarP(&opts.Quiet, "quiet", "q", false, "Never print headers giving file names")

	return cmd
}

// headFile reads and displays the first part of a file
func headFile(filename string, opts *Options, multipleFiles bool) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return headReader(file, opts, filename, multipleFiles)
}

// headReader reads and displays the first part from a reader
func headReader(reader io.Reader, opts *Options, filename string, multipleFiles bool) error {
	// Print header if multiple files and not quiet
	if multipleFiles && !opts.Quiet && filename != "" {
		fmt.Printf("==> %s <==\n", filename)
	}

	// Handle byte mode
	if opts.Bytes > 0 {
		return headBytes(reader, opts.Bytes)
	}

	// Handle line mode (default)
	scanner := bufio.NewScanner(reader)
	lineCount := 0

	for scanner.Scan() && lineCount < opts.Lines {
		fmt.Println(scanner.Text())
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}

// headBytes reads and displays the first N bytes
func headBytes(reader io.Reader, n int) error {
	buf := make([]byte, n)
	bytesRead, err := io.ReadFull(reader, buf)

	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return fmt.Errorf("error reading bytes: %w", err)
	}

	// Write exactly the bytes we read
	if _, err := os.Stdout.Write(buf[:bytesRead]); err != nil {
		return fmt.Errorf("error writing output: %w", err)
	}

	return nil
}
