package actors

import (
	"fmt"
	"time"

	"github.com/Aqaliarept/leaderboard-game/domain"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/scheduler"
)

type CompetitionCompleted struct {
	Id domain.CompetitionId
}

type CompetitionActor struct {
	id         domain.CompetitionId
	players    []PlayerRef
	scheduler  *scheduler.TimerScheduler
	cancelTick scheduler.CancelFunc
}

// Receive implements actor.Actor.
func (state *CompetitionActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		context.Logger().Info(fmt.Sprintf("%#v", state.players))
		state.scheduler = scheduler.NewTimerScheduler(context)
		state.scheduler.SendOnce(5*time.Second, context.Self(), &CompetitionCompleted{state.id})
	case *CompetitionCompleted:
		Request(context, context.Parent(), msg)
		context.Stop(context.Self())
	}
}

func NewCompetitionActor(id domain.CompetitionId, players []PlayerRef) actor.Actor {
	return &CompetitionActor{id, players, nil, nil}
}
