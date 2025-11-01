package sort

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	eve "eve.evalgo.org/common"
	"github.com/spf13/cobra"
)

// Options holds sort configuration
type Options struct {
	Reverse        bool
	Numeric        bool
	Unique         bool
	IgnoreCase     bool
	Key            int
	FieldSeparator string
}

// Command returns the sort command
func Command() *cobra.Command {
	opts := &Options{
		FieldSeparator: " ", // Default to space
	}

	cmd := &cobra.Command{
		Use:   "sort [flags] [files...]",
		Short: "Sort lines of text files",
		Long:  `Sort lines of text files. With no files, or when file is -, read standard input.`,
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			files := args
			if len(files) == 0 {
				files = []string{"-"}
			}

			// Collect all lines from all files
			var allLines []string

			for _, file := range files {
				var lines []string
				var err error

				if file == "-" {
					lines, err = readLines(os.Stdin)
				} else {
					lines, err = readFile(file)
				}

				if err != nil {
					eve.Logger.Error("Failed to read", file, ":", err)
					continue
				}

				allLines = append(allLines, lines...)
			}

			// Sort the lines
			sortedLines := sortLines(allLines, opts)

			// Print sorted lines
			for _, line := range sortedLines {
				fmt.Println(line)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.Reverse, "reverse", "r", false, "Reverse the result of comparisons")
	cmd.Flags().BoolVarP(&opts.Numeric, "numeric-sort", "n", false, "Compare according to string numerical value")
	cmd.Flags().BoolVarP(&opts.Unique, "unique", "u", false, "Output only the first of an equal run")
	cmd.Flags().BoolVarP(&opts.IgnoreCase, "ignore-case", "f", false, "Fold lower case to upper case characters")
	cmd.Flags().IntVarP(&opts.Key, "key", "k", 0, "Sort via a key; 1-indexed field number")
	cmd.Flags().StringVarP(&opts.FieldSeparator, "field-separator", "t", " ", "Use SEP instead of non-blank to blank transition")

	return cmd
}

// readFile reads all lines from a file
func readFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return readLines(file)
}

// readLines reads all lines from a reader
func readLines(reader io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	return lines, nil
}

// sortLines sorts lines according to options
func sortLines(lines []string, opts *Options) []string {
	// Make a copy to avoid modifying original
	sorted := make([]string, len(lines))
	copy(sorted, lines)

	sort.SliceStable(sorted, func(i, j int) bool {
		line1 := sorted[i]
		line2 := sorted[j]

		// Extract key fields if specified
		if opts.Key > 0 {
			line1 = extractKey(line1, opts.Key, opts.FieldSeparator)
			line2 = extractKey(line2, opts.Key, opts.FieldSeparator)
		}

		// Apply case folding if requested
		if opts.IgnoreCase {
			line1 = strings.ToUpper(line1)
			line2 = strings.ToUpper(line2)
		}

		// Compare
		var result bool
		if opts.Numeric {
			num1, err1 := strconv.ParseFloat(strings.TrimSpace(line1), 64)
			num2, err2 := strconv.ParseFloat(strings.TrimSpace(line2), 64)

			if err1 == nil && err2 == nil {
				result = num1 < num2
			} else {
				// Fall back to string comparison if not valid numbers
				result = line1 < line2
			}
		} else {
			result = line1 < line2
		}

		// Reverse if requested
		if opts.Reverse {
			return !result
		}
		return result
	})

	// Apply unique filter if requested
	if opts.Unique {
		return uniqueLines(sorted, opts)
	}

	return sorted
}

// extractKey extracts the Nth field from a line
func extractKey(line string, keyNum int, separator string) string {
	fields := strings.Split(line, separator)

	// Adjust for 1-indexed keys
	index := keyNum - 1
	if index < 0 || index >= len(fields) {
		return line
	}

	return fields[index]
}

// uniqueLines removes consecutive duplicate lines
func uniqueLines(lines []string, opts *Options) []string {
	if len(lines) == 0 {
		return lines
	}

	unique := []string{lines[0]}
	lastLine := lines[0]

	if opts.IgnoreCase {
		lastLine = strings.ToUpper(lastLine)
	}

	for i := 1; i < len(lines); i++ {
		currentLine := lines[i]
		compareLine := currentLine

		if opts.IgnoreCase {
			compareLine = strings.ToUpper(compareLine)
		}

		if compareLine != lastLine {
			unique = append(unique, currentLine)
			lastLine = compareLine
		}
	}

	return unique
}
