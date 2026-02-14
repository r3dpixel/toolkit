package async

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsCancelled(t *testing.T) {
	t.Run("Context Not Cancelled", func(t *testing.T) {
		ctx := context.Background()

		assert.False(t, IsCancelled(ctx))
	})

	t.Run("Context Is Cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		cancel()

		assert.True(t, IsCancelled(ctx))
	})

	t.Run("Context Cancelled After Timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		time.Sleep(20 * time.Millisecond)

		assert.True(t, IsCancelled(ctx))
	})
}
