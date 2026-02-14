package filex

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
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

	return "", false
}

// SanitizePath - Sanitize string so it can be used as a valid file name
func SanitizePath(value string) string {
	// Replace all '/' with '_'
	modifiedPath := strings.ReplaceAll(value, `/`, `_`)
	// Comply with OS path valid names
	modifiedPath = symbols.InvalidPathRegExp.ReplaceAllString(modifiedPath, "")

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

	return ""
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

// NextAvailablePath returns the next available path for the given path, optionally with an extension
// For files without an extension or directories, the extension can be omitted
// In case an extension is provided, but it does not match the intended path, it will be ignored
// Example: NextAvailablePath("foo.png") -> "foo.png" if foo.png does not exist
//
//	NextAvailablePath("foo.png", ".png") -> "foo1.png" if foo.png does exist
//	NextAvailablePath("foo.png", ".png") -> "foo12.png" if foo1.png - foo11.png do exist
func NextAvailablePath(path string, ext ...string) string {
	// Return the path if it does not exist
	if !PathExists(path) {
		return path
	}

	// Get the extension
	suffix := ""
	if len(ext) > 0 {
		suffix = ext[0]
	}

	// Ignore the extension if it does not match the path
	if stringsx.IsNotBlank(suffix) && !strings.HasSuffix(path, suffix) {
		suffix = ""
	}

	// Get the base path
	base := strings.TrimSuffix(path, suffix)
	// Construct the glob pattern
	pattern := base + "*" + suffix

	// Find all files matching the pattern
	matches, _ := filepath.Glob(pattern)
	if len(matches) == 0 {
		return base + "1" + suffix
	}

	// Find the highest numbered file
	max := 0
	for _, m := range matches {
		// Trim the extension
		name := strings.TrimSuffix(m, suffix)
		// Trim the base path, extracting the number
		numStr := strings.TrimPrefix(name, base)
		// Convert the number to an integer
		if n, err := strconv.Atoi(numStr); err == nil && n > max {
			max = n
		}
	}

	// Return the next available path
	return fmt.Sprintf("%s%d%s", base, max+1, suffix)
}
