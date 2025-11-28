package filex

import (
	"io"
	"os"
	"strings"

	"github.com/r3dpixel/toolkit/bytex"
	"github.com/r3dpixel/toolkit/stringsx"
	"github.com/r3dpixel/toolkit/symbols"
)

const (
	DirectoryPermission = 0700 // Default directory permissions given on creation
	FilePermission      = 0644 // Default file permissions given on creation
)

type Entry byte

const (
	File Entry = iota
	Directory
)

type Type string

const (
	Image     Type = "IMAGE"
	JSON      Type = "JSON"
	PNG       Type = "PNG"
	Thumbnail Type = "THUMBNAIL"
)

// PathExists returns true if the specified path exists, false otherwise
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FileExists returns true if the specified path exists AND is a file, false otherwise
func FileExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && !stat.IsDir()
}

// DirExists returns true if the specified path exists AND is a directory, false otherwise
func DirExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

// GetName returns the name of the file/directory at the given path
func GetName(path string) (string, bool) {
	if file, err := os.Stat(path); err == nil {
		return file.Name(), true
	}

	return stringsx.Empty, false
}

// SanitizePath - Sanitize string so it can be used as a valid file name
func SanitizePath(value string) string {
	// Replace all '/' with '_'
	modifiedPath := strings.ReplaceAll(value, `/`, `_`)
	// Comply with OS path valid names
	modifiedPath = symbols.InvalidPathRegExp.ReplaceAllString(modifiedPath, stringsx.Empty)

	// Replace all whitespaces with `-`
	// Split the path into tokens by whitespace separator
	tokens := strings.Fields(modifiedPath)

	// Return sanitized string
	return strings.Join(tokens, symbols.Dash)
}

// GetCWD returns the path a string of the given current working directory
func GetCWD() string {
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}

	return stringsx.Empty
}

// CopyBuffered copies the input io.Reader to the given output io.Writer using a buffer of 32KB
func CopyBuffered(r io.Reader, w io.Writer) error {
	buf := bytex.Buffer32k.Get().([]byte)
	defer bytex.Buffer32k.Put(buf)

	_, err := io.CopyBuffer(w, r, buf)
	return err
}

// CopyFile copies the src file to the dst, using a buffered read/write
func CopyFile(src, dst string) error {
	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy the contents from the source file to the destination file buffered
	return CopyBuffered(srcFile, dstFile)
}
