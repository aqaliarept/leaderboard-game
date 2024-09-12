package grains

import (
	"errors"
	"time"

	generated "github.com/Aqaliarept/leaderboard-game/generated/cluster"
	"github.com/google/go-cmp/cmp"
	"github.com/gr1nd3rz/go-fast-ddd/core"
)

type Clock interface {
	Now() time.Time
}

type clock struct {
}

func (clock) Now() time.Time {
	return time.Now()
}

var none = &generated.None{}

func HasEvent(pack core.EventPack, event any) bool {
	for _, e := range pack {
		if cmp.Equal(e, event) {
			return true
		}
	}
	return false
}

var ErrNotFound = errors.New("not found")
var ErrTooMany = errors.New("too many events")

func EventOfType[T core.Event](pack core.EventPack) (T, error) {
	e := EventsOfType[T](pack)
	var evt T
	if len(e) == 0 {
		return evt, ErrNotFound
	} else if len(e) > 1 {
		return evt, ErrTooMany
	} else {
		return e[0], nil
	}
}

func EventsOfType[T core.Event](pack core.EventPack) []T {
	res := make([]T, 0)
	for _, e := range pack {
		switch evt := e.(type) {
		case T:
			res = append(res, evt)
		}
	}
	return res
}

func NoEvents(pack core.EventPack) bool {
	return len(pack) == 0
}
