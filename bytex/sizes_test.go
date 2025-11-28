package bytex

import (
	"testing"
)

func TestSizeConstants(t *testing.T) {
	tests := []struct {
		name string
		size Size
		want int64
	}{
		{"B", B, 1},
		{"KB", KB, 1000},
		{"KiB", KiB, 1024},
		{"MB", MB, 1000000},
		{"MiB", MiB, 1048576},
		{"GB", GB, 1000000000},
		{"GiB", GiB, 1073741824},
		{"TB", TB, 1000000000000},
		{"TiB", TiB, 1099511627776},
		{"PB", PB, 1000000000000000},
		{"PiB", PiB, 1125899906842624},
		{"EB", EB, 1000000000000000000},
		{"EiB", EiB, 1152921504606846976},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int64(tt.size) != tt.want {
				t.Errorf("Size constant %s = %d, want %d", tt.name, tt.size, tt.want)
			}
		})
	}
}

func TestSizeString(t *testing.T) {
	tests := []struct {
		name string
		size Size
		want string
	}{
		{"Zero", 0, "0B"},
		{"One byte", 1, "1B"},
		{"Negative byte", -1, "-1B"},
		{"1KB", KB, "1KB"},
		{"1KiB", KiB, "1KiB"},
		{"5MB", 5 * MB, "5MB"},
		{"2GiB", 2 * GiB, "2GiB"},
		{"3TB", 3 * TB, "3TB"},
		{"Negative GB", -2 * GB, "-2GB"},
		{"1500 bytes", 1500, "1500B"},
		{"Mixed units", 1536, "1536B"}, // 1.5KiB but not evenly divisible
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.size.String()
			if got != tt.want {
				t.Errorf("Size(%d).String() = %s, want %s", tt.size, got, tt.want)
			}
		})
	}
}

func TestSizeConversions(t *testing.T) {
	size := Size(2 * GiB)

	tests := []struct {
		name   string
		method func() float64
		want   float64
	}{
		{"Bytes", size.Bytes, 2147483648},
		{"Kilobytes", size.Kilobytes, 2147483.648},
		{"Kibibytes", size.Kibibytes, 2097152},
		{"Megabytes", size.Megabytes, 2147.483648},
		{"Mebibytes", size.Mebibytes, 2048},
		{"Gigabytes", size.Gigabytes, 2.147483648},
		{"Gibibytes", size.Gibibytes, 2},
		{"Terabytes", size.Terabytes, 0.002147483648},
		{"Tebibytes", size.Tebibytes, 0.001953125},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.method()
			if got != tt.want {
				t.Errorf("Size(%d).%s() = %f, want %f", size, tt.name, got, tt.want)
			}
		})
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Size
		wantErr bool
	}{
		// Valid inputs
		{"Empty unit means bytes", "100", 100, false},
		{"Explicit bytes", "100B", 100, false},
		{"Kilobytes", "1KB", 1000, false},
		{"Kibibytes", "1KiB", 1024, false},
		{"Megabytes", "5MB", 5000000, false},
		{"Mebibytes", "5MiB", 5242880, false},
		{"Gigabytes", "2GB", 2000000000, false},
		{"Gibibytes", "2GiB", 2147483648, false},
		{"Decimal values", "1.5GB", 1500000000, false},
		{"Decimal KiB", "2.5KiB", 2560, false},
		{"Negative size", "-100MB", -100000000, false},
		{"Positive sign", "+50KB", 50000, false},
		{"Lowercase units", "1gb", 1000000000, false},
		{"Mixed case", "5mB", 5000000, false},
		{"Spaces around unit", "100 KB", 100000, false},
		{"Empty string", "", 0, false}, // Empty string returns 0, no error
		{"Zero with unit", "0GB", 0, false},
		{"Petabytes", "1PB", 1000000000000000, false},
		{"Pebibytes", "1PiB", 1125899906842624, false},
		{"Exabytes", "1EB", 1000000000000000000, false},
		{"Exbibytes", "1EiB", 1152921504606846976, false},

		// Invalid inputs
		{"No number", "GB", 0, true},
		{"Invalid unit", "100XX", 0, true},
		{"Multiple dots", "1.2.3MB", 0, true},
		{"Letters in number", "12a34B", 0, true},
		{"Just sign", "-", 0, true},
		{"Sign without number", "+MB", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSize(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseSize(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestSizeArithmetic(t *testing.T) {
	tests := []struct {
		name string
		op   func() Size
		want Size
	}{
		{"InsertIter sizes", func() Size { return 1*MB + 500*KB }, 1500000},
		{"Subtract sizes", func() Size { return 2*GB - 1*GB }, 1000000000},
		{"Multiply size", func() Size { return 3 * MB }, 3000000},
		{"Divide size", func() Size { return 10 * GB / 5 }, 2000000000},
		{"Mixed binary/decimal", func() Size { return 1*MiB + 1*MB }, 2048576},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.op()
			if got != tt.want {
				t.Errorf("Operation %s = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		size  Size
		check func(t *testing.T)
	}{
		{
			name: "Very large size",
			size: 4 * EiB,
			check: func(t *testing.T) {
				if s := (4 * EiB).String(); s != "4EiB" {
					t.Errorf("Large size string = %s, want 4EiB", s)
				}
			},
		},
		{
			name: "Maximum int64",
			size: Size(9223372036854775807),
			check: func(t *testing.T) {
				// Should not panic
				_ = Size(9223372036854775807).String()
			},
		},
		{
			name: "Negative maximum",
			size: Size(-9223372036854775808),
			check: func(t *testing.T) {
				// Should not panic
				_ = Size(-9223372036854775808).String()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t)
		})
	}
}

func TestHumanReadable(t *testing.T) {
	tests := []struct {
		name string
		size Size
		want string
	}{
		{"Zero", 0, "0B"},
		{"Small bytes", 100, "100.00B"},
		{"1.5 KiB", 1536, "1.50KiB"},
		{"2.5 MiB", Size(2.5 * float64(MiB)), "2.50MiB"},
		{"1.75 GiB", Size(1.75 * float64(GiB)), "1.75GiB"},
		{"Exact 1 GiB", GiB, "1.00GiB"},
		{"Large value", 1500 * GiB, "1.46TiB"},
		{"Very large value", 100 * TiB, "100.00TiB"},
		{"Negative", -2560, "-2.50KiB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.size.HumanReadable()
			if got != tt.want {
				t.Errorf("Size(%d).HumanReadable() = %s, want %s", tt.size, got, tt.want)
			}
		})
	}
}

func TestBytesString(t *testing.T) {
	tests := []struct {
		name string
		size Size
		want string
	}{
		{"Zero", 0, "0B"},
		{"Positive", 1024, "1024B"},
		{"Negative", -500, "-500B"},
		{"Large", 1000000, "1000000B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.size.BytesString()
			if got != tt.want {
				t.Errorf("Size(%d).BytesString() = %s, want %s", tt.size, got, tt.want)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name      string
		size      Size
		unit      Size
		precision int
		want      string
	}{
		{"Auto precision whole", 2 * GB, GB, -1, "2.00GB"},
		{"Auto precision decimal", Size(1.5 * float64(GB)), GB, -1, "1.50GB"},
		{"Fixed precision 0", Size(1.5 * float64(GB)), GB, 0, "2GB"},
		{"Fixed precision 2", Size(1.5 * float64(GB)), GB, 2, "1.50GB"},
		{"Fixed precision 3", Size(1536), KiB, 3, "1.500KiB"},
		{"Cross unit", 1024, KB, -1, "1.02KB"},
		{"Invalid unit", 1000, Size(9999), -1, "1000B (invalid unit: 9999)"},
		{"Binary unit", 2 * MiB, MiB, -1, "2.00MiB"},
		{"Very small value", 100000, GB, 3, "0.000GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.size.Format(tt.unit, tt.precision)
			if got != tt.want {
				t.Errorf("Size(%d).Format(%d, %d) = %s, want %s",
					tt.size, tt.unit, tt.precision, got, tt.want)
			}
		})
	}
}
