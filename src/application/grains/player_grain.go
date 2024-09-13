package grains

import (
	"errors"
	"fmt"

	"github.com/Aqaliarept/leaderboard-game/application"
	"github.com/Aqaliarept/leaderboard-game/domain"
	"github.com/Aqaliarept/leaderboard-game/domain/player"
	generated "github.com/Aqaliarept/leaderboard-game/generated/cluster"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/gr1nd3rz/go-fast-ddd/core"
)

type PlayerGrain struct {
	clock  Clock
	player *player.Player
	repo   application.PlayersRepo
}

type PlayerGrainFactory struct {
	clock Clock
	application.PlayersRepo
}

func NewPlayerGrainFactory(clock Clock, repo application.PlayersRepo) *PlayerGrainFactory {
	return &PlayerGrainFactory{clock, repo}
}

func (f *PlayerGrainFactory) New() generated.Player {
	return &PlayerGrain{f.clock, nil, f.PlayersRepo}
}

// Init implements Hello.
func (state *PlayerGrain) Init(ctx cluster.GrainContext) {
	id := player.PlayerId(ctx.Identity())
	ctx.Logger().Info("PLAYER CREATED", "id", id)
	pi := state.repo.Get(id)
	level := player.Level(1)
	if pi != nil {
		level = pi.Level
	}
	state.player = player.New(core.AggregateId(id), level, "-")
}

// ReceiveDefault implements Hello.
func (g *PlayerGrain) ReceiveDefault(ctx cluster.GrainContext) {
}

// Terminate implements Hello.
func (g *PlayerGrain) Terminate(ctx cluster.GrainContext) {
}

// Hello implements Hello.
func (state *PlayerGrain) Join(req *generated.JoinRequest, ctx cluster.GrainContext) (*generated.None, error) {
	ctx.Logger().Info("JOIN", "id", ctx.Identity())
	pack, err := state.player.Join(state.clock.Now())
	if errors.Is(err, player.ErrAlreadyPlaying) {
		return none, generated.ErrPlayerAlreadyPlaying(err.Error())
	}
	if err != nil {
		return none, err
	}
	_, err = domain.EventOfType[player.WaitingStarted](pack)
	if errors.Is(err, domain.ErrNotFound) {
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
	ctx.Logger().Info("ADD SCORES", "id", ctx.Identity(), "scores", req.Scores)
	pack, err := state.player.AddScores(player.Scores(req.Scores))
	if errors.Is(err, player.ErrNotPlaying) {
		return none, generated.ErrPlayerNotPlaying(err.Error())
	}
	if err != nil {
		return none, err
	}
	e, err := domain.EventOfType[player.ScoresAdded](pack)
	if errors.Is(err, domain.ErrNotFound) {
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
	ctx.Logger().Info("COMPLETE COMPETITION", "id", ctx.Identity())
	_, err := state.player.CompleteCompetition()
	return none, err
}

// StartCompetition implements cluster.Player.
func (state *PlayerGrain) StartCompetition(req *generated.StartCompetitionRequest, ctx cluster.GrainContext) (*generated.None, error) {
	ctx.Logger().Info("JOIN COMPETITION", "id", ctx.Identity(), "competition_id", req.Id)
	_, err := state.player.StartCompetition(player.CompetitionId(req.Id))
	return none, err
}

// WaitingExpired implements cluster.Player.
func (state *PlayerGrain) WaitingExpired(req *generated.None, ctx cluster.GrainContext) (*generated.None, error) {
	ctx.Logger().Info("WAITING EXPIRED", "id", ctx.Identity())
	_, err := state.player.WaitingExpired()
	if err != nil {
		return none, err
	}
	return none, nil
}
