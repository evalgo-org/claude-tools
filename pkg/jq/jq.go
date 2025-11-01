package jq

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// Options holds jq configuration
type Options struct {
	Compact     bool
	RawOutput   bool
	SortKeys    bool
	TabIndent   bool
	ColorOutput bool
	NullInput   bool
	SlurpMode   bool
}

// Command returns the jq command
func Command() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "jq [filter] [file...]",
		Short: "Process JSON data with filters",
		Long: `Process JSON data using a simple filter syntax.
Supports basic JSON querying, filtering, and transformation.

Filter Syntax:
  .              Identity (passthrough)
  .key           Get object key
  .[0]           Get array element
  .[]            Array/object values iterator
  .key1.key2     Nested access
  keys           Get object keys
  length         Get array/object/string length
  type           Get value type`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := args[0]
			files := args[1:]

			if len(files) == 0 || opts.NullInput {
				return processInput(os.Stdin, filter, opts)
			}

			for _, file := range files {
				if err := processFile(file, filter, opts); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.Compact, "compact", "c", false, "Compact output")
	cmd.Flags().BoolVarP(&opts.RawOutput, "raw-output", "r", false, "Output raw strings")
	cmd.Flags().BoolVarP(&opts.SortKeys, "sort-keys", "S", false, "Sort object keys")
	cmd.Flags().BoolVar(&opts.TabIndent, "tab", false, "Use tabs for indentation")
	cmd.Flags().BoolVarP(&opts.ColorOutput, "color-output", "C", false, "Colorize output")
	cmd.Flags().BoolVarP(&opts.NullInput, "null-input", "n", false, "Don't read input")
	cmd.Flags().BoolVarP(&opts.SlurpMode, "slurp", "s", false, "Read entire input into array")

	return cmd
}

// processFile processes a JSON file
func processFile(filename string, filter string, opts *Options) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open '%s': %w", filename, err)
	}
	defer file.Close()

	return processInput(file, filter, opts)
}

// processInput processes JSON from input
func processInput(reader io.Reader, filter string, opts *Options) error {
	if opts.SlurpMode {
		return processSlurp(reader, filter, opts)
	}

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 1MB max

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var data interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		result, err := applyFilter(data, filter)
		if err != nil {
			return err
		}

		if err := outputResult(result, opts); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// processSlurp reads all JSON into array
func processSlurp(reader io.Reader, filter string, opts *Options) error {
	var items []interface{}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var data interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		items = append(items, data)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	result, err := applyFilter(items, filter)
	if err != nil {
		return err
	}

	return outputResult(result, opts)
}

// applyFilter applies a filter to JSON data
func applyFilter(data interface{}, filter string) (interface{}, error) {
	filter = strings.TrimSpace(filter)

	// Identity filter
	if filter == "." {
		return data, nil
	}

	// Built-in functions
	switch filter {
	case "keys":
		return getKeys(data)
	case "length":
		return getLength(data)
	case "type":
		return getType(data), nil
	}

	// Array iterator
	if filter == ".[]" {
		return iterateArray(data)
	}

	// Path-based access
	if strings.HasPrefix(filter, ".") {
		return accessPath(data, filter[1:])
	}

	return nil, fmt.Errorf("unsupported filter: %s", filter)
}

// accessPath accesses nested JSON path
func accessPath(data interface{}, path string) (interface{}, error) {
	if path == "" {
		return data, nil
	}

	parts := parsePath(path)
	current := data

	for _, part := range parts {
		var err error
		current, err = accessPart(current, part)
		if err != nil {
			return nil, err
		}
	}

	return current, nil
}

// parsePath parses dot-separated path
func parsePath(path string) []string {
	var parts []string
	var current strings.Builder
	inBracket := false

	for i := 0; i < len(path); i++ {
		ch := path[i]

		if ch == '[' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			inBracket = true
			current.WriteByte(ch)
		} else if ch == ']' {
			current.WriteByte(ch)
			parts = append(parts, current.String())
			current.Reset()
			inBracket = false
		} else if ch == '.' && !inBracket {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// accessPart accesses single path part
func accessPart(data interface{}, part string) (interface{}, error) {
	// Array index access: [0], [1], etc.
	if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
		indexStr := part[1 : len(part)-1]
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			return nil, fmt.Errorf("invalid array index: %s", part)
		}

		arr, ok := data.([]interface{})
		if !ok {
			return nil, fmt.Errorf("not an array")
		}

		if index < 0 || index >= len(arr) {
			return nil, fmt.Errorf("index out of bounds: %d", index)
		}

		return arr[index], nil
	}

	// Object key access
	obj, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("not an object")
	}

	val, exists := obj[part]
	if !exists {
		return nil, nil
	}

	return val, nil
}

// iterateArray returns array/object values
func iterateArray(data interface{}) (interface{}, error) {
	switch v := data.(type) {
	case []interface{}:
		return v, nil
	case map[string]interface{}:
		values := make([]interface{}, 0, len(v))
		for _, val := range v {
			values = append(values, val)
		}
		return values, nil
	default:
		return nil, fmt.Errorf("cannot iterate over %T", data)
	}
}

// getKeys returns object keys
func getKeys(data interface{}) (interface{}, error) {
	obj, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("keys only works on objects")
	}

	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}

	return keys, nil
}

// getLength returns length of array/object/string
func getLength(data interface{}) (interface{}, error) {
	switch v := data.(type) {
	case []interface{}:
		return len(v), nil
	case map[string]interface{}:
		return len(v), nil
	case string:
		return len(v), nil
	default:
		return nil, fmt.Errorf("length not supported for %T", data)
	}
}

// getType returns JSON type
func getType(data interface{}) string {
	switch data.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case float64:
		return "number"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

// outputResult outputs filtered result
func outputResult(result interface{}, opts *Options) error {
	// Handle array iterator results
	if arr, ok := result.([]interface{}); ok && !opts.SlurpMode {
		for _, item := range arr {
			if err := outputSingle(item, opts); err != nil {
				return err
			}
		}
		return nil
	}

	return outputSingle(result, opts)
}

// outputSingle outputs single result
func outputSingle(result interface{}, opts *Options) error {
	// Raw output for strings
	if opts.RawOutput {
		if str, ok := result.(string); ok {
			fmt.Println(str)
			return nil
		}
	}

	// Handle nil
	if result == nil {
		fmt.Println("null")
		return nil
	}

	// JSON output
	var output []byte
	var err error

	if opts.Compact {
		output, err = json.Marshal(result)
	} else if opts.TabIndent {
		output, err = json.MarshalIndent(result, "", "\t")
	} else {
		output, err = json.MarshalIndent(result, "", "  ")
	}

	if err != nil {
		return fmt.Errorf("cannot encode JSON: %w", err)
	}

	fmt.Println(string(output))
	return nil
}
