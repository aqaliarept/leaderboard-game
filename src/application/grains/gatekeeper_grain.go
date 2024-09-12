package grains

import (
	"fmt"
	"time"

	"github.com/Aqaliarept/leaderboard-game/domain"
	generated "github.com/Aqaliarept/leaderboard-game/generated/cluster"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/scheduler"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

type tick struct{}

type GatekeeperGrain struct {
	queue     *domain.BracketQueue
	clock     clock
	scheduler *scheduler.TimerScheduler
}

// Enqueue implements cluster.Gatekeeper.
func (state *GatekeeperGrain) Enqueue(req *generated.EnqueueRequest, ctx cluster.GrainContext) (*generated.None, error) {
	ctx.Logger().Info("ENQUEUE", "id", req.PlayerId, "level", req.Level)
	state.queue.Push(domain.PlayerId(req.PlayerId), state.clock.Now())
	return none, nil
}

// Init implements cluster.Gatekeeper.
func (state *GatekeeperGrain) Init(ctx cluster.GrainContext) {
	state.queue = domain.NewQueue(10, 1, 5*time.Second, 2*time.Second)
	state.clock = clock{}
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
				players := lo.Map(comp, func(id domain.PlayerId, _ int) string {
					return string(id)
				})
				client := generated.GetCompetitionGrainClient(ctx.Cluster(), uuid.NewString())
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

func NewGatekeeper() generated.Gatekeeper {
	return &GatekeeperGrain{}
}
