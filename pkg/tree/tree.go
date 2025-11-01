package tree

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// Options holds tree configuration
type Options struct {
	AllFiles      bool // Show hidden files
	DirsOnly      bool // List directories only
	FullPath      bool // Print full path
	Level         int  // Max depth
	Pattern       string
	IgnorePattern string
	FileLimit     int
	SortByTime    bool
	SortReverse   bool
	NoIndent      bool
	ShowSize      bool
	ShowPerms     bool
}

// Stats holds tree statistics
type Stats struct {
	Dirs  int
	Files int
}

// Command returns the tree command
func Command() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "tree [directory]",
		Short: "Display directory tree structure",
		Long: `Display directory contents in a tree-like format.
Shows files and directories in a hierarchical view.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			return treeDir(dir, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.AllFiles, "all", "a", false, "Show hidden files")
	cmd.Flags().BoolVarP(&opts.DirsOnly, "dirs-only", "d", false, "List directories only")
	cmd.Flags().BoolVarP(&opts.FullPath, "full-path", "f", false, "Print full path")
	cmd.Flags().IntVarP(&opts.Level, "level", "L", -1, "Max depth (-1 = unlimited)")
	cmd.Flags().StringVarP(&opts.Pattern, "pattern", "P", "", "Only show files matching pattern")
	cmd.Flags().StringVarP(&opts.IgnorePattern, "ignore", "I", "", "Ignore files matching pattern")
	cmd.Flags().IntVar(&opts.FileLimit, "filelimit", -1, "Max files to display (-1 = unlimited)")
	cmd.Flags().BoolVar(&opts.SortByTime, "timesort", false, "Sort by modification time")
	cmd.Flags().BoolVarP(&opts.SortReverse, "reverse", "r", false, "Reverse sort order")
	cmd.Flags().BoolVar(&opts.NoIndent, "noreport", false, "Don't print summary report")
	cmd.Flags().BoolVarP(&opts.ShowSize, "size", "s", false, "Show file sizes")
	cmd.Flags().BoolVarP(&opts.ShowPerms, "perms", "p", false, "Show file permissions")

	return cmd
}

// treeDir displays directory tree
func treeDir(root string, opts *Options) error {
	// Verify directory exists
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("cannot access '%s': %w", root, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("'%s' is not a directory", root)
	}

	stats := &Stats{}
	fileCount := 0

	// Print root
	fmt.Println(root)

	// Walk directory tree
	err = walkTree(root, "", true, 0, opts, stats, &fileCount)
	if err != nil {
		return err
	}

	// Print summary
	if !opts.NoIndent {
		fmt.Printf("\n%d directories", stats.Dirs)
		if !opts.DirsOnly {
			fmt.Printf(", %d files", stats.Files)
		}
		fmt.Println()
	}

	return nil
}

// walkTree recursively walks directory tree
func walkTree(path string, prefix string, isLast bool, depth int, opts *Options, stats *Stats, fileCount *int) error {
	// Check depth limit
	if opts.Level >= 0 && depth > opts.Level {
		return nil
	}

	// Check file limit
	if opts.FileLimit > 0 && *fileCount >= opts.FileLimit {
		return nil
	}

	// Read directory entries
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// Filter entries
	filtered := filterEntries(entries, opts)

	// Sort entries
	sortEntries(filtered, opts)

	// Process each entry
	for i, entry := range filtered {
		if opts.FileLimit > 0 && *fileCount >= opts.FileLimit {
			break
		}

		isLastEntry := i == len(filtered)-1
		name := entry.Name()
		fullPath := filepath.Join(path, name)

		// Build display name
		displayName := name
		if opts.FullPath {
			displayName = fullPath
		}

		// Get entry info for size/perms
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Add size if requested
		if opts.ShowSize && !entry.IsDir() {
			displayName = fmt.Sprintf("%s (%s)", displayName, formatSize(info.Size()))
		}

		// Add permissions if requested
		if opts.ShowPerms {
			displayName = fmt.Sprintf("[%s] %s", info.Mode().String(), displayName)
		}

		// Print entry
		connector := "├── "
		if isLastEntry {
			connector = "└── "
		}

		if entry.IsDir() {
			displayName += "/"
		}

		fmt.Printf("%s%s%s\n", prefix, connector, displayName)

		// Update stats
		if entry.IsDir() {
			stats.Dirs++
		} else {
			stats.Files++
			*fileCount++
		}

		// Recurse into directories
		if entry.IsDir() {
			newPrefix := prefix
			if isLastEntry {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			err = walkTree(fullPath, newPrefix, isLastEntry, depth+1, opts, stats, fileCount)
			if err != nil {
				// Continue on error
				continue
			}
		}
	}

	return nil
}

// filterEntries filters directory entries based on options
func filterEntries(entries []fs.DirEntry, opts *Options) []fs.DirEntry {
	filtered := make([]fs.DirEntry, 0, len(entries))

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless -a
		if !opts.AllFiles && strings.HasPrefix(name, ".") {
			continue
		}

		// Skip files if dirs-only
		if opts.DirsOnly && !entry.IsDir() {
			continue
		}

		// Match pattern
		if opts.Pattern != "" {
			matched, _ := filepath.Match(opts.Pattern, name)
			if !matched {
				continue
			}
		}

		// Ignore pattern
		if opts.IgnorePattern != "" {
			matched, _ := filepath.Match(opts.IgnorePattern, name)
			if matched {
				continue
			}
		}

		filtered = append(filtered, entry)
	}

	return filtered
}

// sortEntries sorts directory entries
func sortEntries(entries []fs.DirEntry, opts *Options) {
	if opts.SortByTime {
		sort.Slice(entries, func(i, j int) bool {
			infoI, errI := entries[i].Info()
			infoJ, errJ := entries[j].Info()
			if errI != nil || errJ != nil {
				return entries[i].Name() < entries[j].Name()
			}
			if opts.SortReverse {
				return infoI.ModTime().After(infoJ.ModTime())
			}
			return infoI.ModTime().Before(infoJ.ModTime())
		})
	} else {
		sort.Slice(entries, func(i, j int) bool {
			if opts.SortReverse {
				return entries[i].Name() > entries[j].Name()
			}
			return entries[i].Name() < entries[j].Name()
		})
	}
}

// formatSize formats file size in human-readable format
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(size)/float64(div), "KMGTPE"[exp])
}
