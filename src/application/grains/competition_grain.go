package grains

import (
	"errors"
	"fmt"
	"time"

	"github.com/Aqaliarept/leaderboard-game/application"
	"github.com/Aqaliarept/leaderboard-game/domain"
	"github.com/Aqaliarept/leaderboard-game/domain/competition"
	"github.com/Aqaliarept/leaderboard-game/domain/player"

	generated "github.com/Aqaliarept/leaderboard-game/generated/cluster"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/scheduler"
	"github.com/gr1nd3rz/go-fast-ddd/core"
	"github.com/samber/lo"
)

var duration = time.Duration(1 * time.Minute)

type CompetitionGrainFactory struct {
	clock   Clock
	storage application.LeaderBoardStorage
}

func NewCompetitionGrainFactory(clock Clock, storage application.LeaderBoardStorage) *CompetitionGrainFactory {
	return &CompetitionGrainFactory{clock, storage}
}

func (f *CompetitionGrainFactory) New() generated.Competition {
	return &CompetitionGrain{f.clock, f.storage, nil, nil}
}

type CompetitionGrain struct {
	clock       Clock
	storage     application.LeaderBoardStorage
	scheduler   *scheduler.TimerScheduler
	competition *competition.Competition
}

// AddScores implements cluster.Competition.
func (state *CompetitionGrain) AddScores(req *generated.AddPlayerScoresRequest, ctx cluster.GrainContext) (*generated.None, error) {
	ctx.Logger().Info("ADD SCORES")
	_, err := state.competition.ReportScores(player.PlayerId(req.PlayerId), player.Scores(req.Scrores))
	if err != nil {
		return none, err
	}
	state.updateReadModel()
	return none, nil
}

// Init implements cluster.Competition.
func (state *CompetitionGrain) Init(ctx cluster.GrainContext) {
	id := ctx.Identity()
	ctx.Logger().Info("COMPETITION CREATED", "id", id)
}

// ReceiveDefault implements cluster.Competition.
func (state *CompetitionGrain) ReceiveDefault(ctx cluster.GrainContext) {
	switch ctx.Message().(type) {
	case *tick:
		// ctx.Logger().Info("COMPETITION TICK")
		pack, err := state.competition.Complete()
		if err != nil {
			ctx.Logger().Error(err.Error())
		}
		e, err := domain.EventOfType[competition.Completed](pack)
		if errors.Is(err, domain.ErrNotFound) {
			return
		}
		state.updateReadModel()
		for _, player := range e.Players {
			client := generated.GetPlayerGrainClient(ctx.Cluster(), string(player))
			_, err := client.CompleteCompetition(none)
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("player error: %s", err.Error()))
			}
		}
	}
}

// Terminate implements cluster.Competition.
func (c *CompetitionGrain) Terminate(ctx cluster.GrainContext) {
}

// Start implements cluster.Competition.
func (state *CompetitionGrain) Start(req *generated.StartRequest, ctx cluster.GrainContext) (*generated.None, error) {
	ctx.Logger().Info("COMPETITION STARTED")
	id := ctx.Identity()
	state.competition = competition.New(core.AggregateId(id),
		lo.Map(req.Players, func(id string, _ int) player.PlayerId {
			return player.PlayerId(id)
		}), state.clock.Now(), duration)
	state.updateReadModel()
	state.scheduler = scheduler.NewTimerScheduler(ctx)
	state.scheduler.SendOnce(duration, ctx.Self(), &tick{})
	for _, player := range req.Players {
		client := generated.GetPlayerGrainClient(ctx.Cluster(), player)
		_, err := client.StartCompetition(&generated.StartCompetitionRequest{Id: id})
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("player error: %s", err.Error()))
		}
	}
	return none, nil
}

func (state *CompetitionGrain) updateReadModel() {
	state.storage.Save(state.competition.GetInfo())
}
