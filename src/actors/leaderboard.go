package actors

import (
	"github.com/Aqaliarept/leaderboard-game/domain"
	"github.com/asynkron/protoactor-go/actor"
)

type (
	PlayerId   string
	PlayerInfo struct {
		Level   domain.Level
		Country domain.Country
	}
)

type (
	// commands
	Join           struct{ Id PlayerId }
	GetPlayerInfo  struct{ Id PlayerId }
	GetLeaderboard struct{ Id domain.CompetitionId }
	AddScores      struct {
		Id    PlayerId
		Score Score
	}

	// results
	NotFound struct{}
)

type LeaderboardActor struct {
	gatekeeperPid *actor.PID
	players       map[PlayerId]PlayerInfo
	children      map[PlayerId]*actor.PID
}

func NewLeaderboard(gatekeeperPid *actor.PID) *LeaderboardActor {
	return &LeaderboardActor{
		gatekeeperPid,
		map[PlayerId]PlayerInfo{
			"one": {1, "-"},
		}, make(map[PlayerId]*actor.PID)}
}

func (state *LeaderboardActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *Join:
		child, ok := state.children[msg.Id]
		if !ok {
			player, ok := state.players[msg.Id]
			if !ok {
				context.Respond(&NotFound{})
				return
			}
			props := actor.PropsFromProducer(func() actor.Actor { return NewPlayerActor(msg.Id, player.Level, player.Country, state.gatekeeperPid) })
			child = context.Spawn(props)
			state.children[msg.Id] = child
		}
		context.Forward(child)
	case *AddScores:
		child, ok := state.children[msg.Id]
		if !ok {
			context.Respond(&NotFound{})
			return
		}
		context.Forward(child)
	case *GetPlayerInfo:
		child, ok := state.children[msg.Id]
		if !ok {
			context.Respond(&NotFound{})
			return
		}
		context.Forward(child)
	}
}
