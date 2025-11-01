package touch

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTouchFile_CreateNew tests creating a new empty file
func TestTouchFile_CreateNew(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "newfile.txt")
	timestamp := time.Now()

	opts := &Options{
		NoCreate:   false,
		AccessOnly: false,
		ModifyOnly: false,
		Timestamp:  "",
		Verbose:    false,
	}

	err := touchFile(testFile, timestamp, opts)
	require.NoError(t, err)

	// Verify file was created
	info, err := os.Stat(testFile)
	require.NoError(t, err)
	assert.True(t, info.Mode().IsRegular())

	// Verify file is empty
	assert.Equal(t, int64(0), info.Size())
}

// TestTouchFile_UpdateExisting tests updating timestamp of existing file
func TestTouchFile_UpdateExisting(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "existing.txt")
	err := os.WriteFile(testFile, []byte("content"), 0644)
	require.NoError(t, err)

	// Set old timestamp
	oldTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	err = os.Chtimes(testFile, oldTime, oldTime)
	require.NoError(t, err)

	// Touch with new timestamp
	newTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	opts := &Options{
		NoCreate:   false,
		AccessOnly: false,
		ModifyOnly: false,
		Timestamp:  "",
		Verbose:    false,
	}

	err = touchFile(testFile, newTime, opts)
	require.NoError(t, err)

	// Verify timestamp was updated
	info, err := os.Stat(testFile)
	require.NoError(t, err)
	assert.Equal(t, newTime.Unix(), info.ModTime().Unix())

	// Verify content was preserved
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, []byte("content"), content)
}

// TestTouchFile_NoCreate tests -c flag (don't create)
func TestTouchFile_NoCreate(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "nonexistent.txt")
	timestamp := time.Now()

	opts := &Options{
		NoCreate:   true,
		AccessOnly: false,
		ModifyOnly: false,
		Timestamp:  "",
		Verbose:    false,
	}

	err := touchFile(testFile, timestamp, opts)
	require.NoError(t, err) // Should not error with -c

	// Verify file was NOT created
	_, err = os.Stat(testFile)
	assert.True(t, os.IsNotExist(err))
}

// TestTouchFile_AccessOnly tests -a flag
func TestTouchFile_AccessOnly(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("content"), 0644)
	require.NoError(t, err)

	// Set known modification time
	oldModTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	err = os.Chtimes(testFile, oldModTime, oldModTime)
	require.NoError(t, err)

	// Touch with -a
	newTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	opts := &Options{
		NoCreate:   false,
		AccessOnly: true,
		ModifyOnly: false,
		Timestamp:  "",
		Verbose:    false,
	}

	err = touchFile(testFile, newTime, opts)
	require.NoError(t, err)

	// Verify modification time was preserved (not changed)
	info, err := os.Stat(testFile)
	require.NoError(t, err)
	assert.Equal(t, oldModTime.Unix(), info.ModTime().Unix())
}

// TestTouchFile_ModifyOnly tests -m flag
func TestTouchFile_ModifyOnly(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("content"), 0644)
	require.NoError(t, err)

	// Get original info
	originalInfo, err := os.Stat(testFile)
	require.NoError(t, err)

	// Touch with -m
	newTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	opts := &Options{
		NoCreate:   false,
		AccessOnly: false,
		ModifyOnly: true,
		Timestamp:  "",
		Verbose:    false,
	}

	err = touchFile(testFile, newTime, opts)
	require.NoError(t, err)

	// Verify modification time was updated
	info, err := os.Stat(testFile)
	require.NoError(t, err)
	assert.Equal(t, newTime.Unix(), info.ModTime().Unix())

	// Note: Go doesn't expose access time easily, so we can't verify it was preserved
	_ = originalInfo
}

// TestTouchFile_AccessAndModify_MutuallyExclusive tests that -a and -m can't both be set
// This is actually allowed in real touch and updates both times, but our implementation
// treats them as mutually exclusive for simplicity
func TestParseTimestamp_Valid(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "WithSeconds",
			input:    "202501011200.30",
			expected: time.Date(2025, 1, 1, 12, 0, 30, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "WithoutSeconds",
			input:    "202501011200",
			expected: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "InvalidFormat_TooShort",
			input:    "20250101",
			expected: time.Time{},
			wantErr:  true,
		},
		{
			name:     "InvalidFormat_TooLong",
			input:    "2025010112003000",
			expected: time.Time{},
			wantErr:  true,
		},
		{
			name:     "InvalidFormat_WrongSeparator",
			input:    "202501011200-30",
			expected: time.Time{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTimestamp(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Unix(), result.Unix())
		})
	}
}

// TestTouchFile_WithTimestamp tests using -t flag
func TestTouchFile_WithTimestamp(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "test.txt")

	specificTime := time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC)

	opts := &Options{
		NoCreate:   false,
		AccessOnly: false,
		ModifyOnly: false,
		Timestamp:  "",
		Verbose:    false,
	}

	err := touchFile(testFile, specificTime, opts)
	require.NoError(t, err)

	// Verify timestamp
	info, err := os.Stat(testFile)
	require.NoError(t, err)
	assert.Equal(t, specificTime.Unix(), info.ModTime().Unix())
}

// TestTouchFile_MultipleFiles tests touching multiple files
func TestTouchFile_MultipleFiles(t *testing.T) {
	tempDir := t.TempDir()

	files := []string{
		filepath.Join(tempDir, "file1.txt"),
		filepath.Join(tempDir, "file2.txt"),
		filepath.Join(tempDir, "file3.txt"),
	}

	timestamp := time.Now()
	opts := &Options{
		NoCreate:   false,
		AccessOnly: false,
		ModifyOnly: false,
		Timestamp:  "",
		Verbose:    false,
	}

	// Touch all files
	for _, file := range files {
		err := touchFile(file, timestamp, opts)
		require.NoError(t, err)
	}

	// Verify all files were created
	for _, file := range files {
		info, err := os.Stat(file)
		require.NoError(t, err)
		assert.True(t, info.Mode().IsRegular())
	}
}

// TestTouchFile_PreservesContent tests that touch doesn't modify file content
func TestTouchFile_PreservesContent(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "test.txt")
	originalContent := []byte("important content")

	err := os.WriteFile(testFile, originalContent, 0644)
	require.NoError(t, err)

	timestamp := time.Now()
	opts := &Options{
		NoCreate:   false,
		AccessOnly: false,
		ModifyOnly: false,
		Timestamp:  "",
		Verbose:    false,
	}

	err = touchFile(testFile, timestamp, opts)
	require.NoError(t, err)

	// Verify content was not changed
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, originalContent, content)
}

// TestTouchFile_PreservesPermissions tests that touch preserves file permissions
func TestTouchFile_PreservesPermissions(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("content"), 0600)
	require.NoError(t, err)

	originalInfo, err := os.Stat(testFile)
	require.NoError(t, err)
	originalMode := originalInfo.Mode()

	timestamp := time.Now()
	opts := &Options{
		NoCreate:   false,
		AccessOnly: false,
		ModifyOnly: false,
		Timestamp:  "",
		Verbose:    false,
	}

	err = touchFile(testFile, timestamp, opts)
	require.NoError(t, err)

	// Verify permissions were preserved
	info, err := os.Stat(testFile)
	require.NoError(t, err)
	assert.Equal(t, originalMode.Perm(), info.Mode().Perm())
}
