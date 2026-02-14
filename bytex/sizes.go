package bytex

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/r3dpixel/toolkit/stringsx"
)

// Size represents a byte size.
// The representation limits the largest representable size to approximately 8 exabytes.
type Size int64

const (
	B   Size = 1
	KB  Size = 1000
	KiB Size = 1024
	MB  Size = 1000 * KB
	MiB Size = 1024 * KiB
	GB  Size = 1000 * MB
	GiB Size = 1024 * MiB
	TB  Size = 1000 * GB
	TiB Size = 1024 * GiB
	PB  Size = 1000 * TB
	PiB Size = 1024 * TiB
	EB  Size = 1000 * PB
	EiB Size = 1024 * PiB

	// defaultPrecision is the default precision for formatting when not specified
	defaultPrecision = 2
)

// unit represents a byte size unit with its value and name
type unit struct {
	Size Size
	Name string
}

// unitList is the ordered list of units from largest to smallest
var unitList = []unit{
	{EiB, "EiB"},
	{EB, "EB"},
	{PiB, "PiB"},
	{PB, "PB"},
	{TiB, "TiB"},
	{TB, "TB"},
	{GiB, "GiB"},
	{GB, "GB"},
	{MiB, "MiB"},
	{MB, "MB"},
	{KiB, "KiB"},
	{KB, "KB"},
	{B, "B"},
}

// unitsBySize maps from Size to unit
var unitsBySize = make(map[Size]unit)

// unitsByName maps from string name to unit
var unitsByName = make(map[string]unit)

func init() {
	for _, u := range unitList {
		unitsBySize[u.Size] = u
		unitsByName[strings.ToUpper(u.Name)] = u
	}
}

// String returns a string representation of the size.
// It uses the largest unit that evenly divides the size.
func (s Size) String() string {
	if s == 0 {
		return "0B"
	}

	sign := ""
	if s < 0 {
		sign, s = "-", -s
	}

	for _, u := range unitList {
		if s >= u.Size && s%u.Size == 0 {
			return fmt.Sprintf("%s%d%s", sign, s/u.Size, u.Name)
		}
	}
	return fmt.Sprintf("%s%dB", sign, s)
}

// HumanReadable returns a human-readable string using the highest appropriate unit with fractions.
// It automatically selects the best unit that keeps the value >= 1.
func (s Size) HumanReadable() string {
	if s == 0 {
		return "0B"
	}

	abs := s
	if s < 0 {
		abs = -s
	}

	// Find the best unit (largest unit where value >= 1)
	for _, u := range unitList {
		if abs >= u.Size {
			return s.Format(u.Size)
		}
	}

	return s.Format(B)
}

// Format returns the size formatted with the specified unit.
// The unit should be one of: B, KB, KiB, MB, MiB, GB, GiB, TB, TiB, PB, PiB, EB, EiB.
// Precision is optional - if not provided or negative, it defaults to defaultPrecision.
func (s Size) Format(unit Size, precision ...int) string {
	u, ok := unitsBySize[unit]
	if !ok {
		return fmt.Sprintf("%dB (invalid unit: %d)", s, unit)
	}

	value := float64(s) / float64(unit)

	prec := defaultPrecision
	if len(precision) > 0 && precision[0] >= 0 {
		prec = precision[0]
	}

	return fmt.Sprintf("%.*f%s", prec, value, u.Name)
}

// BytesString returns the size as a string with "B" suffix.
func (s Size) BytesString() string {
	return fmt.Sprintf("%dB", s)
}

// ParseSize parses a size string. A size string is a possibly signed sequence of
// decimal numbers, each with optional fraction and a unit suffix, such as "300KB", "1.5GiB" or "2.5MB".
// Valid size units are "B", "KB", "KiB", "MB", "MiB", "GB", "GiB", "TB", "TiB", "PB", "PiB", "EB", "EiB".
func ParseSize(s string) (Size, error) {
	if stringsx.IsBlank(s) {
		return 0, nil
	}

	// Extract sign
	sign := 1
	switch s[0] {
	case '+':
		s = s[1:]
		sign = 1
	case '-':
		s = s[1:]
		sign = -1
	}

	// Find where number ends
	i := strings.IndexFunc(s, func(r rune) bool {
		return r != '.' && (r < '0' || r > '9')
	})
	if i == -1 {
		i = len(s)
	}
	if i == 0 {
		return 0, fmt.Errorf("invalid size: missing number")
	}

	value, err := strconv.ParseFloat(s[:i], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size: %v", err)
	}

	// Parse unit
	unit := strings.ToUpper(strings.TrimSpace(s[i:]))
	multiplier := B
	if stringsx.IsNotBlank(unit) {
		u, ok := unitsByName[unit]
		if !ok {
			return 0, fmt.Errorf("invalid size: unknown unit %q", unit)
		}
		multiplier = u.Size
	}

	return Size(value * float64(multiplier) * float64(sign)), nil
}

// Bytes returns the size as a floating-point number of bytes.
func (s Size) Bytes() float64 {
	return float64(s)
}

// Kilobytes returns the size as a floating-point number of kilobytes.
func (s Size) Kilobytes() float64 {
	return float64(s) / float64(KB)
}

// Kibibytes returns the size as a floating-point number of kibibytes.
func (s Size) Kibibytes() float64 {
	return float64(s) / float64(KiB)
}

// Megabytes returns the size as a floating-point number of megabytes.
func (s Size) Megabytes() float64 {
	return float64(s) / float64(MB)
}

// Mebibytes returns the size as a floating-point number of mebibytes.
func (s Size) Mebibytes() float64 {
	return float64(s) / float64(MiB)
}

// Gigabytes returns the size as a floating-point number of gigabytes.
func (s Size) Gigabytes() float64 {
	return float64(s) / float64(GB)
}

// Gibibytes returns the size as a floating-point number of gibibytes.
func (s Size) Gibibytes() float64 {
	return float64(s) / float64(GiB)
}

// Terabytes returns the size as a floating-point number of terabytes.
func (s Size) Terabytes() float64 {
	return float64(s) / float64(TB)
}

// Tebibytes returns the size as a floating-point number of tebibytes.
func (s Size) Tebibytes() float64 {
	return float64(s) / float64(TiB)
}

// Petabytes returns the size as a floating-point number of petabytes.
func (s Size) Petabytes() float64 {
	return float64(s) / float64(PB)
}

// Pebibytes returns the size as a floating-point number of pebibytes.
func (s Size) Pebibytes() float64 {
	return float64(s) / float64(PiB)
}

// Exabytes returns the size as a floating-point number of exabytes.
func (s Size) Exabytes() float64 {
	return float64(s) / float64(EB)
}

// Exbibytes returns the size as a floating-point number of exbibytes.
func (s Size) Exbibytes() float64 {
	return float64(s) / float64(EiB)
}
