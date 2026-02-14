package timestamp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fromDateTestCase struct {
	name        string
	format      string
	dateString  string
	expectError bool
	expectedSec Seconds
	expectedMil Milli
	expectedMic Micro
	expectedNan Nano
}

type toDateTestCase struct {
	name                 string
	timestamp            Nano
	format               string
	expectedUTCSeconds   string
	expectedUTCMilli     string
	expectedUTCMicro     string
	expectedUTCNano      string
	expectedLocalSeconds string
	expectedLocalMilli   string
	expectedLocalMicro   string
	expectedLocalNano    string
}

var fromDateTests = []fromDateTestCase{
	{
		name:        "RFC3339Nano UTC",
		format:      time.RFC3339Nano,
		dateString:  "2025-07-31T04:20:27.123456789Z",
		expectError: false,
		expectedSec: 1753935627,
		expectedMil: 1753935627123,
		expectedMic: 1753935627123456,
		expectedNan: 1753935627123456789,
	},
	{
		name:        "RFC3339Nano With Timezone",
		format:      time.RFC3339Nano,
		dateString:  "2025-07-31T07:20:27.123456789+03:00", // Same instant in time as above
		expectError: false,
		expectedSec: 1753935627,
		expectedMil: 1753935627123,
		expectedMic: 1753935627123456,
		expectedNan: 1753935627123456789,
	},
	{
		name:        "Custom Display Format (No sub-seconds)",
		format:      DisplayFormat,
		dateString:  "31/07/2025,  04:20:27",
		expectError: false,
		expectedSec: 1753935627,
		expectedMil: 1753935627000,
		expectedMic: 1753935627000000,
		expectedNan: 1753935627000000000,
	},
	{
		name:        "Invalid Date String",
		format:      time.RFC3339Nano,
		dateString:  "not a valid date",
		expectError: true,
	},
}

var toDateTests = []toDateTestCase{
	{
		name:      "RFC3339Nano Format",
		timestamp: 1753935627123456789,
		format:    time.RFC3339Nano,

		expectedUTCSeconds: "2025-07-31T04:20:27Z",
		expectedUTCMilli:   "2025-07-31T04:20:27.123Z",
		expectedUTCMicro:   "2025-07-31T04:20:27.123456Z",
		expectedUTCNano:    "2025-07-31T04:20:27.123456789Z",

		expectedLocalSeconds: "2025-07-31T07:20:27+03:00",
		expectedLocalMilli:   "2025-07-31T07:20:27.123+03:00",
		expectedLocalMicro:   "2025-07-31T07:20:27.123456+03:00",
		expectedLocalNano:    "2025-07-31T07:20:27.123456789+03:00",
	},
	{
		name:      "Display Format",
		timestamp: 1753935627123456789,
		format:    DisplayFormat,

		expectedUTCSeconds: "31/07/2025, 04:20:27",
		expectedUTCMilli:   "31/07/2025, 04:20:27",
		expectedUTCMicro:   "31/07/2025, 04:20:27",
		expectedUTCNano:    "31/07/2025, 04:20:27",

		expectedLocalSeconds: "31/07/2025, 07:20:27",
		expectedLocalMilli:   "31/07/2025, 07:20:27",
		expectedLocalMicro:   "31/07/2025, 07:20:27",
		expectedLocalNano:    "31/07/2025, 07:20:27",
	},
}

func TestParse(t *testing.T) {
	t.Run("Valid date", func(t *testing.T) {
		result := Parse(time.RFC3339, "2025-07-31T04:20:27Z")
		assert.Equal(t, Nano(1753935627000000000), result)
	})

	t.Run("Invalid date returns zero", func(t *testing.T) {
		result := Parse(time.RFC3339, "invalid-date")
		assert.Equal(t, Nano(0), result)
	})
}

func TestParseErr(t *testing.T) {
	for _, tc := range fromDateTests {
		t.Run(tc.name, func(t *testing.T) {
			nan, errNan := ParseErr(tc.format, tc.dateString)

			if tc.expectError {
				assert.Error(t, errNan)
			} else {
				assert.NoError(t, errNan)
				assert.Equal(t, tc.expectedNan, nan, "Nano output mismatch")

				// Test conversions from parsed Nano
				sec := ConvertToSeconds(nan)
				mil := ConvertToMilli(nan)
				mic := ConvertToMicro(nan)

				assert.Equal(t, tc.expectedSec, sec, "Seconds conversion mismatch")
				assert.Equal(t, tc.expectedMil, mil, "Milli conversion mismatch")
				assert.Equal(t, tc.expectedMic, mic, "Micro conversion mismatch")
			}
		})
	}
}

func TestFormatAndFormatLocal(t *testing.T) {
	t.Setenv("TZ", "Europe/Bucharest")

	for _, tc := range toDateTests {
		t.Run(tc.name, func(t *testing.T) {
			tsSec := ConvertToSeconds(tc.timestamp)
			tsMil := ConvertToMilli(tc.timestamp)
			tsMic := ConvertToMicro(tc.timestamp)
			tsNan := tc.timestamp

			assert.Equal(t, tc.expectedUTCSeconds, Format(tsSec, tc.format), "Format Seconds mismatch")
			assert.Equal(t, tc.expectedUTCMilli, Format(tsMil, tc.format), "Format Milli mismatch")
			assert.Equal(t, tc.expectedUTCMicro, Format(tsMic, tc.format), "Format Micro mismatch")
			assert.Equal(t, tc.expectedUTCNano, Format(tsNan, tc.format), "Format Nano mismatch")

			assert.Equal(t, tc.expectedLocalSeconds, FormatLocal(tsSec, tc.format), "FormatLocal Seconds mismatch")
			assert.Equal(t, tc.expectedLocalMilli, FormatLocal(tsMil, tc.format), "FormatLocal Milli mismatch")
			assert.Equal(t, tc.expectedLocalMicro, FormatLocal(tsMic, tc.format), "FormatLocal Micro mismatch")
			assert.Equal(t, tc.expectedLocalNano, FormatLocal(tsNan, tc.format), "FormatLocal Nano mismatch")
		})
	}
}

func TestNowFunctions(t *testing.T) {
	delta := int64(100)

	t.Run("NowSeconds", func(t *testing.T) {
		now := time.Now().Unix()
		result := NowSeconds()
		assert.InDelta(t, now, int64(result), 1)
	})

	t.Run("NowMilli", func(t *testing.T) {
		now := time.Now().UnixMilli()
		result := NowMilli()
		assert.InDelta(t, now, int64(result), float64(delta))
	})

	t.Run("NowMicro", func(t *testing.T) {
		now := time.Now().UnixMicro()
		result := NowMicro()
		assert.InDelta(t, now, int64(result), float64(delta*1000))
	})

	t.Run("NowNano", func(t *testing.T) {
		now := time.Now().UnixNano()
		result := NowNano()
		assert.InDelta(t, now, int64(result), float64(delta*1000*1000))
	})
}

func TestConversions(t *testing.T) {
	baseNano := Nano(1753935627123456789)
	baseMicro := ConvertToMicro(baseNano)
	baseMilli := ConvertToMilli(baseNano)
	baseSec := ConvertToSeconds(baseNano)

	t.Run("ConvertToMilli from Seconds", func(t *testing.T) {
		result := ConvertToMilli(baseSec)
		assert.Equal(t, Milli(1753935627000), result)
	})

	t.Run("ConvertToMicro from Seconds", func(t *testing.T) {
		result := ConvertToMicro(baseSec)
		assert.Equal(t, Micro(1753935627000000), result)
	})

	t.Run("ConvertToNano from Seconds", func(t *testing.T) {
		result := ConvertToNano(baseSec)
		assert.Equal(t, Nano(1753935627000000000), result)
	})

	t.Run("ConvertToSeconds from Milli", func(t *testing.T) {
		result := ConvertToSeconds(baseMilli)
		assert.Equal(t, baseSec, result)
	})

	t.Run("ConvertToMicro from Milli", func(t *testing.T) {
		result := ConvertToMicro(baseMilli)
		assert.Equal(t, Micro(1753935627123000), result)
	})

	t.Run("ConvertToNano from Milli", func(t *testing.T) {
		result := ConvertToNano(baseMilli)
		assert.Equal(t, Nano(1753935627123000000), result)
	})

	t.Run("ConvertToSeconds from Micro", func(t *testing.T) {
		result := ConvertToSeconds(baseMicro)
		assert.Equal(t, baseSec, result)
	})

	t.Run("ConvertToMilli from Micro", func(t *testing.T) {
		result := ConvertToMilli(baseMicro)
		assert.Equal(t, baseMilli, result)
	})

	t.Run("ConvertToNano from Micro", func(t *testing.T) {
		result := ConvertToNano(baseMicro)
		assert.Equal(t, Nano(1753935627123456000), result)
	})

	t.Run("ConvertToSeconds from Nano", func(t *testing.T) {
		result := ConvertToSeconds(baseNano)
		assert.Equal(t, baseSec, result)
	})

	t.Run("ConvertToMilli from Nano", func(t *testing.T) {
		result := ConvertToMilli(baseNano)
		assert.Equal(t, baseMilli, result)
	})

	t.Run("ConvertToMicro from Nano", func(t *testing.T) {
		result := ConvertToMicro(baseNano)
		assert.Equal(t, baseMicro, result)
	})
}
