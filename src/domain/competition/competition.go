package competition

import (
	"errors"
	"fmt"

	"github.com/Aqaliarept/leaderboard-game/domain/player"
	"github.com/gr1nd3rz/go-fast-ddd/core"
	"github.com/zavitax/sortedset-go"
)

var _ core.AggregateState = CompetitionState{}

type CompetitionState struct {
	competed    bool
	leaderboard *sortedset.SortedSet[player.PlayerId, player.Scores, player.PlayerId]
	players     []player.PlayerId
}

// events
type Created struct {
	Players []player.PlayerId
}

type ScoresUpdated struct {
	Player player.PlayerId
	Scores player.Scores
}

type Completed struct {
	Players []player.PlayerId
}

var ErrNoSuchPlayer = errors.New("no such player")

// Apply implements core.AggregateState.
func (c CompetitionState) Apply(event core.Event) core.AggregateState {
	switch e := event.(type) {
	case Created:
		c.leaderboard = sortedset.New[player.PlayerId, player.Scores, player.PlayerId]()
		for _, p := range e.Players {
			c.leaderboard.AddOrUpdate(p, 0, p)
		}
		c.players = e.Players
		return c
	case ScoresUpdated:
		c.leaderboard.AddOrUpdate(e.Player, e.Scores, e.Player)
		return c
	case Completed:
		return c
	default:
		panic(fmt.Errorf("%w event [%T]", errors.ErrUnsupported, event))
	}
}

type Competition struct {
	core.Aggregate[CompetitionState]
}

func New(id core.AggregateId, players []player.PlayerId) *Competition {
	c := &Competition{}
	c.Initialize(id, Created{players})
	return c
}

func (c *Competition) ReportScores(player player.PlayerId, scores player.Scores) (core.EventPack, error) {
	node := c.State().leaderboard.GetByKey(player)
	if node == nil {
		return nil, ErrNoSuchPlayer
	}
	return c.ProcessCommand(func(cs *CompetitionState, er core.EventRaiser) error {
		if cs.competed {
			return nil
		}
		node := cs.leaderboard.GetByKey(player)
		// use negative scores for proper sorting
		er.RaiseTrue(!cs.competed, ScoresUpdated{player, node.Score() - scores})
		return nil
	})
}

func (c *Competition) Complete() (core.EventPack, error) {
	return c.ProcessCommand(func(cs *CompetitionState, er core.EventRaiser) error {
		er.RaiseTrue(!cs.competed, Completed{cs.players})
		return nil
	})
}
