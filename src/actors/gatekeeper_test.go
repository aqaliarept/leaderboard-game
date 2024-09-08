package actors

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	t.Run("Dequeue", func(t *testing.T) {
		queue := newQueue()
		require.Equal(t, uint(0), queue.count)

		_, ok := queue.Dequeue(1)
		require.False(t, ok)

		queue.Push(slot{"one", nil, time.Time{}})
		require.Equal(t, uint(1), queue.count)

		queue.Push(slot{"two", nil, time.Time{}})
		require.Equal(t, uint(2), queue.count)

		s, ok := queue.Dequeue(1)
		require.True(t, ok)
		require.Equal(t, []slot{
			{"one", nil, time.Time{}},
		}, s)
		require.Equal(t, uint(1), queue.Len())

		queue.Push(slot{"three", nil, time.Time{}})

		_, ok = queue.Dequeue(3)
		require.False(t, ok)

		s, ok = queue.Dequeue(2)
		require.True(t, ok)
		require.Equal(t, []slot{
			{"two", nil, time.Time{}},
			{"three", nil, time.Time{}},
		}, s)
		require.Equal(t, uint(0), queue.Len())

		_, ok = queue.Dequeue(10)
		require.False(t, ok)
	})
	t.Run("DequeueStaled", func(t *testing.T) {
		queue := newQueue()
		now := time.Now()
		queue.Push(slot{"1", nil, time.Time{}})
		queue.Push(slot{"2", nil, time.Time{}})
		queue.Push(slot{"3", nil, now})
		res := queue.DequeueStaled(now, 20*time.Second)
		require.Equal(t, []slot{
			{"1", nil, time.Time{}},
			{"2", nil, time.Time{}},
		}, res)

		require.Equal(t, uint(1), queue.Len())
	})
}
