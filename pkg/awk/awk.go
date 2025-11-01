package awk

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

// Options holds awk configuration
type Options struct {
	FieldSeparator string
	Program        string
}

// Context holds awk execution context
type Context struct {
	NR     int      // Number of records (lines)
	NF     int      // Number of fields
	Fields []string // Current line fields
	Line   string   // Current line
	FS     string   // Field separator
}

// Command returns the awk command
func Command() *cobra.Command {
	opts := &Options{
		FieldSeparator: " ",
	}

	cmd := &cobra.Command{
		Use:   "awk [options] 'program' [file...]",
		Short: "Pattern scanning and text processing",
		Long: `Pattern scanning and text processing language.
Simplified awk implementation with common features.

Program Syntax:
  pattern { action }       Execute action when pattern matches
  BEGIN { action }         Execute before processing input
  END { action }           Execute after processing input
  { action }               Execute for every line

Special Variables:
  $0     Whole line
  $1,$2  Field 1, field 2, etc.
  NR     Current line number
  NF     Number of fields
  FS     Field separator

Examples:
  awk '{print $1}'                Print first field
  awk '{print $1, $3}'            Print fields 1 and 3
  awk '/pattern/ {print $0}'      Print lines matching pattern
  awk 'NR==5 {print}'             Print line 5
  awk '{sum+=$1} END {print sum}' Sum first field`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Program = args[0]
			files := args[1:]

			if len(files) == 0 {
				return processInput(os.Stdin, opts)
			}

			for _, file := range files {
				if err := processFile(file, opts); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.FieldSeparator, "field-separator", "F", " ", "Field separator")

	return cmd
}

// processFile processes a file
func processFile(filename string, opts *Options) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open '%s': %w", filename, err)
	}
	defer file.Close()

	return processInput(file, opts)
}

// processInput processes input stream
func processInput(reader io.Reader, opts *Options) error {
	program, err := parseProgram(opts.Program)
	if err != nil {
		return err
	}

	ctx := &Context{
		FS: opts.FieldSeparator,
	}

	// Execute BEGIN
	if program.Begin != nil {
		if err := program.Begin.Execute(ctx); err != nil {
			return err
		}
	}

	// Process lines
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		ctx.NR++
		ctx.Line = scanner.Text()
		ctx.Fields = splitFields(ctx.Line, ctx.FS)
		ctx.NF = len(ctx.Fields)

		// Execute rules
		for _, rule := range program.Rules {
			if rule.Pattern == nil || rule.Pattern.Match(ctx) {
				if err := rule.Action.Execute(ctx); err != nil {
					return err
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Execute END
	if program.End != nil {
		if err := program.End.Execute(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Program represents awk program
type Program struct {
	Begin *Action
	Rules []*Rule
	End   *Action
}

// Rule represents pattern-action rule
type Rule struct {
	Pattern Pattern
	Action  *Action
}

// Pattern interface
type Pattern interface {
	Match(ctx *Context) bool
}

// AlwaysPattern always matches
type AlwaysPattern struct{}

func (p *AlwaysPattern) Match(ctx *Context) bool {
	return true
}

// RegexPattern matches regex
type RegexPattern struct {
	Regex *regexp.Regexp
}

func (p *RegexPattern) Match(ctx *Context) bool {
	return p.Regex.MatchString(ctx.Line)
}

// LinePattern matches specific line
type LinePattern struct {
	LineNumber int
}

func (p *LinePattern) Match(ctx *Context) bool {
	return ctx.NR == p.LineNumber
}

// Action represents action to execute
type Action struct {
	Statements []Statement
	Variables  map[string]float64
}

// Execute executes action
func (a *Action) Execute(ctx *Context) error {
	if a.Variables == nil {
		a.Variables = make(map[string]float64)
	}

	for _, stmt := range a.Statements {
		if err := stmt.Execute(ctx, a.Variables); err != nil {
			return err
		}
	}
	return nil
}

// Statement interface
type Statement interface {
	Execute(ctx *Context, vars map[string]float64) error
}

// PrintStatement prints fields
type PrintStatement struct {
	Fields []FieldRef
}

func (s *PrintStatement) Execute(ctx *Context, vars map[string]float64) error {
	if len(s.Fields) == 0 {
		fmt.Println(ctx.Line)
		return nil
	}

	parts := make([]string, len(s.Fields))
	for i, field := range s.Fields {
		parts[i] = field.GetValue(ctx, vars)
	}
	fmt.Println(strings.Join(parts, " "))
	return nil
}

// AssignStatement assigns value to variable
type AssignStatement struct {
	Variable string
	Expr     Expression
}

func (s *AssignStatement) Execute(ctx *Context, vars map[string]float64) error {
	value := s.Expr.Evaluate(ctx, vars)
	vars[s.Variable] = value
	return nil
}

// Expression interface
type Expression interface {
	Evaluate(ctx *Context, vars map[string]float64) float64
}

// FieldExpression evaluates field value
type FieldExpression struct {
	FieldNum int
}

func (e *FieldExpression) Evaluate(ctx *Context, vars map[string]float64) float64 {
	if e.FieldNum == 0 {
		return 0
	}
	if e.FieldNum > 0 && e.FieldNum <= len(ctx.Fields) {
		val, _ := strconv.ParseFloat(ctx.Fields[e.FieldNum-1], 64)
		return val
	}
	return 0
}

// VariableExpression evaluates variable
type VariableExpression struct {
	Name string
}

func (e *VariableExpression) Evaluate(ctx *Context, vars map[string]float64) float64 {
	return vars[e.Name]
}

// BinaryExpression evaluates binary operation
type BinaryExpression struct {
	Left  Expression
	Op    string
	Right Expression
}

func (e *BinaryExpression) Evaluate(ctx *Context, vars map[string]float64) float64 {
	left := e.Left.Evaluate(ctx, vars)
	right := e.Right.Evaluate(ctx, vars)

	switch e.Op {
	case "+":
		return left + right
	case "-":
		return left - right
	case "*":
		return left * right
	case "/":
		if right != 0 {
			return left / right
		}
		return 0
	default:
		return 0
	}
}

// FieldRef references a field
type FieldRef struct {
	Field int
	Var   string
}

func (f *FieldRef) GetValue(ctx *Context, vars map[string]float64) string {
	if f.Var != "" {
		return fmt.Sprintf("%v", vars[f.Var])
	}
	if f.Field == 0 {
		return ctx.Line
	}
	if f.Field > 0 && f.Field <= len(ctx.Fields) {
		return ctx.Fields[f.Field-1]
	}
	return ""
}

// parseProgram parses awk program
func parseProgram(prog string) (*Program, error) {
	prog = strings.TrimSpace(prog)

	program := &Program{
		Rules: make([]*Rule, 0),
	}

	// Parse BEGIN
	if strings.HasPrefix(prog, "BEGIN") {
		endIdx := strings.Index(prog, "}")
		if endIdx == -1 {
			return nil, fmt.Errorf("missing closing brace for BEGIN")
		}
		actionStr := prog[5 : endIdx+1]
		action, err := parseAction(actionStr)
		if err != nil {
			return nil, err
		}
		program.Begin = action
		prog = strings.TrimSpace(prog[endIdx+1:])
	}

	// Parse END
	if idx := strings.Index(prog, "END"); idx >= 0 {
		actionStr := prog[idx+3:]
		action, err := parseAction(actionStr)
		if err != nil {
			return nil, err
		}
		program.End = action
		prog = strings.TrimSpace(prog[:idx])
	}

	// Parse rules
	if prog != "" {
		rule, err := parseRule(prog)
		if err != nil {
			return nil, err
		}
		program.Rules = append(program.Rules, rule)
	}

	return program, nil
}

// parseRule parses pattern-action rule
func parseRule(ruleStr string) (*Rule, error) {
	ruleStr = strings.TrimSpace(ruleStr)

	rule := &Rule{}

	// Parse pattern
	if strings.HasPrefix(ruleStr, "/") {
		endIdx := strings.Index(ruleStr[1:], "/")
		if endIdx == -1 {
			return nil, fmt.Errorf("missing closing / for pattern")
		}
		patternStr := ruleStr[1 : endIdx+1]
		regex, err := regexp.Compile(patternStr)
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %w", err)
		}
		rule.Pattern = &RegexPattern{Regex: regex}
		ruleStr = strings.TrimSpace(ruleStr[endIdx+2:])
	} else if strings.HasPrefix(ruleStr, "NR==") {
		// Line number pattern
		parts := strings.SplitN(ruleStr, " ", 2)
		lineStr := strings.TrimPrefix(parts[0], "NR==")
		lineNum, err := strconv.Atoi(lineStr)
		if err != nil {
			return nil, fmt.Errorf("invalid line number: %s", lineStr)
		}
		rule.Pattern = &LinePattern{LineNumber: lineNum}
		if len(parts) > 1 {
			ruleStr = strings.TrimSpace(parts[1])
		} else {
			ruleStr = ""
		}
	} else {
		rule.Pattern = &AlwaysPattern{}
	}

	// Parse action
	if ruleStr != "" {
		action, err := parseAction(ruleStr)
		if err != nil {
			return nil, err
		}
		rule.Action = action
	} else {
		// Default action: print
		rule.Action = &Action{
			Statements: []Statement{&PrintStatement{}},
		}
	}

	return rule, nil
}

// parseAction parses action block
func parseAction(actionStr string) (*Action, error) {
	actionStr = strings.TrimSpace(actionStr)
	if !strings.HasPrefix(actionStr, "{") || !strings.HasSuffix(actionStr, "}") {
		return nil, fmt.Errorf("action must be enclosed in braces")
	}

	actionStr = actionStr[1 : len(actionStr)-1]
	actionStr = strings.TrimSpace(actionStr)

	action := &Action{
		Statements: make([]Statement, 0),
	}

	// Parse statements
	if actionStr != "" {
		stmt, err := parseStatement(actionStr)
		if err != nil {
			return nil, err
		}
		action.Statements = append(action.Statements, stmt)
	}

	return action, nil
}

// parseStatement parses statement
func parseStatement(stmtStr string) (Statement, error) {
	stmtStr = strings.TrimSpace(stmtStr)

	// Print statement
	if strings.HasPrefix(stmtStr, "print") {
		return parsePrint(stmtStr)
	}

	// Assignment: var+=expr or var=expr
	if strings.Contains(stmtStr, "+=") {
		parts := strings.SplitN(stmtStr, "+=", 2)
		varName := strings.TrimSpace(parts[0])
		exprStr := strings.TrimSpace(parts[1])
		expr, err := parseExpression(exprStr)
		if err != nil {
			return nil, err
		}
		return &AssignStatement{
			Variable: varName,
			Expr: &BinaryExpression{
				Left:  &VariableExpression{Name: varName},
				Op:    "+",
				Right: expr,
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported statement: %s", stmtStr)
}

// parsePrint parses print statement
func parsePrint(printStr string) (*PrintStatement, error) {
	printStr = strings.TrimPrefix(printStr, "print")
	printStr = strings.TrimSpace(printStr)

	stmt := &PrintStatement{
		Fields: make([]FieldRef, 0),
	}

	if printStr == "" || printStr == "$0" {
		return stmt, nil
	}

	// Parse field list
	fields := strings.Split(printStr, ",")
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if strings.HasPrefix(field, "$") {
			fieldNum, err := strconv.Atoi(field[1:])
			if err != nil {
				return nil, fmt.Errorf("invalid field: %s", field)
			}
			stmt.Fields = append(stmt.Fields, FieldRef{Field: fieldNum})
		} else {
			stmt.Fields = append(stmt.Fields, FieldRef{Var: field})
		}
	}

	return stmt, nil
}

// parseExpression parses expression
func parseExpression(exprStr string) (Expression, error) {
	exprStr = strings.TrimSpace(exprStr)

	// Field reference
	if strings.HasPrefix(exprStr, "$") {
		fieldNum, err := strconv.Atoi(exprStr[1:])
		if err != nil {
			return nil, fmt.Errorf("invalid field: %s", exprStr)
		}
		return &FieldExpression{FieldNum: fieldNum}, nil
	}

	// Variable
	return &VariableExpression{Name: exprStr}, nil
}

// splitFields splits line into fields
func splitFields(line, sep string) []string {
	if sep == " " {
		return strings.Fields(line)
	}
	return strings.Split(line, sep)
}
