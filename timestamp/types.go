package timestamp

import "time"

type Timestamp interface {
	~int64
	ToTime() time.Time
	ToNanos() Nano
}

type Seconds int64 // type representing a timestamp in seconds

// ToTime converts Seconds timestamp to time.Time
func (s Seconds) ToTime() time.Time { return time.Unix(int64(s), 0) }

// ToNanos converts Seconds to nanoseconds
func (s Seconds) ToNanos() Nano { return Nano(s) * 1_000_000_000 }

type Milli int64 // type representing a timestamp in milliseconds

// ToTime converts Milli timestamp to time.Time
func (m Milli) ToTime() time.Time { return time.UnixMilli(int64(m)) }

// ToNanos converts Milli to nanoseconds
func (m Milli) ToNanos() Nano { return Nano(m) * 1_000_000 }

type Micro int64 // // type representing a timestamp in microseconds

// ToTime converts Micro timestamp to time.Time
func (m Micro) ToTime() time.Time { return time.UnixMicro(int64(m)) }

// ToNanos converts Micro to nanoseconds
func (m Micro) ToNanos() Nano { return Nano(m) * 1_000 }

type Nano int64 // // type representing a timestamp in nanoseconds

// ToTime converts Nano timestamp to time.Time
func (n Nano) ToTime() time.Time { return time.Unix(0, int64(n)) }

// ToNanos converts Nano to nanoseconds
func (n Nano) ToNanos() Nano { return n }
