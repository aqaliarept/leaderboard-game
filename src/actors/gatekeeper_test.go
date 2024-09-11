package actors

import (
	"fmt"
	"testing"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	t.Run("Dequeue", func(t *testing.T) {
		queue := newQueue()
		require.Equal(t, uint(0), queue.count)

		_, ok := queue.dequeue(1)
		require.False(t, ok)

		queue.Push(slot{PlayerRef{"one", nil}, time.Time{}})
		require.Equal(t, uint(1), queue.count)

		queue.Push(slot{PlayerRef{"two", nil}, time.Time{}})
		require.Equal(t, uint(2), queue.count)

		s, ok := queue.dequeue(1)
		require.True(t, ok)
		require.Equal(t, []slot{
			{PlayerRef{"one", nil}, time.Time{}},
		}, s)
		require.Equal(t, uint(1), queue.Len())

		queue.Push(slot{PlayerRef{"three", nil}, time.Time{}})

		_, ok = queue.dequeue(3)
		require.False(t, ok)

		s, ok = queue.dequeue(2)
		require.True(t, ok)
		require.Equal(t, []slot{
			{PlayerRef{"two", nil}, time.Time{}},
			{PlayerRef{"three", nil}, time.Time{}},
		}, s)
		require.Equal(t, uint(0), queue.Len())

		_, ok = queue.dequeue(10)
		require.False(t, ok)
	})

	t.Run("DequeueStaled", func(t *testing.T) {
		queue := newQueue()
		now := time.Now()
		queue.Push(slot{PlayerRef{"1", nil}, time.Time{}})
		queue.Push(slot{PlayerRef{"2", nil}, time.Time{}})
		queue.Push(slot{PlayerRef{"3", nil}, now})
		res := queue.dequeueStaled(now)
		require.Equal(t, []slot{
			{PlayerRef{"1", nil}, time.Time{}},
			{PlayerRef{"2", nil}, time.Time{}},
		}, res)

		require.Equal(t, uint(1), queue.Len())
	})

	t.Run(`Given a single player far from timeout reached`, func(t *testing.T) {
		queue := newQueue()
		now := time.Now()
		queue.Push(slot{PlayerRef{"1", nil}, now})
		result := queue.GetResult(now)
		require.Empty(t, result.staled)
		require.Empty(t, result.Competitions)
	})

	t.Run(`Given a single player close to deadline (2 sec)`, func(t *testing.T) {
		queue := newQueue()
		now := time.Now()
		queue.Push(slot{PlayerRef{"1", nil}, now.Add(-28 * time.Second)})
		result := queue.GetResult(now)
		require.Empty(t, result.staled)
		require.Empty(t, result.Competitions)
	})

	t.Run(`Given 2 players far from timeout reached`, func(t *testing.T) {
		queue := newQueue()
		now := time.Now()
		queue.Push(slot{PlayerRef{"1", nil}, now.Add(-5 * time.Second)})
		queue.Push(slot{PlayerRef{"2", nil}, now})
		result := queue.GetResult(now)
		require.Empty(t, result.staled)
		require.Empty(t, result.Competitions)
	})

	t.Run(`Given 12 players far from timeout reached
		Then form 1 full group`, func(t *testing.T) {
		queue := newQueue()
		now := time.Now()
		for i := range lo.Range(12) {
			queue.Push(slot{PlayerRef{PlayerId(fmt.Sprintf("%d", i)), nil}, now})
		}
		result := queue.GetResult(now)
		require.Empty(t, result.staled)
		require.Equal(t, 1, len(result.Competitions))
		require.Equal(t, 10, len(result.Competitions[0]))
	})

	t.Run(`Given 2 staled players and 12 players close to timeout
		Then return 2 staled and 2 groups`, func(t *testing.T) {
		queue := newQueue()
		now := time.Now()
		queue.Push(slot{PlayerRef{"s1", nil}, now.Add(-31 * time.Second)})
		queue.Push(slot{PlayerRef{"s2", nil}, now.Add(-30 * time.Second)})
		for i := range lo.Range(12) {
			queue.Push(slot{PlayerRef{PlayerId(fmt.Sprintf("%d", i)), nil}, now.Add(-29 * time.Second)})
		}
		result := queue.GetResult(now)
		require.Equal(t, 2, len(result.staled))
		require.Equal(t, 2, len(result.Competitions))
		require.Equal(t, 10, len(result.Competitions[0]))
		require.Equal(t, 2, len(result.Competitions[1]))
	})

}

type testClock struct {
	now time.Time
}

func (t *testClock) Now() time.Time {
	return t.now
}

func TestGatekeeperActor(t *testing.T) {
	system := actor.NewActorSystem()
	var coordinatorMessages = make([]any, 0)
	coordinator := system.Root.Spawn(actor.PropsFromFunc(func(c actor.Context) {
		coordinatorMessages = append(coordinatorMessages, c.Message())
	}))
	playerMessages := make([]any, 0)
	player := system.Root.Spawn(actor.PropsFromFunc(func(c actor.Context) {
		playerMessages = append(playerMessages, c.Message())
	}))
	start := time.Now()
	clock := &testClock{start}
	gatekeeper := NewGatekeeperActor(coordinator, clock)
	p := actor.PropsFromProducer(func() actor.Actor { return gatekeeper })
	pid := system.Root.Spawn(p)
	system.Root.RequestFuture(pid, &JoinWaiting{player, "one", 10, "-"}, 1*time.Second).Result()
	require.Equal(t, uint(1), gatekeeper.bracket0_10.Len())
	clock.now = start.Add(31 * time.Second)
	system.Root.RequestFuture(pid, &tick{}, 1*time.Second).Result()

	_, _, r := lo.WaitFor(func(_ int) bool {
		return gatekeeper.bracket0_10.Len() == 0
	}, 1*time.Second, 100*time.Millisecond)
	require.True(t, r)

	_, _, r = lo.WaitFor(func(_ int) bool {
		return len(playerMessages) == 1
	}, 1*time.Second, 100*time.Millisecond)
	require.True(t, r)
}
