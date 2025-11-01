package wc

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unicode"

	eve "eve.evalgo.org/common"
	"github.com/spf13/cobra"
)

// Options holds wc configuration
type Options struct {
	Lines      bool
	Words      bool
	Chars      bool
	Bytes      bool
	MaxLineLen bool
}

// Counts holds the counts for a file
type Counts struct {
	Lines      int64
	Words      int64
	Chars      int64
	Bytes      int64
	MaxLineLen int64
}

// Command returns the wc command
func Command() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "wc [flags] [files...]",
		Short: "Print newline, word, and byte counts for each file",
		Long:  `Print newline, word, and byte counts for each file. With no files, or when file is -, read standard input.`,
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no flags specified, default to lines, words, and bytes
			if !opts.Lines && !opts.Words && !opts.Chars && !opts.Bytes && !opts.MaxLineLen {
				opts.Lines = true
				opts.Words = true
				opts.Bytes = true
			}

			files := args
			if len(files) == 0 {
				files = []string{"-"}
			}

			totalCounts := &Counts{}
			multipleFiles := len(files) > 1

			// Process each file
			for _, file := range files {
				var counts *Counts
				var err error
				var name string

				if file == "-" {
					counts, err = countReader(os.Stdin, opts)
					name = ""
				} else {
					counts, err = countFile(file, opts)
					name = file
				}

				if err != nil {
					eve.Logger.Error("Failed to count", file, ":", err)
					continue
				}

				printCounts(counts, opts, name)

				// Add to totals
				if multipleFiles {
					totalCounts.Lines += counts.Lines
					totalCounts.Words += counts.Words
					totalCounts.Chars += counts.Chars
					totalCounts.Bytes += counts.Bytes
					if counts.MaxLineLen > totalCounts.MaxLineLen {
						totalCounts.MaxLineLen = counts.MaxLineLen
					}
				}
			}

			// Print totals if multiple files
			if multipleFiles {
				printCounts(totalCounts, opts, "total")
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.Lines, "lines", "l", false, "Print the newline counts")
	cmd.Flags().BoolVarP(&opts.Words, "words", "w", false, "Print the word counts")
	cmd.Flags().BoolVarP(&opts.Chars, "chars", "m", false, "Print the character counts")
	cmd.Flags().BoolVarP(&opts.Bytes, "bytes", "c", false, "Print the byte counts")
	cmd.Flags().BoolVarP(&opts.MaxLineLen, "max-line-length", "L", false, "Print the maximum display width")

	return cmd
}

// countFile counts lines, words, and bytes in a file
func countFile(filename string, opts *Options) (*Counts, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return countReader(file, opts)
}

// countReader counts lines, words, and bytes from a reader
func countReader(reader io.Reader, opts *Options) (*Counts, error) {
	counts := &Counts{}
	scanner := bufio.NewScanner(reader)

	inWord := false

	for scanner.Scan() {
		line := scanner.Text()
		counts.Lines++

		// Count bytes (including newline)
		counts.Bytes += int64(len(line)) + 1 // +1 for newline

		// Count characters
		lineLen := int64(0)
		for _, r := range line {
			counts.Chars++
			lineLen++

			// Count words
			if unicode.IsSpace(r) {
				inWord = false
			} else {
				if !inWord {
					counts.Words++
					inWord = true
				}
			}
		}

		// Track max line length
		if lineLen > counts.MaxLineLen {
			counts.MaxLineLen = lineLen
		}

		// Reset word state for next line
		inWord = false
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	return counts, nil
}

// printCounts prints the counts according to options
func printCounts(counts *Counts, opts *Options, filename string) {
	output := ""

	if opts.Lines {
		output += fmt.Sprintf("%8d", counts.Lines)
	}
	if opts.Words {
		output += fmt.Sprintf("%8d", counts.Words)
	}
	if opts.Chars {
		output += fmt.Sprintf("%8d", counts.Chars)
	}
	if opts.Bytes {
		output += fmt.Sprintf("%8d", counts.Bytes)
	}
	if opts.MaxLineLen {
		output += fmt.Sprintf("%8d", counts.MaxLineLen)
	}

	if filename != "" {
		output += " " + filename
	}

	fmt.Println(output)
}
