package actors

import (
	"github.com/Aqaliarept/leaderboard-game/domain"
	"github.com/asynkron/protoactor-go/actor"
)

type CompetitionActor struct {
	id      domain.CompetitionId
	players []PlayerRef
}

// Receive implements actor.Actor.
func (*CompetitionActor) Receive(c actor.Context) {
	panic("unimplemented")
}

func NewCompetitionActor(id domain.CompetitionId, players []PlayerRef) actor.Actor {
	return &CompetitionActor{id, players}
}
