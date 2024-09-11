package actors

import (
	"container/list"
	"errors"
	"fmt"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/scheduler"
	"github.com/samber/lo"
)

type (
	tick struct{}
)

type Clock interface {
	Now() time.Time
}

type GatekeeperActor struct {
	competiotionCoordinator *actor.PID
	clock                   Clock
	bracket0_10             *BracketQueue
	bracket11_20            *BracketQueue
	bracket21_30            *BracketQueue
	waitingPlayers          map[PlayerId]bool
	scheduler               *scheduler.TimerScheduler
	cancelTick              scheduler.CancelFunc
}

func NewGatekeeperActor(competiotionCoordinator *actor.PID, clock Clock) *GatekeeperActor {
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
	ref          PlayerRef
	waitingSince time.Time
}

var ErrAlreadyWaiting = errors.New("already waiting")

func (state *GatekeeperActor) enqueuePlayer(slot slot, queue *BracketQueue) error {
	if state.waitingPlayers[slot.ref.Id] {
		return ErrAlreadyWaiting
	}
	state.waitingPlayers[slot.ref.Id] = true
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
		context.Logger().Info("Joined")
		state.cancelTick()
		err := func() error {
			s := slot{PlayerRef{msg.Id, msg.Pid}, state.clock.Now()}
			if msg.Level <= 10 {
				return state.enqueuePlayer(s, state.bracket0_10)
			} else if msg.Level > 10 && msg.Level <= 20 {
				return state.enqueuePlayer(s, state.bracket0_10)
			} else {
				return state.enqueuePlayer(s, state.bracket0_10)
			}
		}()
		if err != nil && !errors.Is(err, ErrAlreadyWaiting) {
			context.Respond(&Error{err.Error()})
			return
		}
		context.Respond(&OK{})
		state.processQueues(context)
		state.scheduleTick(context)
	case *tick:
		context.Logger().Info("TICK")
		state.processQueues(context)
		state.scheduleTick(context)
	}
}

func (state *GatekeeperActor) processQueues(context actor.Context) {
	cleanup := func(q *BracketQueue) {
		res := q.GetResult(state.clock.Now())
		fmt.Printf("RESULT: [%#v]\n", res)
		if len(res.Competitions) > 0 {
			comps := lo.Map(res.Competitions, func(s []slot, _ int) []PlayerRef {
				return lo.Map(s, func(s slot, _ int) PlayerRef {
					return s.ref
				})
			})
			Request(context, state.competiotionCoordinator, &StartCompetition{comps})
		}
		for _, staled := range res.staled {
			fmt.Printf("STALED: [%#v]\n", res)
			Request(context, staled.ref.Pid, &WaitingTimeoutExeeded{})
		}
	}
	cleanup(state.bracket0_10)
	cleanup(state.bracket11_20)
	cleanup(state.bracket21_30)
}

const groupSize = 10
const waitingTimeoutSec = 30
const closeToDeadlineSec = 2

type BracketQueue struct {
	list    *list.List
	count   uint
	timeout time.Duration
}

func newQueue() *BracketQueue {
	return &BracketQueue{list.New(), 0, waitingTimeoutSec * time.Second}
}

func (q *BracketQueue) Len() uint {
	return q.count
}

func (q *BracketQueue) Push(slot slot) {
	q.count++
	q.list.PushBack(slot)
}

func (q *BracketQueue) dequeue(count uint) ([]slot, bool) {
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

func (q *BracketQueue) dequeueStaled(now time.Time) []slot {
	res := make([]slot, 0)
	for {
		el := q.list.Front()
		if el == nil {
			return nil
		}
		s := el.Value.(slot)
		if now.Sub(s.waitingSince) >= q.timeout {
			res = append(res, s)
			q.list.Remove(el)
			q.count--
		} else {
			break
		}
	}
	return res
}

type Result struct {
	Competitions [][]slot
	staled       []slot
}

func (q *BracketQueue) GetResult(now time.Time) Result {
	// cleanup from timeouted players
	staled := q.dequeueStaled(now)
	res := make([][]slot, 0)
	// form full competitions
	for q.Len() >= groupSize {
		s, _ := q.dequeue(groupSize)
		res = append(res, s)
	}
	// allow players close to deadline to play at least with somebody
	if q.Len() < 2 {
		return Result{res, staled}
	}
	front := q.list.Front()
	s := front.Value.(slot)
	if now.Sub(s.waitingSince) >= (waitingTimeoutSec-closeToDeadlineSec)*time.Second {
		s, _ := q.dequeue(q.Len())
		res = append(res, s)
	}
	return Result{res, staled}
}
