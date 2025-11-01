package mv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMoveFiles_SimpleRename tests renaming a single file
func TestMoveFiles_SimpleRename(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	destFile := filepath.Join(tempDir, "dest.txt")

	content := []byte("test content")
	err := os.WriteFile(srcFile, content, 0644)
	require.NoError(t, err)

	opts := &Options{
		Force:     false,
		NoClobber: false,
		Verbose:   false,
	}

	err = moveFiles([]string{srcFile}, destFile, opts)
	require.NoError(t, err)

	// Verify source was removed
	_, err = os.Stat(srcFile)
	assert.True(t, os.IsNotExist(err))

	// Verify destination exists with correct content
	destContent, err := os.ReadFile(destFile)
	require.NoError(t, err)
	assert.Equal(t, content, destContent)
}

// TestMoveFiles_ToDirectory tests moving files into a directory
func TestMoveFiles_ToDirectory(t *testing.T) {
	tempDir := t.TempDir()

	src1 := filepath.Join(tempDir, "file1.txt")
	src2 := filepath.Join(tempDir, "file2.txt")
	err := os.WriteFile(src1, []byte("content1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(src2, []byte("content2"), 0644)
	require.NoError(t, err)

	destDir := filepath.Join(tempDir, "dest")
	err = os.Mkdir(destDir, 0755)
	require.NoError(t, err)

	opts := &Options{
		Force:     false,
		NoClobber: false,
		Verbose:   false,
	}

	err = moveFiles([]string{src1, src2}, destDir, opts)
	require.NoError(t, err)

	// Verify sources were removed
	_, err = os.Stat(src1)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(src2)
	assert.True(t, os.IsNotExist(err))

	// Verify files were moved to destination
	dest1 := filepath.Join(destDir, "file1.txt")
	dest2 := filepath.Join(destDir, "file2.txt")

	content1, err := os.ReadFile(dest1)
	require.NoError(t, err)
	assert.Equal(t, []byte("content1"), content1)

	content2, err := os.ReadFile(dest2)
	require.NoError(t, err)
	assert.Equal(t, []byte("content2"), content2)
}

// TestMoveFiles_ExistingFile_WithoutForce tests error when destination exists without -f
func TestMoveFiles_ExistingFile_WithoutForce(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	destFile := filepath.Join(tempDir, "dest.txt")

	err := os.WriteFile(srcFile, []byte("source"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(destFile, []byte("existing"), 0644)
	require.NoError(t, err)

	opts := &Options{
		Force:     false,
		NoClobber: false,
		Verbose:   false,
	}

	err = moveFiles([]string{srcFile}, destFile, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Verify source still exists
	_, err = os.Stat(srcFile)
	assert.NoError(t, err)
}

// TestMoveFiles_ExistingFile_WithForce tests overwriting with -f
func TestMoveFiles_ExistingFile_WithForce(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	destFile := filepath.Join(tempDir, "dest.txt")

	srcContent := []byte("new content")
	err := os.WriteFile(srcFile, srcContent, 0644)
	require.NoError(t, err)
	err = os.WriteFile(destFile, []byte("old content"), 0644)
	require.NoError(t, err)

	opts := &Options{
		Force:     true,
		NoClobber: false,
		Verbose:   false,
	}

	err = moveFiles([]string{srcFile}, destFile, opts)
	require.NoError(t, err)

	// Verify source was removed
	_, err = os.Stat(srcFile)
	assert.True(t, os.IsNotExist(err))

	// Verify destination has new content
	destContent, err := os.ReadFile(destFile)
	require.NoError(t, err)
	assert.Equal(t, srcContent, destContent)
}

// TestMoveFiles_NoClobber tests -n flag
func TestMoveFiles_NoClobber(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	destFile := filepath.Join(tempDir, "dest.txt")

	srcContent := []byte("source")
	destContent := []byte("existing")

	err := os.WriteFile(srcFile, srcContent, 0644)
	require.NoError(t, err)
	err = os.WriteFile(destFile, destContent, 0644)
	require.NoError(t, err)

	opts := &Options{
		Force:     false,
		NoClobber: true,
		Verbose:   false,
	}

	err = moveFiles([]string{srcFile}, destFile, opts)
	require.NoError(t, err) // -n should not error, just skip

	// Verify source still exists
	_, err = os.Stat(srcFile)
	assert.NoError(t, err)

	// Verify destination was not overwritten
	content, err := os.ReadFile(destFile)
	require.NoError(t, err)
	assert.Equal(t, destContent, content)
}

// TestMoveFiles_ForceAndNoClobber tests error when both -f and -n are specified
func TestMoveFiles_ForceAndNoClobber(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	destFile := filepath.Join(tempDir, "dest.txt")

	err := os.WriteFile(srcFile, []byte("source"), 0644)
	require.NoError(t, err)

	opts := &Options{
		Force:     true,
		NoClobber: true,
		Verbose:   false,
	}

	err = moveFiles([]string{srcFile}, destFile, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot specify both")
}

// TestMoveFiles_Directory tests moving a directory
func TestMoveFiles_Directory(t *testing.T) {
	tempDir := t.TempDir()

	// Create source directory with content
	srcDir := filepath.Join(tempDir, "source")
	err := os.Mkdir(srcDir, 0755)
	require.NoError(t, err)

	srcFile := filepath.Join(srcDir, "file.txt")
	err = os.WriteFile(srcFile, []byte("content"), 0644)
	require.NoError(t, err)

	destDir := filepath.Join(tempDir, "dest")

	opts := &Options{
		Force:     false,
		NoClobber: false,
		Verbose:   false,
	}

	err = moveFiles([]string{srcDir}, destDir, opts)
	require.NoError(t, err)

	// Verify source directory was removed
	_, err = os.Stat(srcDir)
	assert.True(t, os.IsNotExist(err))

	// Verify destination directory exists with content
	destFile := filepath.Join(destDir, "file.txt")
	content, err := os.ReadFile(destFile)
	require.NoError(t, err)
	assert.Equal(t, []byte("content"), content)
}

// TestMoveFiles_MultipleToNonDirectory tests error when moving multiple to non-directory
func TestMoveFiles_MultipleToNonDirectory(t *testing.T) {
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
		Force:     false,
		NoClobber: false,
		Verbose:   false,
	}

	err = moveFiles([]string{src1, src2}, destFile, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

// TestMoveFiles_PreservesPermissions tests that permissions are preserved
func TestMoveFiles_PreservesPermissions(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	destFile := filepath.Join(tempDir, "dest.txt")

	err := os.WriteFile(srcFile, []byte("content"), 0600)
	require.NoError(t, err)

	srcInfo, err := os.Stat(srcFile)
	require.NoError(t, err)
	srcMode := srcInfo.Mode()

	opts := &Options{
		Force:     false,
		NoClobber: false,
		Verbose:   false,
	}

	err = moveFiles([]string{srcFile}, destFile, opts)
	require.NoError(t, err)

	// Verify permissions were preserved
	destInfo, err := os.Stat(destFile)
	require.NoError(t, err)
	assert.Equal(t, srcMode.Perm(), destInfo.Mode().Perm())
}

// TestCopyAndDelete_CrossFilesystem simulates cross-filesystem move
func TestCopyAndDelete_CrossFilesystem(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	destFile := filepath.Join(tempDir, "dest.txt")

	content := []byte("test content")
	err := os.WriteFile(srcFile, content, 0644)
	require.NoError(t, err)

	srcInfo, err := os.Stat(srcFile)
	require.NoError(t, err)

	// Test copyAndDelete directly
	err = copyAndDelete(srcFile, destFile, srcInfo)
	require.NoError(t, err)

	// Verify source was removed
	_, err = os.Stat(srcFile)
	assert.True(t, os.IsNotExist(err))

	// Verify destination exists with correct content
	destContent, err := os.ReadFile(destFile)
	require.NoError(t, err)
	assert.Equal(t, content, destContent)
}
