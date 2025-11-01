package uniq

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Options holds uniq configuration
type Options struct {
	Count      bool
	Repeated   bool
	Unique     bool
	IgnoreCase bool
}

// Command returns the uniq command
func Command() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "uniq [flags] [input [output]]",
		Short: "Report or omit repeated lines",
		Long:  `Filter adjacent matching lines from input (or standard input), writing to output (or standard output).`,
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var input io.Reader = os.Stdin
			var output io.Writer = os.Stdout

			// Open input file if specified
			if len(args) >= 1 && args[0] != "-" {
				file, err := os.Open(args[0])
				if err != nil {
					return fmt.Errorf("failed to open input file: %w", err)
				}
				defer file.Close()
				input = file
			}

			// Open output file if specified
			if len(args) >= 2 {
				file, err := os.Create(args[1])
				if err != nil {
					return fmt.Errorf("failed to create output file: %w", err)
				}
				defer file.Close()
				output = file
			}

			return processUniq(input, output, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Count, "count", "c", false, "Prefix lines by the number of occurrences")
	cmd.Flags().BoolVarP(&opts.Repeated, "repeated", "d", false, "Only print duplicate lines, one for each group")
	cmd.Flags().BoolVarP(&opts.Unique, "unique", "u", false, "Only print unique lines")
	cmd.Flags().BoolVarP(&opts.IgnoreCase, "ignore-case", "i", false, "Ignore differences in case when comparing")

	return cmd
}

// processUniq processes input and writes unique lines to output
func processUniq(input io.Reader, output io.Writer, opts *Options) error {
	scanner := bufio.NewScanner(input)
	writer := bufio.NewWriter(output)
	defer writer.Flush()

	if !scanner.Scan() {
		// Empty input
		return nil
	}

	currentLine := scanner.Text()
	currentCount := 1
	currentCompareLine := getCompareLine(currentLine, opts)

	for scanner.Scan() {
		line := scanner.Text()
		compareLine := getCompareLine(line, opts)

		if compareLine == currentCompareLine {
			// Same as previous line
			currentCount++
		} else {
			// Different line - output previous group
			if err := outputLine(writer, currentLine, currentCount, opts); err != nil {
				return err
			}

			// Start new group
			currentLine = line
			currentCompareLine = compareLine
			currentCount = 1
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	// Output last group
	if err := outputLine(writer, currentLine, currentCount, opts); err != nil {
		return err
	}

	return nil
}

// getCompareLine returns the line to use for comparison
func getCompareLine(line string, opts *Options) string {
	if opts.IgnoreCase {
		return strings.ToLower(line)
	}
	return line
}

// outputLine outputs a line according to options
func outputLine(writer io.Writer, line string, count int, opts *Options) error {
	// Apply filtering
	if opts.Repeated && count == 1 {
		// Skip unique lines when -d flag
		return nil
	}
	if opts.Unique && count > 1 {
		// Skip repeated lines when -u flag
		return nil
	}

	// Format output
	var output string
	if opts.Count {
		output = fmt.Sprintf("%7d %s\n", count, line)
	} else {
		output = fmt.Sprintf("%s\n", line)
	}

	if _, err := fmt.Fprint(writer, output); err != nil {
		return fmt.Errorf("error writing output: %w", err)
	}

	return nil
}
