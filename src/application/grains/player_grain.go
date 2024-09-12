package grains

import (
	"errors"
	"fmt"

	"github.com/Aqaliarept/leaderboard-game/domain"
	generated "github.com/Aqaliarept/leaderboard-game/generated/cluster"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/gr1nd3rz/go-fast-ddd/core"
)

type PlayerGrain struct {
	clock  Clock
	player *domain.Player
}

type PlayerGrainFactory struct {
	clock Clock
}

func NewPlayerGrainFactory(clock Clock) *PlayerGrainFactory {
	return &PlayerGrainFactory{clock}
}

func (f *PlayerGrainFactory) New() generated.Player {
	return &PlayerGrain{f.clock, nil}
}

// Init implements Hello.
func (state *PlayerGrain) Init(ctx cluster.GrainContext) {
	id := ctx.Identity()
	ctx.Logger().Info("PLAYER CREATED", "id", id)
	state.player = domain.NewPlayer(core.AggregateId(id), 1, "-")
}

// ReceiveDefault implements Hello.
func (g *PlayerGrain) ReceiveDefault(ctx cluster.GrainContext) {
}

// Terminate implements Hello.
func (g *PlayerGrain) Terminate(ctx cluster.GrainContext) {
}

// Hello implements Hello.
func (state *PlayerGrain) Join(req *generated.JoinRequest, ctx cluster.GrainContext) (*generated.None, error) {
	ctx.Logger().Info("JOIN")
	pack, err := state.player.Join(state.clock.Now())
	if err != nil {
		return none, err
	}

	_, err = EventOfType[domain.WaitingStarted](pack)
	if errors.Is(err, ErrNotFound) {
		return none, nil
	}
	client := generated.GetGatekeeperGrainClient(ctx.Cluster(), "gatekeeper")
	_, err = client.Enqueue(&generated.EnqueueRequest{
		PlayerId: string(state.player.Id()),
		Level:    int32(state.player.State().Level),
	})
	if err != nil {
		return none, fmt.Errorf("gatekeeper error: %w", err)
	}
	return none, nil
}

// AddScores implements cluster.Player.
func (state *PlayerGrain) AddScores(req *generated.AddScoresRequest, ctx cluster.GrainContext) (*generated.None, error) {
	ctx.Logger().Info("ADD SCORES")
	pack, err := state.player.AddScores(domain.Scores(req.Scores))
	if err != nil {
		return none, err
	}
	e, err := EventOfType[domain.ScoresAdded](pack)
	if errors.Is(err, ErrNotFound) {
		return none, nil
	}
	client := generated.GetCompetitionGrainClient(ctx.Cluster(), string(e.Competition))
	_, err = client.AddScores(&generated.AddPlayerScoresRequest{
		PlayerId: string(state.player.Id()),
		Scrores:  int32(e.Scores),
	})
	if err != nil {
		return none, fmt.Errorf("competition [%s] error: %w", e.Competition, err)
	}
	return none, nil
}

// CompleteCompetition implements cluster.Player.
func (state *PlayerGrain) CompleteCompetition(req *generated.None, ctx cluster.GrainContext) (*generated.None, error) {
	ctx.Logger().Info("COMPLETE COMPETITION")
	_, err := state.player.CompleteCompetition()
	return none, err
}

// StartCompetition implements cluster.Player.
func (state *PlayerGrain) StartCompetition(req *generated.StartCompetitionRequest, ctx cluster.GrainContext) (*generated.None, error) {
	ctx.Logger().Info("START COMPETITION", "id", req.Id)
	_, err := state.player.StartCompetition(domain.CompetitionId(req.Id))
	return none, err
}

// WaitingExpired implements cluster.Player.
func (state *PlayerGrain) WaitingExpired(req *generated.None, ctx cluster.GrainContext) (*generated.None, error) {
	ctx.Logger().Info("WAITING EXPIRED")
	_, err := state.player.WaitingExpired()
	if err != nil {
		return none, err
	}
	return none, nil
}
