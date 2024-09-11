package grains

import (
	"errors"
	"fmt"
	"time"

	generated "github.com/Aqaliarept/leaderboard-game/cluster"
	"github.com/Aqaliarept/leaderboard-game/domain"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/scheduler"
	"github.com/gr1nd3rz/go-fast-ddd/core"
	"github.com/samber/lo"
)

type CompetitionGrain struct {
	scheduler   *scheduler.TimerScheduler
	competition *domain.Competition
}

func NewCompetitionGrain() generated.Competition {
	return &CompetitionGrain{}
}

// AddScores implements cluster.Competition.
func (state *CompetitionGrain) AddScores(req *generated.AddPlayerScoresRequest, ctx cluster.GrainContext) (*generated.None, error) {
	ctx.Logger().Info("ADD SCORES")
	_, err := state.competition.ReportScores(domain.PlayerId(req.PlayerId), domain.Scores(req.Scrores))
	if err != nil {
		return none, err
	}
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
		e, err := EventOfType[domain.Completed](pack)
		if errors.Is(err, ErrNotFound) {
			return
		}
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
	state.competition = domain.NewCompetition(core.AggregateId(id),
		lo.Map(req.Players, func(id string, _ int) domain.PlayerId {
			return domain.PlayerId(id)
		}))
	state.scheduler = scheduler.NewTimerScheduler(ctx)
	state.scheduler.SendOnce(5*time.Second, ctx.Self(), &tick{})
	for _, player := range req.Players {
		client := generated.GetPlayerGrainClient(ctx.Cluster(), player)
		_, err := client.StartCompetition(&generated.StartCompetitionRequest{Id: id})
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("player error: %s", err.Error()))
		}
	}
	return none, nil
}
