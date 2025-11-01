package sed

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// Options holds sed configuration
type Options struct {
	InPlace    bool
	Quiet      bool
	Extended   bool
	Expression string
	LineNumber int
}

// Command returns the sed command
func Command() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "sed [options] 'command' [file...]",
		Short: "Stream editor for filtering and transforming text",
		Long: `Stream editor for filtering and transforming text.
Supports basic sed commands with simplified syntax.

Commands:
  s/pattern/replacement/[g]  Substitute
  /pattern/d                 Delete matching lines
  /pattern/p                 Print matching lines
  [line]d                    Delete specific line
  [line]p                    Print specific line

Examples:
  sed 's/foo/bar/' file.txt          Replace first foo with bar
  sed 's/foo/bar/g' file.txt         Replace all foo with bar
  sed '/pattern/d' file.txt          Delete lines matching pattern
  sed '5d' file.txt                  Delete line 5
  sed -n '/pattern/p' file.txt       Print only matching lines`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Expression = args[0]
			files := args[1:]

			if len(files) == 0 {
				return processInput(os.Stdin, opts, "")
			}

			for _, file := range files {
				if err := processFile(file, opts); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.InPlace, "in-place", "i", false, "Edit files in place")
	cmd.Flags().BoolVarP(&opts.Quiet, "quiet", "n", false, "Suppress automatic printing")
	cmd.Flags().BoolVarP(&opts.Extended, "extended", "E", false, "Use extended regex")

	return cmd
}

// processFile processes a file
func processFile(filename string, opts *Options) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open '%s': %w", filename, err)
	}
	defer file.Close()

	if opts.InPlace {
		return processInPlace(file, filename, opts)
	}

	return processInput(file, opts, filename)
}

// processInPlace edits file in place
func processInPlace(file *os.File, filename string, opts *Options) error {
	// Read entire file
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	file.Close()

	// Process lines
	result, err := processLines(lines, opts)
	if err != nil {
		return err
	}

	// Write back
	output, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("cannot write '%s': %w", filename, err)
	}
	defer output.Close()

	writer := bufio.NewWriter(output)
	for _, line := range result {
		fmt.Fprintln(writer, line)
	}

	return writer.Flush()
}

// processInput processes input stream
func processInput(reader io.Reader, opts *Options, filename string) error {
	scanner := bufio.NewScanner(reader)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		output, skip, err := processLine(line, lineNum, opts)
		if err != nil {
			return err
		}

		if !skip && !opts.Quiet {
			fmt.Println(output)
		}
	}

	return scanner.Err()
}

// processLines processes multiple lines
func processLines(lines []string, opts *Options) ([]string, error) {
	var result []string

	for i, line := range lines {
		output, skip, err := processLine(line, i+1, opts)
		if err != nil {
			return nil, err
		}

		if !skip {
			result = append(result, output)
		}
	}

	return result, nil
}

// processLine processes a single line
func processLine(line string, lineNum int, opts *Options) (string, bool, error) {
	expr := opts.Expression

	// Parse command
	cmd, err := parseCommand(expr, opts)
	if err != nil {
		return "", false, err
	}

	// Execute command
	return cmd.Execute(line, lineNum)
}

// Command interface
type SedCommand interface {
	Execute(line string, lineNum int) (string, bool, error)
}

// SubstituteCommand - s/pattern/replacement/flags
type SubstituteCommand struct {
	Pattern     *regexp.Regexp
	Replacement string
	Global      bool
}

func (s *SubstituteCommand) Execute(line string, lineNum int) (string, bool, error) {
	if s.Global {
		result := s.Pattern.ReplaceAllString(line, s.Replacement)
		return result, false, nil
	}
	// Replace only first occurrence
	result := line
	if loc := s.Pattern.FindStringIndex(line); loc != nil {
		result = line[:loc[0]] + s.Replacement + line[loc[1]:]
	}
	return result, false, nil
}

// DeleteCommand - /pattern/d or [line]d
type DeleteCommand struct {
	Pattern    *regexp.Regexp
	LineNumber int
}

func (d *DeleteCommand) Execute(line string, lineNum int) (string, bool, error) {
	if d.Pattern != nil {
		if d.Pattern.MatchString(line) {
			return "", true, nil // Skip line
		}
	} else if d.LineNumber > 0 {
		if lineNum == d.LineNumber {
			return "", true, nil // Skip line
		}
	}
	return line, false, nil
}

// PrintCommand - /pattern/p or [line]p
type PrintCommand struct {
	Pattern    *regexp.Regexp
	LineNumber int
}

func (p *PrintCommand) Execute(line string, lineNum int) (string, bool, error) {
	if p.Pattern != nil {
		if p.Pattern.MatchString(line) {
			return line, false, nil
		}
		return "", true, nil // Skip non-matching
	} else if p.LineNumber > 0 {
		if lineNum == p.LineNumber {
			return line, false, nil
		}
		return "", true, nil // Skip other lines
	}
	return line, false, nil
}

// parseCommand parses sed command expression
func parseCommand(expr string, opts *Options) (SedCommand, error) {
	expr = strings.TrimSpace(expr)

	// Substitute command: s/pattern/replacement/[flags]
	if strings.HasPrefix(expr, "s") {
		return parseSubstitute(expr, opts)
	}

	// Delete command: /pattern/d or [line]d
	if strings.HasSuffix(expr, "d") {
		return parseDelete(expr, opts)
	}

	// Print command: /pattern/p or [line]p
	if strings.HasSuffix(expr, "p") {
		return parsePrint(expr, opts)
	}

	return nil, fmt.Errorf("unsupported command: %s", expr)
}

// parseSubstitute parses s/pattern/replacement/flags
func parseSubstitute(expr string, opts *Options) (*SubstituteCommand, error) {
	// Remove 's' prefix
	expr = expr[1:]

	// Find delimiter
	if len(expr) == 0 {
		return nil, fmt.Errorf("invalid substitute command")
	}
	delim := expr[0]

	// Split by delimiter
	parts := strings.Split(expr[1:], string(delim))
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid substitute command")
	}

	pattern := parts[0]
	replacement := parts[1]
	flags := ""
	if len(parts) > 2 {
		flags = parts[2]
	}

	// Compile regex
	regexFlags := ""
	if opts.Extended {
		regexFlags = "(?m)"
	}
	re, err := regexp.Compile(regexFlags + pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}

	return &SubstituteCommand{
		Pattern:     re,
		Replacement: replacement,
		Global:      strings.Contains(flags, "g"),
	}, nil
}

// parseDelete parses /pattern/d or [line]d
func parseDelete(expr string, opts *Options) (*DeleteCommand, error) {
	expr = strings.TrimSuffix(expr, "d")
	expr = strings.TrimSpace(expr)

	// Line number delete: [num]d
	if num, err := strconv.Atoi(expr); err == nil {
		return &DeleteCommand{LineNumber: num}, nil
	}

	// Pattern delete: /pattern/d
	if strings.HasPrefix(expr, "/") && strings.HasSuffix(expr, "/") {
		pattern := expr[1 : len(expr)-1]
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %w", err)
		}
		return &DeleteCommand{Pattern: re}, nil
	}

	return nil, fmt.Errorf("invalid delete command: %s", expr)
}

// parsePrint parses /pattern/p or [line]p
func parsePrint(expr string, opts *Options) (*PrintCommand, error) {
	expr = strings.TrimSuffix(expr, "p")
	expr = strings.TrimSpace(expr)

	// Line number print: [num]p
	if num, err := strconv.Atoi(expr); err == nil {
		return &PrintCommand{LineNumber: num}, nil
	}

	// Pattern print: /pattern/p
	if strings.HasPrefix(expr, "/") && strings.HasSuffix(expr, "/") {
		pattern := expr[1 : len(expr)-1]
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %w", err)
		}
		return &PrintCommand{Pattern: re}, nil
	}

	return nil, fmt.Errorf("invalid print command: %s", expr)
}
