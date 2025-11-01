package touch

import (
	"fmt"
	"os"
	"time"

	eve "eve.evalgo.org/common"
	"github.com/spf13/cobra"
)

// Options holds touch configuration
type Options struct {
	NoCreate   bool
	AccessOnly bool
	ModifyOnly bool
	Timestamp  string
	Verbose    bool
}

// Command returns the touch command
func Command() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "touch [flags] files...",
		Short: "Create empty files or update file timestamps",
		Long: `Update the access and modification times of each file to the current time.

If a file does not exist, it is created empty, unless -c is specified.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate options
			if opts.AccessOnly && opts.ModifyOnly {
				return fmt.Errorf("cannot specify both -a and -m")
			}

			// Parse timestamp if provided
			var timestamp time.Time
			var err error
			if opts.Timestamp != "" {
				timestamp, err = parseTimestamp(opts.Timestamp)
				if err != nil {
					return fmt.Errorf("invalid timestamp format: %w", err)
				}
			} else {
				timestamp = time.Now()
			}

			for _, path := range args {
				if err := touchFile(path, timestamp, opts); err != nil {
					eve.Logger.Error("Failed to touch", path, ":", err)
					return err
				}

				if opts.Verbose {
					fmt.Printf("touched '%s'\n", path)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.NoCreate, "no-create", "c", false, "Do not create files that do not exist")
	cmd.Flags().BoolVarP(&opts.AccessOnly, "access", "a", false, "Change only the access time")
	cmd.Flags().BoolVarP(&opts.ModifyOnly, "modify", "m", false, "Change only the modification time")
	cmd.Flags().StringVarP(&opts.Timestamp, "time", "t", "", "Use specified time instead of current time (format: YYYYMMDDhhmm[.ss])")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Explain what is being done")

	return cmd
}

// touchFile creates or updates a file's timestamp
func touchFile(path string, timestamp time.Time, opts *Options) error {
	// Check if file exists
	info, err := os.Stat(path)
	fileExists := err == nil

	if !fileExists {
		if os.IsNotExist(err) {
			if opts.NoCreate {
				// -c flag: don't create file
				return nil
			}

			// Create empty file
			file, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			file.Close()

			// Set initial timestamps
			if err := os.Chtimes(path, timestamp, timestamp); err != nil {
				return fmt.Errorf("failed to set timestamps: %w", err)
			}

			return nil
		}
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// File exists, update timestamps
	accessTime := timestamp
	modifyTime := timestamp

	// If only changing one time, preserve the other
	if opts.AccessOnly {
		modifyTime = info.ModTime()
	} else if opts.ModifyOnly {
		// Get current access time (use modtime as fallback since Go doesn't expose atime easily)
		accessTime = info.ModTime()
	}

	if err := os.Chtimes(path, accessTime, modifyTime); err != nil {
		return fmt.Errorf("failed to update timestamps: %w", err)
	}

	return nil
}

// parseTimestamp parses timestamp in format YYYYMMDDhhmm[.ss]
func parseTimestamp(s string) (time.Time, error) {
	var t time.Time
	var err error

	// Try with seconds: YYYYMMDDhhmm.ss
	if len(s) == 15 && s[12] == '.' {
		t, err = time.Parse("200601021504.05", s)
		if err == nil {
			return t, nil
		}
	}

	// Try without seconds: YYYYMMDDhhmm
	if len(s) == 12 {
		t, err = time.Parse("200601021504", s)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("timestamp must be in format YYYYMMDDhhmm[.ss], got: %s", s)
}
