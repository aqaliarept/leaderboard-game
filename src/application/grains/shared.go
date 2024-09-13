package grains

import (
	"time"

	generated "github.com/Aqaliarept/leaderboard-game/generated/cluster"
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
