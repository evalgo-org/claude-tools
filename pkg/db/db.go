package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

// DBConfig represents database configuration from .claude-project.json
type DBConfig struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
	Location string `json:"location"`
}

// ClaudeProject represents .claude-project.json structure
type ClaudeProject struct {
	Database DBConfig `json:"database"`
}

// LoadConfig loads database configuration from .claude-project.json
func LoadConfig() (*DBConfig, error) {
	// Look for .claude-project.json in current directory or parents
	configPath, err := findClaudeProjectFile()
	if err != nil {
		return nil, fmt.Errorf("failed to find .claude-project.json: %w", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var project ClaudeProject
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &project.Database, nil
}

// findClaudeProjectFile searches for .claude-project.json in current and parent directories
func findClaudeProjectFile() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		configPath := filepath.Join(dir, ".claude-project.json")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf(".claude-project.json not found in current directory or parents")
}

// Connect establishes a database connection
func Connect(config *DBConfig) (*sql.DB, error) {
	// Use defaults if not specified
	user := config.User
	if user == "" {
		user = "claude"
	}

	password := config.Password
	if password == "" {
		password = "claude_dev_password"
	}

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		config.Host, config.Port, config.Name, user, password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// Query executes a SQL query and returns results
func Query(db *sql.DB, query string, format string) error {
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	switch format {
	case "json":
		return printJSON(rows, columns)
	case "csv":
		return printCSV(rows, columns)
	default:
		return printTable(rows, columns)
	}
}

// printTable prints results in table format
func printTable(rows *sql.Rows, columns []string) error {
	// Print header
	fmt.Println(strings.Join(columns, " | "))
	fmt.Println(strings.Repeat("-", len(columns)*20))

	// Print rows
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}

		row := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				row[i] = "NULL"
			} else {
				row[i] = fmt.Sprintf("%v", val)
			}
		}
		fmt.Println(strings.Join(row, " | "))
	}

	return rows.Err()
}

// printJSON prints results in JSON format
func printJSON(rows *sql.Rows, columns []string) error {
	results := []map[string]interface{}{}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

// printCSV prints results in CSV format
func printCSV(rows *sql.Rows, columns []string) error {
	// Print header
	fmt.Println(strings.Join(columns, ","))

	// Print rows
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}

		row := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				row[i] = ""
			} else {
				row[i] = fmt.Sprintf("%v", val)
			}
		}
		fmt.Println(strings.Join(row, ","))
	}

	return rows.Err()
}

// ListTables lists all tables in the database
func ListTables(db *sql.DB) error {
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		ORDER BY table_name;
	`
	return Query(db, query, "table")
}

// GetRules retrieves rules by category
func GetRules(db *sql.DB, category string) error {
	query := fmt.Sprintf(`
		SELECT rule_id, title, category, priority
		FROM rules
		WHERE category = '%s'
		ORDER BY priority DESC, rule_id;
	`, category)
	return Query(db, query, "table")
}

// GetConfigs retrieves CI configs by type
func GetConfigs(db *sql.DB, configType string) error {
	query := fmt.Sprintf(`
		SELECT config_name, config_type, notes
		FROM ci_config
		WHERE config_type = '%s'
		ORDER BY config_name;
	`, configType)
	return Query(db, query, "table")
}

// ListProjects lists all tracked projects
func ListProjects(db *sql.DB) error {
	query := `
		SELECT project_id, project_name, project_type, project_path
		FROM project_metadata
		ORDER BY project_id;
	`
	return Query(db, query, "table")
}

// Command returns the db command for claude-tools
func Command() *cobra.Command {
	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "Query claude-memory database",
		Long: `Query the claude-memory TimescaleDB database.

Reads database configuration from .claude-project.json in current or parent directories.

Examples:
  claude-tools db query "SELECT * FROM rules"
  claude-tools db tables
  claude-tools db rules --category metarules
  claude-tools db configs --type nixpacks
  claude-tools db projects`,
	}

	// Query subcommand
	queryCmd := &cobra.Command{
		Use:   "query <sql>",
		Short: "Execute a SQL query",
		Long: `Execute a custom SQL query against the database.

Examples:
  claude-tools db query "SELECT * FROM rules WHERE priority > 3"
  claude-tools db query "SELECT config_name FROM ci_config" --format json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			conn, err := Connect(config)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer conn.Close()

			format, _ := cmd.Flags().GetString("format")
			return Query(conn, args[0], format)
		},
	}
	queryCmd.Flags().StringP("format", "f", "table", "Output format (table, json, csv)")

	// Tables subcommand
	tablesCmd := &cobra.Command{
		Use:   "tables",
		Short: "List all tables in the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			conn, err := Connect(config)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer conn.Close()

			return ListTables(conn)
		},
	}

	// Rules subcommand
	rulesCmd := &cobra.Command{
		Use:   "rules",
		Short: "List rules by category",
		Long: `List development rules by category.

Categories: metarules, best-practices, workflows, error-handling, tools-usage, profiles

Examples:
  claude-tools db rules
  claude-tools db rules --category best-practices
  claude-tools db rules -c workflows`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			conn, err := Connect(config)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer conn.Close()

			category, _ := cmd.Flags().GetString("category")
			return GetRules(conn, category)
		},
	}
	rulesCmd.Flags().StringP("category", "c", "metarules", "Rule category to query")

	// Configs subcommand
	configsCmd := &cobra.Command{
		Use:   "configs",
		Short: "List CI/CD configurations by type",
		Long: `List CI/CD configurations by type.

Types: github-actions, golangci-lint, nixpacks, pre-commit, project-template

Examples:
  claude-tools db configs
  claude-tools db configs --type nixpacks
  claude-tools db configs -t pre-commit`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			conn, err := Connect(config)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer conn.Close()

			configType, _ := cmd.Flags().GetString("type")
			return GetConfigs(conn, configType)
		},
	}
	configsCmd.Flags().StringP("type", "t", "github-actions", "Config type to query")

	// Projects subcommand
	projectsCmd := &cobra.Command{
		Use:   "projects",
		Short: "List all tracked projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			conn, err := Connect(config)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer conn.Close()

			return ListProjects(conn)
		},
	}

	dbCmd.AddCommand(queryCmd)
	dbCmd.AddCommand(tablesCmd)
	dbCmd.AddCommand(rulesCmd)
	dbCmd.AddCommand(configsCmd)
	dbCmd.AddCommand(projectsCmd)

	return dbCmd
}
