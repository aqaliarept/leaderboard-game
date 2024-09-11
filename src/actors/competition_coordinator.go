package actors

import (
	"fmt"

	"github.com/Aqaliarept/leaderboard-game/domain"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/google/uuid"
)

type (
	PlayerRef struct {
		Id  PlayerId
		Pid *actor.PID
	}
	StartCompetition struct{ Competitions [][]PlayerRef }
)

type CompetitionCoordinatorActor struct {
	competitions map[domain.CompetitionId]*actor.PID
}

func NewCompetitionCoordinatorActor() actor.Actor {
	return &CompetitionCoordinatorActor{make(map[domain.CompetitionId]*actor.PID)}
}

// Receive implements actor.Actor.
func (state *CompetitionCoordinatorActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *StartCompetition:
		context.Logger().Info("Start competition")
		for _, c := range msg.Competitions {
			competitionId := domain.CompetitionId(uuid.NewString())
			props := actor.PropsFromProducer(func() actor.Actor {
				return NewCompetitionActor(competitionId, c)
			})
			pid, err := context.SpawnNamed(props, fmt.Sprintf("comp-%s", competitionId))
			if err != nil {
				panic(err)
			}
			state.competitions[competitionId] = pid
		}
		context.Respond(&OK{})
	case *CompetitionCompleted:
		context.Logger().Info(fmt.Sprintf("Completed [%s]", msg.Id))
		delete(state.competitions, msg.Id)
		context.Respond(&OK{})
	}
}
