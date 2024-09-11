package actors

import (
	"log/slog"

	"github.com/Aqaliarept/leaderboard-game/domain"
	"github.com/asynkron/protoactor-go/actor"
)

type (
	Score uint
)

const InvalidState = "invalid state"

type (
	// commands
	JoinCompetition struct {
		CompetitionId domain.CompetitionId
	}
	JoinWaiting struct {
		Pid     *actor.PID
		Id      PlayerId
		Level   domain.Level
		Country domain.Country
	}
	CompleteCompetition   struct{}
	WaitingTimeoutExeeded struct{}

	// results
	OK    struct{}
	Error struct{ message string }
)

type PlayerActor struct {
	id            PlayerId
	behavior      actor.Behavior
	level         domain.Level
	country       domain.Country
	competition   domain.CompetitionId
	gatekeeperPid *actor.PID
}

func NewPlayerActor(id PlayerId, level domain.Level, coutry domain.Country, gatekeeperPid *actor.PID) actor.Actor {
	a := &PlayerActor{id, actor.NewBehavior(), level, coutry, domain.CompetitionIdUnset, gatekeeperPid}
	a.behavior.Become(a.Idle)
	return a
}

func respondError(context actor.Context) {
	context.Respond(&Error{InvalidState})
	context.Logger().Error("player", slog.String("state", "error"))
}

func (state *PlayerActor) Idle(context actor.Context) {
	switch context.Message().(type) {
	case *Join:
		// report to bucket router
		Request(context, state.gatekeeperPid, &JoinWaiting{context.Self(), state.id, state.level, state.country})
		state.behavior.Become(state.Waiting)
		context.Logger().Info("player", slog.String("state", "waiting"))
		context.Respond(&OK{})
	case *JoinCompetition:
		respondError(context)
	case *WaitingTimeoutExeeded:
		respondError(context)
	case *AddScores:
		respondError(context)
	case *CompleteCompetition:
		respondError(context)
	}
}

func (state *PlayerActor) Waiting(context actor.Context) {
	switch msg := context.Message().(type) {
	case *Join:
		respondError(context)
	case *JoinCompetition:
		state.competition = msg.CompetitionId
		state.behavior.Become(state.Playing)
		context.Logger().Info("player", slog.String("state", "playing"))
		context.Respond(&OK{})
	case *WaitingTimeoutExeeded:
		state.behavior.Become(state.Idle)
		context.Logger().Info("player", slog.String("state", "idle"))
		context.Respond(&OK{})
	case *AddScores:
		respondError(context)
	case *CompleteCompetition:
		respondError(context)
	}
}

func (state *PlayerActor) Playing(context actor.Context) {
	switch context.Message().(type) {
	case *Join:
		respondError(context)
	case *JoinCompetition:
		respondError(context)
	case *AddScores:
		// report to comp router
		context.Logger().Info("player ", slog.String("state", "scores"))
		context.Respond(&OK{})
	case *CompleteCompetition:
		state.behavior.Become(state.Idle)
		context.Logger().Info("player ", slog.String("state", "idle"))
		context.Respond(&OK{})
	}
}

func (state *PlayerActor) Receive(context actor.Context) {
	state.behavior.Receive(context)
}
