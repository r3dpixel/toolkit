package timestamp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSecondsToTime(t *testing.T) {
	s := Seconds(1753935627)
	expected := time.Unix(1753935627, 0)
	assert.Equal(t, expected, s.ToTime())
}

func TestSecondsToNanos(t *testing.T) {
	s := Seconds(1753935627)
	expected := Nano(1753935627000000000)
	assert.Equal(t, expected, s.ToNanos())
}

func TestMilliToTime(t *testing.T) {
	m := Milli(1753935627123)
	expected := time.UnixMilli(1753935627123)
	assert.Equal(t, expected, m.ToTime())
}

func TestMilliToNanos(t *testing.T) {
	m := Milli(1753935627123)
	expected := Nano(1753935627123000000)
	assert.Equal(t, expected, m.ToNanos())
}

func TestMicroToTime(t *testing.T) {
	m := Micro(1753935627123456)
	expected := time.UnixMicro(1753935627123456)
	assert.Equal(t, expected, m.ToTime())
}

func TestMicroToNanos(t *testing.T) {
	m := Micro(1753935627123456)
	expected := Nano(1753935627123456000)
	assert.Equal(t, expected, m.ToNanos())
}

func TestNanoToTime(t *testing.T) {
	n := Nano(1753935627123456789)
	expected := time.Unix(0, 1753935627123456789)
	assert.Equal(t, expected, n.ToTime())
}

func TestNanoToNanos(t *testing.T) {
	n := Nano(1753935627123456789)
	assert.Equal(t, n, n.ToNanos())
}
