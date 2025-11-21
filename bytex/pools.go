package bytex

import "sync"

// Buffer32k pool to reduce memory pressure when handling large amount of file operations (32KB buffers)
var Buffer32k = sync.Pool{
	New: func() any {
		return make([]byte, 32*KiB)
	},
}
