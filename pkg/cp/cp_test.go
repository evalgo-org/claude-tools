package cp

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCopyFile_Simple tests basic file copying
func TestCopyFile_Simple(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	destFile := filepath.Join(tempDir, "dest.txt")

	content := []byte("test content")
	err := os.WriteFile(srcFile, content, 0644)
	require.NoError(t, err)

	opts := &Options{
		Recursive: false,
		Preserve:  false,
		Verbose:   false,
		Force:     false,
	}

	err = copyFile(srcFile, destFile, opts)
	require.NoError(t, err)

	// Verify content
	destContent, err := os.ReadFile(destFile)
	require.NoError(t, err)
	assert.Equal(t, content, destContent)
}

// TestCopyFile_PreserveTimestamps tests -p flag
func TestCopyFile_PreserveTimestamps(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	destFile := filepath.Join(tempDir, "dest.txt")

	err := os.WriteFile(srcFile, []byte("content"), 0644)
	require.NoError(t, err)

	// Set specific modification time
	modTime := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	err = os.Chtimes(srcFile, modTime, modTime)
	require.NoError(t, err)

	opts := &Options{
		Recursive: false,
		Preserve:  true,
		Verbose:   false,
		Force:     false,
	}

	err = copyFile(srcFile, destFile, opts)
	require.NoError(t, err)

	// Verify timestamps
	destInfo, err := os.Stat(destFile)
	require.NoError(t, err)
	assert.Equal(t, modTime.Unix(), destInfo.ModTime().Unix())
}

// TestCopyFile_ExistingFile_WithoutForce tests that existing files error without -f
func TestCopyFile_ExistingFile_WithoutForce(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	destFile := filepath.Join(tempDir, "dest.txt")

	err := os.WriteFile(srcFile, []byte("source"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(destFile, []byte("existing"), 0644)
	require.NoError(t, err)

	opts := &Options{
		Recursive: false,
		Preserve:  false,
		Verbose:   false,
		Force:     false,
	}

	err = copyFile(srcFile, destFile, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// TestCopyFile_ExistingFile_WithForce tests overwriting with -f
func TestCopyFile_ExistingFile_WithForce(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	destFile := filepath.Join(tempDir, "dest.txt")

	srcContent := []byte("new content")
	err := os.WriteFile(srcFile, srcContent, 0644)
	require.NoError(t, err)
	err = os.WriteFile(destFile, []byte("old content"), 0644)
	require.NoError(t, err)

	opts := &Options{
		Recursive: false,
		Preserve:  false,
		Verbose:   false,
		Force:     true,
	}

	err = copyFile(srcFile, destFile, opts)
	require.NoError(t, err)

	// Verify content was overwritten
	destContent, err := os.ReadFile(destFile)
	require.NoError(t, err)
	assert.Equal(t, srcContent, destContent)
}

// TestCopyFiles_MultipleToDirectory tests copying multiple files to directory
func TestCopyFiles_MultipleToDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create source files
	src1 := filepath.Join(tempDir, "file1.txt")
	src2 := filepath.Join(tempDir, "file2.txt")
	err := os.WriteFile(src1, []byte("content1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(src2, []byte("content2"), 0644)
	require.NoError(t, err)

	// Create destination directory
	destDir := filepath.Join(tempDir, "dest")
	err = os.Mkdir(destDir, 0755)
	require.NoError(t, err)

	opts := &Options{
		Recursive: false,
		Preserve:  false,
		Verbose:   false,
		Force:     false,
	}

	err = copyFiles([]string{src1, src2}, destDir, opts)
	require.NoError(t, err)

	// Verify files were copied
	dest1 := filepath.Join(destDir, "file1.txt")
	dest2 := filepath.Join(destDir, "file2.txt")

	content1, err := os.ReadFile(dest1)
	require.NoError(t, err)
	assert.Equal(t, []byte("content1"), content1)

	content2, err := os.ReadFile(dest2)
	require.NoError(t, err)
	assert.Equal(t, []byte("content2"), content2)
}

// TestCopyFiles_MultipleToNonDirectory tests error when copying multiple to non-directory
func TestCopyFiles_MultipleToNonDirectory(t *testing.T) {
	tempDir := t.TempDir()

	src1 := filepath.Join(tempDir, "file1.txt")
	src2 := filepath.Join(tempDir, "file2.txt")
	destFile := filepath.Join(tempDir, "dest.txt")

	err := os.WriteFile(src1, []byte("1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(src2, []byte("2"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(destFile, []byte("dest"), 0644)
	require.NoError(t, err)

	opts := &Options{
		Recursive: false,
		Preserve:  false,
		Verbose:   false,
		Force:     false,
	}

	err = copyFiles([]string{src1, src2}, destFile, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

// TestCopyDir_Recursive tests copying directories with -r
func TestCopyDir_Recursive(t *testing.T) {
	tempDir := t.TempDir()

	// Create source directory structure
	srcDir := filepath.Join(tempDir, "source")
	err := os.Mkdir(srcDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644)
	require.NoError(t, err)

	subDir := filepath.Join(srcDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("content2"), 0644)
	require.NoError(t, err)

	destDir := filepath.Join(tempDir, "dest")

	opts := &Options{
		Recursive: true,
		Preserve:  false,
		Verbose:   false,
		Force:     false,
	}

	err = copyDir(srcDir, destDir, opts)
	require.NoError(t, err)

	// Verify structure was copied
	file1 := filepath.Join(destDir, "file1.txt")
	content1, err := os.ReadFile(file1)
	require.NoError(t, err)
	assert.Equal(t, []byte("content1"), content1)

	file2 := filepath.Join(destDir, "subdir", "file2.txt")
	content2, err := os.ReadFile(file2)
	require.NoError(t, err)
	assert.Equal(t, []byte("content2"), content2)
}

// TestCopyFiles_DirectoryWithoutRecursive tests error when copying directory without -r
func TestCopyFiles_DirectoryWithoutRecursive(t *testing.T) {
	tempDir := t.TempDir()

	srcDir := filepath.Join(tempDir, "source")
	err := os.Mkdir(srcDir, 0755)
	require.NoError(t, err)

	destDir := filepath.Join(tempDir, "dest")

	opts := &Options{
		Recursive: false,
		Preserve:  false,
		Verbose:   false,
		Force:     false,
	}

	err = copyFiles([]string{srcDir}, destDir, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is a directory")
	assert.Contains(t, err.Error(), "use -r")
}

// TestCopyDir_PreservePermissions tests that directory permissions are preserved
func TestCopyDir_PreservePermissions(t *testing.T) {
	tempDir := t.TempDir()

	srcDir := filepath.Join(tempDir, "source")
	err := os.Mkdir(srcDir, 0755)
	require.NoError(t, err)

	destDir := filepath.Join(tempDir, "dest")

	opts := &Options{
		Recursive: true,
		Preserve:  true,
		Verbose:   false,
		Force:     false,
	}

	err = copyDir(srcDir, destDir, opts)
	require.NoError(t, err)

	// Verify permissions
	srcInfo, err := os.Stat(srcDir)
	require.NoError(t, err)
	destInfo, err := os.Stat(destDir)
	require.NoError(t, err)

	assert.Equal(t, srcInfo.Mode().Perm(), destInfo.Mode().Perm())
}
