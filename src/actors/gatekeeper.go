package actors

import (
	"container/list"
	"errors"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/scheduler"
)

type (
	tick struct{}
)

const waitingTimeout = 30

type Clock interface {
	Now() time.Time
}

type GatekeeperActor struct {
	competiotionCoordinator *actor.PID
	clock                   Clock
	queue0_10               *queue
	queue11_20              *queue
	queue21_30              *queue
	waitingPlayers          map[PlayerId]bool
	scheduler               *scheduler.TimerScheduler
	cancelTick              scheduler.CancelFunc
}

func NewGatekeeperActor(competiotionCoordinator *actor.PID, clock Clock) actor.Actor {
	return &GatekeeperActor{
		competiotionCoordinator,
		clock,
		newQueue(),
		newQueue(),
		newQueue(),
		make(map[PlayerId]bool),
		nil,
		nil,
	}
}

type slot struct {
	id           PlayerId
	pid          *actor.PID
	waitingSince time.Time
}

var ErrAlreadyWaiting = errors.New("already waiting")

func (state *GatekeeperActor) enqueuePlayer(slot slot, queue *queue) error {
	if state.waitingPlayers[slot.id] {
		return ErrAlreadyWaiting
	}
	state.waitingPlayers[slot.id] = true
	queue.Push(slot)
	return nil
}

func (state *GatekeeperActor) scheduleTick(context actor.Context) {
	state.cancelTick = state.scheduler.SendOnce(2*time.Second, context.Self(), &tick{})
}

// Receive implements actor.Actor.
func (state *GatekeeperActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		state.scheduler = scheduler.NewTimerScheduler(context)
		state.scheduleTick(context)
	case *JoinWaiting:
		state.cancelTick()
		err := func() error {
			s := slot{msg.Id, msg.Pid, state.clock.Now()}
			if msg.Level <= 10 {
				return state.enqueuePlayer(s, state.queue0_10)
			} else if msg.Level > 10 && msg.Level <= 20 {
				return state.enqueuePlayer(s, state.queue0_10)
			} else {
				return state.enqueuePlayer(s, state.queue0_10)
			}
		}()
		if errors.Is(err, ErrAlreadyWaiting) {
			context.Respond(&OK{})
		} else if err != nil {
			context.Respond(&Error{err.Error()})
		}
		state.Cleanup(context)
		state.scheduleTick(context)
	case *tick:
		context.Logger().Info("TICK")
		state.scheduleTick(context)
	}

}

func (state *GatekeeperActor) Cleanup(context actor.Context) {
	cleanup := func(q *queue) {
		for _, v := range q.DequeueStaled(state.clock.Now(), waitingTimeout) {
			req, err := context.RequestFuture(v.pid, &WaitingTimeoutExeeded{}, 1*time.Second).Result()
			if err != nil {
				context.Logger().Error("sending WaitingTimeoutExeeded", "msg", err.Error())
			}
			e, ok := req.(Error)
			if ok {
				context.Logger().Error("sending WaitingTimeoutExeeded", "msg", e.message)
			}
		}
	}
	cleanup(state.queue0_10)
	cleanup(state.queue11_20)
	cleanup(state.queue21_30)
}

type queue struct {
	list  *list.List
	count uint
}

func newQueue() *queue {
	return &queue{list.New(), 0}
}

func (q *queue) Len() uint {
	return q.count
}

func (q *queue) Push(slot slot) {
	q.count++
	q.list.PushBack(slot)
}

func (q *queue) Dequeue(count uint) ([]slot, bool) {
	if q.count < count {
		return nil, false
	}
	q.count -= count
	res := make([]slot, count)
	for i := uint(0); i < count; i++ {
		el := q.list.Front()
		res[i] = el.Value.(slot)
		q.list.Remove(el)
	}
	return res, true
}

func (q *queue) DequeueStaled(now time.Time, timeout time.Duration) []slot {
	res := make([]slot, 0)
	for {
		el := q.list.Front()
		s := el.Value.(slot)
		if now.Sub(s.waitingSince) > timeout {
			res = append(res, s)
			q.list.Remove(el)
			q.count--
		} else {
			break
		}
	}
	return res
}
