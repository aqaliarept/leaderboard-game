package grains

import (
	"fmt"
	"time"

	"github.com/Aqaliarept/leaderboard-game/application"
	"github.com/Aqaliarept/leaderboard-game/domain/player"
	"github.com/Aqaliarept/leaderboard-game/domain/waiting_queue"
	generated "github.com/Aqaliarept/leaderboard-game/generated/cluster"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/scheduler"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

type tick struct{}

type GatekeeperGrain struct {
	config    *application.Config
	queue     *waiting_queue.WaitingQueue
	clock     Clock
	scheduler *scheduler.TimerScheduler
}

type GatekeeperFactory struct {
	clock  Clock
	config *application.Config
}

func NewGatekeeperFactory(clock Clock, config *application.Config) *GatekeeperFactory {
	return &GatekeeperFactory{clock, config}
}

func (f *GatekeeperFactory) New() generated.Gatekeeper {
	return &GatekeeperGrain{
		f.config,
		waiting_queue.NewWaitingQueue(
			f.config.CompetitionSize,
			f.config.MinCompetitionSize,
			f.config.QueueWaitingTimeout,
			2*time.Second),
		f.clock, nil,
	}
}

// Enqueue implements cluster.Gatekeeper.
func (state *GatekeeperGrain) Enqueue(req *generated.EnqueueRequest, ctx cluster.GrainContext) (*generated.None, error) {
	ctx.Logger().Info("ENQUEUE", "id", req.PlayerId, "level", req.Level)
	state.queue.Push(player.PlayerId(req.PlayerId), player.Level(req.Level), state.clock.Now())
	return none, nil
}

// Init implements cluster.Gatekeeper.
func (state *GatekeeperGrain) Init(ctx cluster.GrainContext) {
	state.scheduler = scheduler.NewTimerScheduler(ctx)
	state.scheduler.SendRepeatedly(1*time.Second, 1*time.Second, ctx.Self(), &tick{})
	ctx.Logger().Info("GATEKEEPER CREATED")
}

// ReceiveDefault implements cluster.Gatekeeper.
func (state *GatekeeperGrain) ReceiveDefault(ctx cluster.GrainContext) {
	switch ctx.Message().(type) {
	case *tick:
		// ctx.Logger().Info("GATEKEPER TICK")
		result := state.queue.Next(state.clock.Now())
		if len(result.Competitions) > 0 {
			ctx.Logger().Info("COMPETIONS", "count", len(result.Competitions))
			for _, comp := range result.Competitions {
				players := lo.Map(comp, func(id player.PlayerId, _ int) string {
					return string(id)
				})
				competitionId := uuid.NewString()
				client := generated.GetCompetitionGrainClient(ctx.Cluster(), competitionId)
				ctx.Logger().Info("INITIATE COMPETITION", "id", competitionId, "players", players)
				_, err := client.Start(&generated.StartRequest{
					Players: players,
				})
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("competition error: %s", err.Error()))
				}
			}
		}
		if len(result.Staled) > 0 {
			ctx.Logger().Info("STALED", "count", len(result.Staled))
			for _, staledPlayer := range result.Staled {
				client := generated.GetPlayerGrainClient(ctx.Cluster(), string(staledPlayer))
				_, err := client.WaitingExpired(none)
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("player error: %s", err.Error()))
				}
			}
		}
	}
}

// Terminate implements cluster.Gatekeeper.
func (g *GatekeeperGrain) Terminate(ctx cluster.GrainContext) {
}
