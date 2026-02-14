package timestamp

import (
	"time"

	"github.com/rs/zerolog/log"
)

const (
	DisplayFormat string = "02/01/2006, 15:04:05"
)

// Parse parses a date string using the given format and returns a timestamp of the specified type
func Parse(format, date string) Nano {
	return ParseM(format, date, nil)
}

// ParseF parses a date string using the given format and returns a timestamp of the specified type (attaches fields to log)
func ParseF(format, date string, fields ...any) Nano {
	t, err := ParseErr(format, date)
	if err != nil {
		log.Error().Fields(fields).Err(err).Msgf("Could not parse date %s using format %s", date, format)
	}

	return t
}

// ParseM parses a date string using the given format and returns a timestamp of the specified type (attaches fields to log)
func ParseM(format, date string, fields map[string]any) Nano {
	t, err := ParseErr(format, date)
	if err != nil {
		log.Error().Fields(fields).Err(err).Msgf("Could not parse date %s using format %s", date, format)
	}

	return t
}

// ParseErr parses a date string using the given format and returns a timestamp of the specified type
func ParseErr(format, date string) (Nano, error) {
	t, err := time.Parse(format, date)
	if err != nil {
		return 0, err
	}

	return Nano(t.UnixNano()), nil
}

// Format formats a timestamp as a string using the given format in UTC
func Format[T Timestamp](ts T, format string) string {
	return ts.ToTime().UTC().Format(format)
}

// FormatLocal formats a timestamp as a string using the given format in local time
func FormatLocal[T Timestamp](ts T, format string) string {
	return ts.ToTime().Local().Format(format)
}

// NowSeconds returns the current time as Seconds
func NowSeconds() Seconds {
	return Seconds(time.Now().Unix())
}

// NowMilli returns the current time as Milli
func NowMilli() Milli {
	return Milli(time.Now().UnixMilli())
}

// NowMicro returns the current time as Micro
func NowMicro() Micro {
	return Micro(time.Now().UnixMicro())
}

// NowNano returns the current time as Nano
func NowNano() Nano {
	return Nano(time.Now().UnixNano())
}

// ConvertToSeconds converts any Timestamp to Seconds
func ConvertToSeconds[T Timestamp](from T) Seconds {
	return Seconds(from.ToNanos() / 1_000_000_000)
}

// ConvertToMilli converts any Timestamp to Milli
func ConvertToMilli[T Timestamp](from T) Milli {
	return Milli(from.ToNanos() / 1_000_000)
}

// ConvertToMicro converts any Timestamp to Micro
func ConvertToMicro[T Timestamp](from T) Micro {
	return Micro(from.ToNanos() / 1_000)
}

// ConvertToNano converts any Timestamp to Nano
func ConvertToNano[T Timestamp](from T) Nano {
	return from.ToNanos()
}
