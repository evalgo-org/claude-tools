package rm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRemovePath_File tests removing a single file
func TestRemovePath_File(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	opts := &Options{
		Recursive: false,
		Force:     false,
		Verbose:   false,
	}

	err = removePath(testFile, opts)
	require.NoError(t, err)

	// Verify file was removed
	_, err = os.Stat(testFile)
	assert.True(t, os.IsNotExist(err))
}

// TestRemovePath_Directory_WithoutRecursive tests that removing directory fails without -r
func TestRemovePath_Directory_WithoutRecursive(t *testing.T) {
	tempDir := t.TempDir()

	testDir := filepath.Join(tempDir, "testdir")
	err := os.Mkdir(testDir, 0755)
	require.NoError(t, err)

	opts := &Options{
		Recursive: false,
		Force:     false,
		Verbose:   false,
	}

	err = removePath(testDir, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Is a directory")

	// Verify directory still exists
	_, err = os.Stat(testDir)
	assert.NoError(t, err)
}

// TestRemovePath_Directory_WithRecursive tests removing directory with -r
func TestRemovePath_Directory_WithRecursive(t *testing.T) {
	tempDir := t.TempDir()

	// Create directory with files
	testDir := filepath.Join(tempDir, "testdir")
	err := os.Mkdir(testDir, 0755)
	require.NoError(t, err)

	testFile := filepath.Join(testDir, "file.txt")
	err = os.WriteFile(testFile, []byte("content"), 0644)
	require.NoError(t, err)

	opts := &Options{
		Recursive: true,
		Force:     false,
		Verbose:   false,
	}

	err = removePath(testDir, opts)
	require.NoError(t, err)

	// Verify directory was removed
	_, err = os.Stat(testDir)
	assert.True(t, os.IsNotExist(err))
}

// TestRemovePath_NonexistentFile_WithForce tests that nonexistent files are ignored with -f
func TestRemovePath_NonexistentFile_WithForce(t *testing.T) {
	tempDir := t.TempDir()

	nonexistent := filepath.Join(tempDir, "nonexistent.txt")

	opts := &Options{
		Recursive: false,
		Force:     true,
		Verbose:   false,
	}

	err := removePath(nonexistent, opts)
	assert.NoError(t, err) // With -f, nonexistent files should not error
}

// TestRemovePath_NonexistentFile_WithoutForce tests that nonexistent files error without -f
func TestRemovePath_NonexistentFile_WithoutForce(t *testing.T) {
	tempDir := t.TempDir()

	nonexistent := filepath.Join(tempDir, "nonexistent.txt")

	opts := &Options{
		Recursive: false,
		Force:     false,
		Verbose:   false,
	}

	err := removePath(nonexistent, opts)
	assert.Error(t, err)
}

// TestRemovePath_NestedDirectory tests removing nested directories
func TestRemovePath_NestedDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create nested structure
	nestedPath := filepath.Join(tempDir, "a", "b", "c")
	err := os.MkdirAll(nestedPath, 0755)
	require.NoError(t, err)

	// Add files at different levels
	err = os.WriteFile(filepath.Join(tempDir, "a", "file1.txt"), []byte("1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "a", "b", "file2.txt"), []byte("2"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(nestedPath, "file3.txt"), []byte("3"), 0644)
	require.NoError(t, err)

	opts := &Options{
		Recursive: true,
		Force:     false,
		Verbose:   false,
	}

	topDir := filepath.Join(tempDir, "a")
	err = removePath(topDir, opts)
	require.NoError(t, err)

	// Verify entire tree was removed
	_, err = os.Stat(topDir)
	assert.True(t, os.IsNotExist(err))
}

// TestRemovePath_MultipleFiles tests removing multiple files
func TestRemovePath_MultipleFiles(t *testing.T) {
	tempDir := t.TempDir()

	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	opts := &Options{
		Recursive: false,
		Force:     false,
		Verbose:   false,
	}

	// Create files
	for _, f := range files {
		path := filepath.Join(tempDir, f)
		err := os.WriteFile(path, []byte("content"), 0644)
		require.NoError(t, err)
	}

	// Remove files
	for _, f := range files {
		path := filepath.Join(tempDir, f)
		err := removePath(path, opts)
		require.NoError(t, err)
	}

	// Verify all files were removed
	for _, f := range files {
		path := filepath.Join(tempDir, f)
		_, err := os.Stat(path)
		assert.True(t, os.IsNotExist(err))
	}
}

// TestRemovePath_Symlink tests removing symlinks
func TestRemovePath_Symlink(t *testing.T) {
	tempDir := t.TempDir()

	// Create target file
	targetFile := filepath.Join(tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target"), 0644)
	require.NoError(t, err)

	// Create symlink
	linkPath := filepath.Join(tempDir, "link.txt")
	err = os.Symlink(targetFile, linkPath)
	require.NoError(t, err)

	opts := &Options{
		Recursive: false,
		Force:     false,
		Verbose:   false,
	}

	// Remove symlink (should not remove target)
	err = removePath(linkPath, opts)
	require.NoError(t, err)

	// Verify symlink was removed but target still exists
	_, err = os.Lstat(linkPath)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(targetFile)
	assert.NoError(t, err)
}
