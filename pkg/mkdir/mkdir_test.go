package mkdir

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateDirectory tests basic directory creation
func TestCreateDirectory(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		opts    *Options
		wantErr bool
	}{
		{
			name: "SimpleDirectory",
			path: filepath.Join(tempDir, "test1"),
			opts: &Options{
				Mode:    0755,
				Parents: false,
			},
			wantErr: false,
		},
		{
			name: "DirectoryWithParents",
			path: filepath.Join(tempDir, "parent", "child"),
			opts: &Options{
				Mode:    0755,
				Parents: true,
			},
			wantErr: false,
		},
		{
			name: "DirectoryWithoutParents_ShouldFail",
			path: filepath.Join(tempDir, "nonexistent", "child"),
			opts: &Options{
				Mode:    0755,
				Parents: false,
			},
			wantErr: true,
		},
		{
			name: "CustomMode",
			path: filepath.Join(tempDir, "custom_mode"),
			opts: &Options{
				Mode:    0700,
				Parents: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createDirectory(tt.path, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify directory was created
			info, err := os.Stat(tt.path)
			require.NoError(t, err)
			assert.True(t, info.IsDir())

			// Verify permissions (on Unix-like systems)
			if tt.opts.Mode != 0 {
				assert.Equal(t, tt.opts.Mode, info.Mode().Perm())
			}
		})
	}
}

// TestCreateDirectory_AlreadyExists tests handling of existing directories
func TestCreateDirectory_AlreadyExists(t *testing.T) {
	tempDir := t.TempDir()

	// Create directory first
	existingDir := filepath.Join(tempDir, "existing")
	err := os.Mkdir(existingDir, 0755)
	require.NoError(t, err)

	tests := []struct {
		name    string
		path    string
		opts    *Options
		wantErr bool
	}{
		{
			name: "ExistingDirectory_WithoutParents",
			path: existingDir,
			opts: &Options{
				Mode:    0755,
				Parents: false,
			},
			wantErr: true,
		},
		{
			name: "ExistingDirectory_WithParents",
			path: existingDir,
			opts: &Options{
				Mode:    0755,
				Parents: true,
			},
			wantErr: false, // With -p, existing directories are OK
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createDirectory(tt.path, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCreateDirectory_FileExists tests error when path is a file
func TestCreateDirectory_FileExists(t *testing.T) {
	tempDir := t.TempDir()

	// Create a file
	filePath := filepath.Join(tempDir, "file.txt")
	err := os.WriteFile(filePath, []byte("test"), 0644)
	require.NoError(t, err)

	opts := &Options{
		Mode:    0755,
		Parents: false,
	}

	err = createDirectory(filePath, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exists but is not a directory")
}

// TestCreateDirectory_NestedPaths tests creating nested directories
func TestCreateDirectory_NestedPaths(t *testing.T) {
	tempDir := t.TempDir()

	nestedPath := filepath.Join(tempDir, "a", "b", "c", "d")
	opts := &Options{
		Mode:    0755,
		Parents: true,
	}

	err := createDirectory(nestedPath, opts)
	require.NoError(t, err)

	// Verify all directories were created
	info, err := os.Stat(nestedPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify parent directories exist
	for _, path := range []string{
		filepath.Join(tempDir, "a"),
		filepath.Join(tempDir, "a", "b"),
		filepath.Join(tempDir, "a", "b", "c"),
	} {
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	}
}

// TestCreateDirectory_RelativePath tests handling of relative paths
func TestCreateDirectory_RelativePath(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(origDir)
		require.NoError(t, err)
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	opts := &Options{
		Mode:    0755,
		Parents: false,
	}

	relativePath := "relative_test"
	err = createDirectory(relativePath, opts)
	require.NoError(t, err)

	// Verify directory was created
	info, err := os.Stat(relativePath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

// TestCreateDirectory_SpecialCharacters tests paths with special characters
func TestCreateDirectory_SpecialCharacters(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		dirName string
		wantErr bool
		skipWin bool // Skip on Windows
	}{
		{
			name:    "SpacesInName",
			dirName: "dir with spaces",
			wantErr: false,
		},
		{
			name:    "UnderscoresAndDashes",
			dirName: "dir_with-dashes",
			wantErr: false,
		},
		{
			name:    "Dots",
			dirName: "dir.with.dots",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tempDir, tt.dirName)
			opts := &Options{
				Mode:    0755,
				Parents: false,
			}

			err := createDirectory(path, opts)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			info, err := os.Stat(path)
			require.NoError(t, err)
			assert.True(t, info.IsDir())
		})
	}
}

// BenchmarkCreateDirectory benchmarks directory creation
func BenchmarkCreateDirectory(b *testing.B) {
	tempDir := b.TempDir()
	opts := &Options{
		Mode:    0755,
		Parents: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := filepath.Join(tempDir, "bench", string(rune(i)))
		_ = createDirectory(path, opts)
	}
}

// BenchmarkCreateDirectory_WithParents benchmarks directory creation with parents
func BenchmarkCreateDirectory_WithParents(b *testing.B) {
	tempDir := b.TempDir()
	opts := &Options{
		Mode:    0755,
		Parents: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := filepath.Join(tempDir, "bench", "nested", string(rune(i)))
		_ = createDirectory(path, opts)
	}
}
