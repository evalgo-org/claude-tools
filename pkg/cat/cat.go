package cat

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	eve "eve.evalgo.org/common"
	"github.com/spf13/cobra"
)

// Options holds cat configuration
type Options struct {
	NumberLines     bool
	ShowNonPrinting bool
	SqueezeBlank    bool
}

// Command returns the cat command
func Command() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "cat [flags] [files...]",
		Short: "Concatenate and display file contents",
		Long:  `Concatenate files and print on the standard output. Compatible with common cat flags.`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			files := args

			// If no files specified, read from stdin
			if len(files) == 0 {
				return catReader(os.Stdin, opts, false)
			}

			// Process each file
			for _, file := range files {
				if err := catFile(file, opts); err != nil {
					eve.Logger.Error("Failed to cat file", file, ":", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.NumberLines, "number", "n", false, "Number all output lines")
	cmd.Flags().BoolVarP(&opts.ShowNonPrinting, "show-all", "A", false, "Show non-printing characters")
	cmd.Flags().BoolVarP(&opts.SqueezeBlank, "squeeze-blank", "s", false, "Squeeze multiple blank lines")

	return cmd
}

// catFile reads and displays a file
func catFile(filename string, opts *Options) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return catReader(file, opts, true)
}

// catReader reads and displays content from a reader
func catReader(file *os.File, opts *Options, showFilename bool) error {
	scanner := bufio.NewScanner(file)
	lineNum := 0
	lastLineBlank := false

	for scanner.Scan() {
		line := scanner.Text()
		isBlank := strings.TrimSpace(line) == ""

		// Handle squeeze blank option
		if opts.SqueezeBlank && isBlank && lastLineBlank {
			continue
		}
		lastLineBlank = isBlank

		lineNum++

		// Build output line
		output := ""

		// Add line numbers if requested
		if opts.NumberLines {
			output = fmt.Sprintf("%6d  ", lineNum)
		}

		// Process line content
		if opts.ShowNonPrinting {
			output += showNonPrintingChars(line)
		} else {
			output += line
		}

		fmt.Println(output)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}

// showNonPrintingChars replaces non-printing characters with visible representations
func showNonPrintingChars(line string) string {
	var result strings.Builder

	for _, r := range line {
		switch r {
		case '\t':
			result.WriteString("^I")
		case '\r':
			result.WriteString("^M")
		case '\n':
			result.WriteString("$\n")
		default:
			if r < 32 || r == 127 {
				// Control characters
				if r == 127 {
					result.WriteString("^?")
				} else {
					result.WriteString(fmt.Sprintf("^%c", r+64))
				}
			} else if r > 127 {
				// Non-ASCII
				result.WriteString(fmt.Sprintf("M-%c", r-128))
			} else {
				result.WriteRune(r)
			}
		}
	}

	return result.String()
}
