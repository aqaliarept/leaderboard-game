package waiting_queue

import (
	"fmt"
	"testing"
	"time"

	"github.com/Aqaliarept/leaderboard-game/domain/player"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func newTestBracketQueue() *bracketQueue {
	return newBracketQueue(10, 2, 5*time.Second, 2*time.Second)
}
func newTestBracketQueueMinSize(minSize uint) *bracketQueue {
	return newBracketQueue(10, minSize, 5*time.Second, 2*time.Second)
}

func newTestWaitingQueueMinSize(groupSize uint, minSize uint) *WaitingQueue {
	return NewWaitingQueue(groupSize, minSize, 5*time.Second, 2*time.Second)
}

func TestBracketQueue(t *testing.T) {
	t.Run("Dequeue", func(t *testing.T) {
		queue := newTestBracketQueue()
		require.Equal(t, uint(0), queue.count)

		_, ok := queue.dequeue(1)
		require.False(t, ok)

		queue.Push("one", time.Time{})
		require.Equal(t, uint(1), queue.count)

		queue.Push("two", time.Time{})
		require.Equal(t, uint(2), queue.count)

		s, ok := queue.dequeue(1)
		require.True(t, ok)
		require.Equal(t, []player.PlayerId{"one"}, s)
		require.Equal(t, uint(1), queue.Len())

		queue.Push("three", time.Time{})

		_, ok = queue.dequeue(3)
		require.False(t, ok)

		s, ok = queue.dequeue(2)
		require.True(t, ok)
		require.Equal(t, []player.PlayerId{
			"two",
			"three",
		}, s)
		require.Equal(t, uint(0), queue.Len())

		_, ok = queue.dequeue(10)
		require.False(t, ok)
	})

	t.Run("DequeueStaled", func(t *testing.T) {
		queue := newTestBracketQueue()
		now := time.Now()
		queue.Push("1", time.Time{})
		queue.Push("2", time.Time{})
		queue.Push("3", now)
		res := queue.dequeueStaled(now)
		require.Equal(t, []player.PlayerId{
			"1",
			"2",
		}, res)
		require.Equal(t, uint(1), queue.Len())
	})

	t.Run(`Given a single player far from timeout reached`, func(t *testing.T) {
		queue := newTestBracketQueue()
		now := time.Now()
		queue.Push("1", now)
		result := queue.Next(now)
		require.Empty(t, result.Staled)
		require.Empty(t, result.Competitions)
	})

	t.Run(`Given a single player close to deadline (2 sec)`, func(t *testing.T) {
		queue := newTestBracketQueue()
		now := time.Now()
		queue.Push("1", now.Add(-4*time.Second))
		result := queue.Next(now)
		require.Empty(t, result.Staled)
		require.Empty(t, result.Competitions)
	})

	t.Run(`Given a queue with min size of one
		And a single player close to deadline (2 sec)`, func(t *testing.T) {
		queue := newTestBracketQueueMinSize(1)
		now := time.Now()
		queue.Push("1", now.Add(-4*time.Second))
		result := queue.Next(now)
		require.Empty(t, result.Staled)
		require.Equal(t, 1, len(result.Competitions))
	})

	t.Run(`Given 2 players far from timeout reached`, func(t *testing.T) {
		queue := newTestBracketQueue()
		now := time.Now()
		queue.Push("1", now.Add(-1*time.Second))
		queue.Push("2", now)
		result := queue.Next(now)
		require.Empty(t, result.Staled)
		require.Empty(t, result.Competitions)
	})

	t.Run(`Given 12 players far from timeout reached
		Then form 1 full group`, func(t *testing.T) {
		queue := newTestBracketQueue()
		now := time.Now()
		for i := range lo.Range(12) {
			queue.Push(player.PlayerId(fmt.Sprintf("%d", i)), now)
		}
		result := queue.Next(now)
		require.Empty(t, result.Staled)
		require.Equal(t, 1, len(result.Competitions))
		require.Equal(t, 10, len(result.Competitions[0]))
	})

	t.Run(`Given 2 staled players and 12 players close to timeout
		Then return 2 staled and 2 groups`, func(t *testing.T) {
		queue := newTestBracketQueue()
		now := time.Now()
		queue.Push("s1", now.Add(-6*time.Second))
		queue.Push("s2", now.Add(-5*time.Second))
		for i := range lo.Range(12) {
			queue.Push(player.PlayerId(fmt.Sprintf("%d", i)), now.Add(-4*time.Second))
		}
		result := queue.Next(now)
		require.Equal(t, []player.PlayerId{"s1", "s2"}, result.Staled)
		require.Equal(t, 2, len(result.Competitions))
		require.Equal(t, 10, len(result.Competitions[0]))
		require.Equal(t, 2, len(result.Competitions[1]))
	})

	t.Run(`Give MinSize = 1
		When 10 players and 1 player close to timeout
		Then returns 2 groups with 10 and 1 players`, func(t *testing.T) {
		queue := newTestBracketQueueMinSize(1)
		now := time.Now()
		for i := range lo.Range(10) {
			queue.Push(player.PlayerId(fmt.Sprintf("%d", i)), now.Add(-4*time.Second))
		}
		queue.Push("s1", now.Add(-4*time.Second))

		result := queue.Next(now)
		require.Empty(t, result.Staled)
		require.Equal(t, 2, len(result.Competitions))
		require.Equal(t, 10, len(result.Competitions[0]))
		require.Equal(t, 1, len(result.Competitions[1]))
	})

	t.Run(`Given a single player with min count 1`, func(t *testing.T) {
		queue := newBracketQueue(10, 2, 5*time.Second, 2*time.Second)
		now := time.Now()
		queue.Push("s1", now.Add(-6*time.Second))
		result := queue.Next(now)
		require.Equal(t, []player.PlayerId{"s1"}, result.Staled)
	})
}

func TestWaitingQeueu(t *testing.T) {
	t.Run(`When add players with various levels
		Then should form competitions in respect with brackets`, func(t *testing.T) {
		queue := newTestWaitingQueueMinSize(3, 1)
		now := time.Now()
		queue.Push("10", 10, now.Add(-5*time.Second))
		queue.Push("20", 20, now.Add(-5*time.Second))
		queue.Push("30", 30, now.Add(-5*time.Second))
		queue.Push("1", 1, now)
		queue.Push("11", 11, now)
		queue.Push("2", 2, now)
		queue.Push("12", 12, now)
		queue.Push("3", 3, now)
		queue.Push("13", 13, now)
		queue.Push("21", 21, now)
		queue.Push("22", 22, now)
		queue.Push("23", 23, now)

		require.Equal(t, uint(12), queue.Len())

		result := queue.Next(now)
		require.Equal(t, uint(0), queue.Len())

		require.Equal(t, 3, len(result.Competitions))
		require.Equal(t, 3, len(lo.Intersect([]player.PlayerId{"1", "2", "3"}, result.Competitions[0])))
		require.Equal(t, 3, len(lo.Intersect([]player.PlayerId{"11", "12", "13"}, result.Competitions[1])))
		require.Equal(t, 3, len(lo.Intersect([]player.PlayerId{"21", "22", "23"}, result.Competitions[2])))
		require.Equal(t, []player.PlayerId{"10", "20", "30"}, result.Staled)

	})

	t.Run(`When same player pushed to queue more than once with different levels
		Then this ignored (idempotence scenario)
		And player can be pushed again after he left the queue`, func(t *testing.T) {
		queue := newTestWaitingQueueMinSize(3, 1)
		now := time.Now()
		queue.Push("1", 1, now.Add(-4*time.Second))
		queue.Push("10", 10, now.Add(-4*time.Second))
		require.Equal(t, uint(2), queue.Len())
		queue.Push("10", 1, now.Add(-4*time.Second))
		queue.Push("10", 22, now.Add(-4*time.Second))
		require.Equal(t, uint(2), queue.Len())
		result := queue.Next(now)
		require.Equal(t, 2, len(result.Competitions[0]))
		require.Equal(t, uint(0), queue.Len())
		queue.Push("10", 10, now.Add(-4*time.Second))
		require.Equal(t, uint(1), queue.Len())
	})
}
