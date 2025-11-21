package filex

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fileTestPaths struct {
	tempDir         string
	tempSubDir      string
	tempFile        string
	nonExistentPath string
}

func setupFileTests(t *testing.T) (fileTestPaths, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp(stringsx.Empty, "filex_test_*")
	require.NoError(t, err, "failed to create temp dir")

	paths := fileTestPaths{
		tempDir:         tempDir,
		tempSubDir:      filepath.Join(tempDir, "subdir"),
		tempFile:        filepath.Join(tempDir, "testfile.txt"),
		nonExistentPath: filepath.Join(tempDir, "nonexistent"),
	}

	require.NoError(t, os.Mkdir(paths.tempSubDir, DirectoryPermission))
	require.NoError(t, os.WriteFile(paths.tempFile, []byte("hello"), FilePermission))

	// Return the paths and the cleanup function.
	return paths, func() { _ = os.RemoveAll(tempDir) }
}

func TestPathExists(t *testing.T) {
	paths, cleanup := setupFileTests(t)
	defer cleanup()

	assert.True(t, PathExists(paths.tempFile), "Expected existing file path to exist")
	assert.True(t, PathExists(paths.tempSubDir), "Expected existing directory path to exist")
	assert.False(t, PathExists(paths.nonExistentPath), "Expected non-existent path to not exist")
}

func TestFileExists(t *testing.T) {
	paths, cleanup := setupFileTests(t)
	defer cleanup()

	assert.True(t, FileExists(paths.tempFile), "Expected existing file to be found")
	assert.False(t, FileExists(paths.tempSubDir), "Expected a directory to not be considered a file")
	assert.False(t, FileExists(paths.nonExistentPath), "Expected a non-existent path to not be a file")
}

func TestDirExists(t *testing.T) {
	paths, cleanup := setupFileTests(t)
	defer cleanup()

	assert.True(t, DirExists(paths.tempSubDir), "Expected existing directory to be found")
	assert.False(t, DirExists(paths.tempFile), "Expected a file to not be considered a directory")
	assert.False(t, DirExists(paths.nonExistentPath), "Expected a non-existent path to not be a directory")
}

func TestGetName(t *testing.T) {
	paths, cleanup := setupFileTests(t)
	defer cleanup()

	name, ok := GetName(paths.tempFile)
	assert.True(t, ok)
	assert.Equal(t, "testfile.txt", name)

	name, ok = GetName(paths.tempSubDir)
	assert.True(t, ok)
	assert.Equal(t, "subdir", name)

	name, ok = GetName(paths.nonExistentPath)
	assert.False(t, ok)
	assert.Empty(t, name)
}

func TestSanitizePath(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "No changes needed", input: "clean-path", expected: "clean-path"},
		{name: "Replace spaces with dashes", input: "this has spaces", expected: "this-has-spaces"},
		{name: "Replace forward slashes", input: "path/to/file", expected: "path_to_file"},
		{name: "Remove invalid characters", input: `a<b>c:d"e/f|g?h*i.`, expected: "abcde_fghi."},
		{name: "Combination of all rules", input: `  invalid / path with?  spaces  `, expected: "invalid-_-path-with-spaces"},
		{name: "Empty string", input: stringsx.Empty, expected: stringsx.Empty},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizePath(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetCWD(t *testing.T) {
	result := GetCWD()

	expected, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	assert.Equal(t, expected, result, "GetCWD should return the current working directory")
	assert.NotEmpty(t, result, "GetCWD should not return an empty string under normal circumstances")
}

func TestCopyFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		srcFile, err := os.CreateTemp(stringsx.Empty, "source-*.txt")
		assert.NoError(t, err)
		defer os.Remove(srcFile.Name())

		content := "hello world"
		_, err = srcFile.WriteString(content)
		assert.NoError(t, err)
		_ = srcFile.Close()

		dstFile, err := os.CreateTemp(stringsx.Empty, "dest-*.txt")
		assert.NoError(t, err)
		dstPath := dstFile.Name()
		_ = dstFile.Close()
		defer os.Remove(dstPath)

		err = CopyFile(srcFile.Name(), dstPath)
		assert.NoError(t, err)

		dstContent, err := os.ReadFile(dstPath)
		assert.NoError(t, err)
		assert.Equal(t, content, string(dstContent))
	})

	t.Run("Source does not exist", func(t *testing.T) {
		err := CopyFile("non-existent-file.txt", "destination.txt")
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})
}
