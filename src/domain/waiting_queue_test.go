package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func newQueue() *BracketQueue {
	return NewQueue(10, 2, 5*time.Second, 2*time.Second)
}
func newQueueMinSize(minSize uint) *BracketQueue {
	return NewQueue(10, minSize, 5*time.Second, 2*time.Second)
}

func TestQueue(t *testing.T) {
	t.Run("Dequeue", func(t *testing.T) {
		queue := newQueue()
		require.Equal(t, uint(0), queue.count)

		_, ok := queue.dequeue(1)
		require.False(t, ok)

		queue.Push("one", time.Time{})
		require.Equal(t, uint(1), queue.count)

		queue.Push("two", time.Time{})
		require.Equal(t, uint(2), queue.count)

		s, ok := queue.dequeue(1)
		require.True(t, ok)
		require.Equal(t, []PlayerId{"one"}, s)
		require.Equal(t, uint(1), queue.Len())

		queue.Push("three", time.Time{})

		_, ok = queue.dequeue(3)
		require.False(t, ok)

		s, ok = queue.dequeue(2)
		require.True(t, ok)
		require.Equal(t, []PlayerId{
			"two",
			"three",
		}, s)
		require.Equal(t, uint(0), queue.Len())

		_, ok = queue.dequeue(10)
		require.False(t, ok)
	})

	t.Run("DequeueStaled", func(t *testing.T) {
		queue := newQueue()
		now := time.Now()
		queue.Push("1", time.Time{})
		queue.Push("2", time.Time{})
		queue.Push("3", now)
		res := queue.dequeueStaled(now)
		require.Equal(t, []PlayerId{
			"1",
			"2",
		}, res)
		require.Equal(t, uint(1), queue.Len())
	})

	t.Run(`Given a single player far from timeout reached`, func(t *testing.T) {
		queue := newQueue()
		now := time.Now()
		queue.Push("1", now)
		result := queue.Next(now)
		require.Empty(t, result.Staled)
		require.Empty(t, result.Competitions)
	})

	t.Run(`Given a single player close to deadline (2 sec)`, func(t *testing.T) {
		queue := newQueue()
		now := time.Now()
		queue.Push("1", now.Add(-4*time.Second))
		result := queue.Next(now)
		require.Empty(t, result.Staled)
		require.Empty(t, result.Competitions)
	})

	t.Run(`Given a queue with min size of one
		And a single player close to deadline (2 sec)`, func(t *testing.T) {
		queue := newQueueMinSize(1)
		now := time.Now()
		queue.Push("1", now.Add(-4*time.Second))
		result := queue.Next(now)
		require.Empty(t, result.Staled)
		require.Equal(t, 1, len(result.Competitions))
	})

	t.Run(`Given 2 players far from timeout reached`, func(t *testing.T) {
		queue := newQueue()
		now := time.Now()
		queue.Push("1", now.Add(-1*time.Second))
		queue.Push("2", now)
		result := queue.Next(now)
		require.Empty(t, result.Staled)
		require.Empty(t, result.Competitions)
	})

	t.Run(`Given 12 players far from timeout reached
		Then form 1 full group`, func(t *testing.T) {
		queue := newQueue()
		now := time.Now()
		for i := range lo.Range(12) {
			queue.Push(PlayerId(fmt.Sprintf("%d", i)), now)
		}
		result := queue.Next(now)
		require.Empty(t, result.Staled)
		require.Equal(t, 1, len(result.Competitions))
		require.Equal(t, 10, len(result.Competitions[0]))
	})

	t.Run(`Given 2 staled players and 12 players close to timeout
		Then return 2 staled and 2 groups`, func(t *testing.T) {
		queue := newQueue()
		now := time.Now()
		queue.Push("s1", now.Add(-6*time.Second))
		queue.Push("s2", now.Add(-5*time.Second))
		for i := range lo.Range(12) {
			queue.Push(PlayerId(fmt.Sprintf("%d", i)), now.Add(-4*time.Second))
		}
		result := queue.Next(now)
		require.Equal(t, []PlayerId{"s1", "s2"}, result.Staled)
		require.Equal(t, 2, len(result.Competitions))
		require.Equal(t, 10, len(result.Competitions[0]))
		require.Equal(t, 2, len(result.Competitions[1]))
	})

	t.Run(`Given a single player with min count 1`, func(t *testing.T) {
		queue := NewQueue(10, 2, 5*time.Second, 2*time.Second)
		now := time.Now()
		queue.Push("s1", now.Add(-6*time.Second))
		result := queue.Next(now)
		require.Equal(t, []PlayerId{"s1"}, result.Staled)
	})
}
