package async

import "context"

// IsCancelled returns the cancelled status of the context at the current moment of time (non-blocking)
func IsCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
