package domain

import (
	"container/list"
	"time"
)

type slot struct {
	playerId     PlayerId
	waitingSince time.Time
}

type BracketQueue struct {
	list            *list.List
	count           uint
	groupSize       uint
	minGroupSize    uint
	waitingTimeout  time.Duration
	closeToDeadline time.Duration
}

func NewQueue(desiredGroupSize uint, minGroupSize uint, waitingTimeout time.Duration, closeToDeadline time.Duration) *BracketQueue {
	return &BracketQueue{list.New(), 0, desiredGroupSize, minGroupSize, waitingTimeout, closeToDeadline}
}

func (q *BracketQueue) Len() uint {
	return q.count
}

func (q *BracketQueue) Push(player PlayerId, now time.Time) {
	q.count++
	q.list.PushBack(slot{player, now})
}

func (q *BracketQueue) dequeue(count uint) ([]PlayerId, bool) {
	if q.count < count {
		return nil, false
	}
	q.count -= count
	res := make([]PlayerId, count)
	for i := uint(0); i < count; i++ {
		el := q.list.Front()
		res[i] = el.Value.(slot).playerId
		q.list.Remove(el)
	}
	return res, true
}

func (q *BracketQueue) dequeueStaled(now time.Time) []PlayerId {
	res := make([]PlayerId, 0)
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
	Competitions [][]PlayerId
	Staled       []PlayerId
}

func (q *BracketQueue) Next(now time.Time) Result {
	// cleanup from timeouted players
	staled := q.dequeueStaled(now)
	res := make([][]PlayerId, 0)
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
	// fmt.Printf("diff: %s closeToDeadline: %s\n", now.Sub(s.waitingSince), q.waitingTimeout-q.closeToDeadline)
	if now.Sub(s.waitingSince) >= (q.waitingTimeout - q.closeToDeadline) {
		s, _ := q.dequeue(q.Len())
		res = append(res, s)
	}
	return Result{res, staled}
}
