# Claude Tools

[![Tests and Build](https://github.com/evalgo-org/claude-tools/actions/workflows/tests.yml/badge.svg)](https://github.com/evalgo-org/claude-tools/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/evalgo-org/claude-tools)](https://goreportcard.com/report/github.com/evalgo-org/claude-tools)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Cross-platform CLI tools for development, built in Go for consistent behavior across Windows, Linux, and macOS.

## Overview

Claude Tools provides Go implementations of common Linux/Unix command-line utilities, ensuring consistent behavior across all operating systems without external dependencies. Built for teams that need reliable, cross-platform development tools.

## Features

- **Cross-Platform**: Works identically on Windows, Linux, and macOS
- **No Dependencies**: Single binary with no external runtime requirements
- **Go Performance**: Fast execution with low memory footprint
- **Familiar Interface**: Compatible with common Unix tool flags and options
- **Integrated Logging**: Uses EVE library for consistent logging across tools

## Installation

### From Source

```bash
git clone https://github.com/evalgo-org/claude-tools.git
cd claude-tools
go build -o claude-tools cmd/claude-tools/main.go
```

### Using Go Install

```bash
go install github.com/evalgo-org/claude-tools/cmd/claude-tools@latest
```

### Manual Installation

1. Download the binary for your platform from [Releases](https://github.com/evalgo-org/claude-tools/releases)
2. Add it to your PATH
3. Make it executable (Unix-like systems): `chmod +x claude-tools`

## Available Tools

### grep - Pattern Searching

Search for patterns in files using regular expressions.

```bash
# Basic search
claude-tools grep "pattern" file.txt

# Recursive search with line numbers
claude-tools grep -r -n "TODO" .

# Case-insensitive search
claude-tools grep -i "error" *.log

# Show only filenames
claude-tools grep -l "import" *.go

# Count matches
claude-tools grep -c "func" main.go
```

**Flags:**
- `-i, --ignore-case`: Case-insensitive search
- `-r, --recursive`: Search directories recursively
- `-n, --line-number`: Show line numbers
- `-A NUM`: Show NUM lines after match
- `-B NUM`: Show NUM lines before match
- `-C NUM`: Show NUM lines before and after match
- `-v, --invert-match`: Show non-matching lines
- `-l, --files-with-matches`: Show only filenames
- `-c, --count`: Show count of matches

### find - File Finding

Find files and directories by name, type, or other criteria.

```bash
# Find all Go files
claude-tools find . --name "*.go" --type f

# Find directories
claude-tools find /path --type d

# Case-insensitive name search
claude-tools find . --iname "readme*"

# Limit search depth
claude-tools find . --name "*.go" --maxdepth 2
```

**Flags:**
- `-n, --name`: Find by name pattern (case-sensitive)
- `--iname`: Find by name pattern (case-insensitive)
- `-t, --type`: Filter by type (f=file, d=directory, l=symlink)
- `--maxdepth`: Maximum depth to search
- `--mindepth`: Minimum depth to search

### cat - File Display

Concatenate and display file contents.

```bash
# Display file
claude-tools cat file.txt

# Display with line numbers
claude-tools cat -n file.go

# Show non-printing characters
claude-tools cat -A file.txt

# Squeeze blank lines
claude-tools cat -s file.txt
```

**Flags:**
- `-n, --number`: Number all output lines
- `-A, --show-all`: Show non-printing characters
- `-s, --squeeze-blank`: Squeeze multiple blank lines

## Usage Examples

### Code Analysis

```bash
# Find all TODO comments with line numbers
claude-tools grep -r -n "TODO" . | head -20

# Count Go files in project
claude-tools find . --name "*.go" --type f | wc -l

# Find files using specific function
claude-tools grep -r "Logger.Fatal" --files-with-matches

# Display file with line numbers and search
claude-tools cat -n main.go | claude-tools grep "func"
```

### Project Exploration

```bash
# Find all test files
claude-tools find . --name "*_test.go"

# Search for error handling
claude-tools grep -r "if err != nil" --count

# List all directories at depth 2
claude-tools find . --type d --maxdepth 2
```

## Architecture

### Project Structure

```
claude-tools/
├── cmd/
│   └── claude-tools/
│       └── main.go          # CLI entry point
├── pkg/
│   ├── grep/
│   │   └── grep.go         # grep implementation
│   ├── find/
│   │   └── find.go         # find implementation
│   └── cat/
│       └── cat.go          # cat implementation
├── .github/
│   └── workflows/
│       └── tests.yml       # CI/CD pipeline
└── go.mod                  # Dependencies
```

### Dependencies

- [cobra](https://github.com/spf13/cobra) v1.10.1 - CLI framework
- [eve.evalgo.org](https://eve.evalgo.org) v0.0.13 - Logging and utilities

### Design Principles

1. **No Panic Rule**: All functions return errors instead of panicking
2. **Library First**: Leverage EVE library for common functionality
3. **Cross-Platform**: Test on Linux, macOS, and Windows
4. **Memory First**: Store configuration and metadata in database
5. **Standard Compliance**: Follow Unix tool conventions where applicable

## Development

### Prerequisites

- Go 1.24 or later
- Make (optional, for using Makefile)

### Building

```bash
# Build for current platform
go build -o claude-tools cmd/claude-tools/main.go

# Build for all platforms
GOOS=linux GOARCH=amd64 go build -o claude-tools-linux-amd64 cmd/claude-tools/main.go
GOOS=darwin GOARCH=amd64 go build -o claude-tools-darwin-amd64 cmd/claude-tools/main.go
GOOS=windows GOARCH=amd64 go build -o claude-tools-windows-amd64.exe cmd/claude-tools/main.go
```

### Testing

```bash
# Run all tests
go test -v ./...

# Run with race detection
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Run with auto-fix
golangci-lint run --fix
```

## Roadmap

### Phase 1: Core Tools (v0.1.0) ✅
- [x] grep - Pattern searching
- [x] find - File finding
- [x] cat - File display

### Phase 2: File Utilities (v0.2.0)
- [ ] ls - Directory listing
- [ ] head - Display first lines
- [ ] tail - Display last lines
- [ ] wc - Word/line counting

### Phase 3: Advanced Tools (v0.3.0)
- [ ] tree - Directory tree display
- [ ] jq - JSON processing
- [ ] sed - Stream editing
- [ ] awk - Text processing

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Standards

- Follow Go best practices and idioms
- Add tests for new functionality
- Update documentation as needed
- Run `golangci-lint` before submitting
- Follow the no-panic rule: return errors instead

## License

MIT License - see [LICENSE](LICENSE) file for details

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) CLI framework
- Uses [EVE](https://eve.evalgo.org) library for logging and utilities
- Inspired by Unix/Linux command-line tools

## Support

- **Issues**: [GitHub Issues](https://github.com/evalgo-org/claude-tools/issues)
- **Documentation**: [Wiki](https://github.com/evalgo-org/claude-tools/wiki)
- **Discussions**: [GitHub Discussions](https://github.com/evalgo-org/claude-tools/discussions)

---

**Version**: 0.1.0
**Status**: Production Ready
**Maintained by**: evalgo.org
