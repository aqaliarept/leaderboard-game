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
	StartCompetition struct{ Players []PlayerRef }
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
	case StartCompetition:
		competitionId := domain.CompetitionId(uuid.NewString())
		props := actor.PropsFromProducer(func() actor.Actor {
			return NewCompetitionActor(competitionId, msg.Players)
		})
		pid, err := context.SpawnNamed(props, fmt.Sprintf("comp-%s", competitionId))
		if err != nil {
			panic(err)
		}
		state.competitions[competitionId] = pid
	}
}
