package ls

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	eve "eve.evalgo.org/common"
	"github.com/spf13/cobra"
)

// Options holds ls configuration
type Options struct {
	All        bool
	Long       bool
	Human      bool
	Recursive  bool
	SortByTime bool
	SortBySize bool
	Reverse    bool
}

// FileEntry represents a file/directory entry
type FileEntry struct {
	Name    string
	Info    fs.FileInfo
	Path    string
	IsDir   bool
	ModTime time.Time
	Size    int64
}

// Command returns the ls command
func Command() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "ls [flags] [paths...]",
		Short: "List directory contents",
		Long:  `List information about files and directories. With no paths, list the current directory.`,
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := args
			if len(paths) == 0 {
				paths = []string{"."}
			}

			for i, path := range paths {
				if err := listPath(path, opts, len(paths) > 1); err != nil {
					eve.Logger.Error("Failed to list", path, ":", err)
				}

				// Add blank line between paths (except after last)
				if i < len(paths)-1 && len(paths) > 1 {
					fmt.Println()
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "Do not ignore entries starting with .")
	cmd.Flags().BoolVarP(&opts.Long, "long", "l", false, "Use a long listing format")
	cmd.Flags().BoolVar(&opts.Human, "human-readable", false, "With -l, print sizes in human readable format")
	cmd.Flags().BoolVarP(&opts.Recursive, "recursive", "R", false, "List subdirectories recursively")
	cmd.Flags().BoolVarP(&opts.SortByTime, "time", "t", false, "Sort by modification time, newest first")
	cmd.Flags().BoolVarP(&opts.SortBySize, "size", "S", false, "Sort by file size, largest first")
	cmd.Flags().BoolVarP(&opts.Reverse, "reverse", "r", false, "Reverse order while sorting")

	return cmd
}

// listPath lists files in a path
func listPath(path string, opts *Options, multiplePaths bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	// If path is a file, just list it
	if !info.IsDir() {
		if opts.Long {
			printLongFormat(&FileEntry{
				Name:    filepath.Base(path),
				Info:    info,
				Path:    path,
				IsDir:   false,
				ModTime: info.ModTime(),
				Size:    info.Size(),
			}, opts)
		} else {
			fmt.Println(path)
		}
		return nil
	}

	// List directory contents
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Print directory name if multiple paths
	if multiplePaths {
		fmt.Printf("%s:\n", path)
	}

	// Convert to FileEntry slice
	fileEntries := make([]FileEntry, 0, len(entries))
	for _, entry := range entries {
		// Skip hidden files unless -a flag
		if !opts.All && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			eve.Logger.Error("Failed to get info for", entry.Name(), ":", err)
			continue
		}

		fileEntries = append(fileEntries, FileEntry{
			Name:    entry.Name(),
			Info:    info,
			Path:    filepath.Join(path, entry.Name()),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime(),
			Size:    info.Size(),
		})
	}

	// Sort entries
	sortEntries(fileEntries, opts)

	// Print entries
	for _, entry := range fileEntries {
		if opts.Long {
			printLongFormat(&entry, opts)
		} else {
			fmt.Println(entry.Name)
		}
	}

	// Handle recursive listing
	if opts.Recursive {
		for _, entry := range fileEntries {
			if entry.IsDir {
				fmt.Println()
				if err := listPath(entry.Path, opts, true); err != nil {
					eve.Logger.Error("Failed to list", entry.Path, ":", err)
				}
			}
		}
	}

	return nil
}

// sortEntries sorts file entries according to options
func sortEntries(entries []FileEntry, opts *Options) {
	sort.Slice(entries, func(i, j int) bool {
		// Sort by time if requested
		if opts.SortByTime {
			if opts.Reverse {
				return entries[i].ModTime.Before(entries[j].ModTime)
			}
			return entries[i].ModTime.After(entries[j].ModTime)
		}

		// Sort by size if requested
		if opts.SortBySize {
			if opts.Reverse {
				return entries[i].Size < entries[j].Size
			}
			return entries[i].Size > entries[j].Size
		}

		// Default: sort by name
		if opts.Reverse {
			return entries[i].Name > entries[j].Name
		}
		return entries[i].Name < entries[j].Name
	})
}

// printLongFormat prints a file entry in long format
func printLongFormat(entry *FileEntry, opts *Options) {
	mode := entry.Info.Mode()
	modTime := entry.ModTime.Format("Jan 02 15:04")
	size := entry.Size

	// Format size
	sizeStr := fmt.Sprintf("%8d", size)
	if opts.Human {
		sizeStr = formatHumanSize(size)
	}

	// Format permissions
	perms := mode.String()

	fmt.Printf("%s %s %s %s\n", perms, sizeStr, modTime, entry.Name)
}

// formatHumanSize formats size in human-readable format
func formatHumanSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%7dB", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"K", "M", "G", "T", "P", "E"}
	return fmt.Sprintf("%6.1f%s", float64(size)/float64(div), units[exp])
}
