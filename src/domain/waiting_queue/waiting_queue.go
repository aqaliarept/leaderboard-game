package waiting_queue

import (
	"container/list"
	"fmt"
	"time"

	"github.com/Aqaliarept/leaderboard-game/domain/player"
)

type WaitingQueue struct {
	players      map[player.PlayerId]bool
	bracket0_10  *bracketQueue
	bracket11_20 *bracketQueue
	bracket21_30 *bracketQueue
}

func NewWaitingQueue(desiredGroupSize uint, minGroupSize uint, waitingTimeout time.Duration, closeToDeadline time.Duration) *WaitingQueue {
	return &WaitingQueue{
		make(map[player.PlayerId]bool),
		&bracketQueue{list.New(), 0, desiredGroupSize, minGroupSize, waitingTimeout, closeToDeadline},
		&bracketQueue{list.New(), 0, desiredGroupSize, minGroupSize, waitingTimeout, closeToDeadline},
		&bracketQueue{list.New(), 0, desiredGroupSize, minGroupSize, waitingTimeout, closeToDeadline},
	}
}

func (q *WaitingQueue) Push(player player.PlayerId, level player.Level, now time.Time) {
	// checking for idemptency
	if q.players[player] {
		return
	}
	q.players[player] = true
	if level <= 10 {
		q.bracket0_10.Push(player, now)
	} else if level <= 20 {
		q.bracket11_20.Push(player, now)
	} else {
		q.bracket21_30.Push(player, now)
	}
}

func (q *WaitingQueue) Len() uint {
	return q.bracket0_10.Len() + q.bracket11_20.Len() + q.bracket21_30.Len()
}

func (q *WaitingQueue) Next(now time.Time) Result {
	clean := func(res *Result) {
		for _, s := range res.Competitions {
			for _, p := range s {
				delete(q.players, p)
			}
		}
		for _, staled := range res.Staled {
			delete(q.players, staled)
		}
	}
	res1 := q.bracket0_10.Next(now)
	res2 := q.bracket11_20.Next(now)
	res3 := q.bracket21_30.Next(now)
	clean(&res1)
	clean(&res2)
	clean(&res3)
	return Result{
		append(append(res1.Competitions, res2.Competitions...), res3.Competitions...),
		append(append(res1.Staled, res2.Staled...), res3.Staled...),
	}
}

type slot struct {
	playerId     player.PlayerId
	waitingSince time.Time
}

type bracketQueue struct {
	list            *list.List
	count           uint
	groupSize       uint
	minGroupSize    uint
	waitingTimeout  time.Duration
	closeToDeadline time.Duration
}

func newBracketQueue(desiredGroupSize uint, minGroupSize uint, waitingTimeout time.Duration, closeToDeadline time.Duration) *bracketQueue {
	return &bracketQueue{list.New(), 0, desiredGroupSize, minGroupSize, waitingTimeout, closeToDeadline}
}

func (q *bracketQueue) Len() uint {
	return q.count
}

func (q *bracketQueue) Push(player player.PlayerId, now time.Time) {
	q.count++
	q.list.PushBack(slot{player, now})
}

func (q *bracketQueue) dequeue(count uint) ([]player.PlayerId, bool) {
	if q.count < count {
		return nil, false
	}
	q.count -= count
	res := make([]player.PlayerId, count)
	for i := uint(0); i < count; i++ {
		el := q.list.Front()
		res[i] = el.Value.(slot).playerId
		q.list.Remove(el)
	}
	return res, true
}

func (q *bracketQueue) dequeueStaled(now time.Time) []player.PlayerId {
	res := make([]player.PlayerId, 0)
	for {
		el := q.list.Front()
		if el == nil {
			break
		}
		s := el.Value.(slot)
		if now.Sub(s.waitingSince) >= q.waitingTimeout {
			res = append(res, s.playerId)
			q.list.Remove(el)
			q.count--
		} else {
			break
		}
	}
	return res
}

type Result struct {
	Competitions [][]player.PlayerId
	Staled       []player.PlayerId
}

func (q *bracketQueue) Next(now time.Time) Result {
	// cleanup from timeouted players
	staled := q.dequeueStaled(now)
	res := make([][]player.PlayerId, 0)
	// form full competitions
	for q.Len() >= uint(q.groupSize) {
		s, _ := q.dequeue(q.groupSize)
		res = append(res, s)
	}
	// allow players close to deadline to play at least with somebody
	if q.Len() < q.minGroupSize {
		return Result{res, staled}
	}
	front := q.list.Front()
	s := front.Value.(slot)
	fmt.Printf("diff: %s closeToDeadline: %s\n", now.Sub(s.waitingSince), q.waitingTimeout-q.closeToDeadline)
	if now.Sub(s.waitingSince) >= (q.waitingTimeout - q.closeToDeadline) {
		s, _ := q.dequeue(q.Len())
		res = append(res, s)
	}
	result := Result{res, staled}
	fmt.Printf("players queue: [%d] result: %#v\n", q.Len(), result)
	return result
}
