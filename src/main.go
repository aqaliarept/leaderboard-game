package main

import (
	"context"
	"log"

	httpapi "github.com/Aqaliarept/leaderboard-game/adapters/in/http_api"
	"github.com/Aqaliarept/leaderboard-game/application/grains"
	"github.com/Aqaliarept/leaderboard-game/application/services"
	"github.com/Aqaliarept/leaderboard-game/generated/server/restapi"
	"github.com/asynkron/protoactor-go/cluster"
	"go.uber.org/fx"
)

func configureContainer() fx.Option {
	return fx.Options(
		fx.Provide(NewClock),
		fx.Provide(grains.NewPlayerGrainFactory),
		fx.Provide(NewCluster),
		fx.Provide(NewWebServer),
		fx.Provide(httpapi.NewLeaderboardApiImpl),
		fx.Provide(services.NewLeaderboardService),
		fx.Invoke(registerHooks),
	)
}

func main() {
	app := fx.New(
		configureContainer(),
	)
	app.Run()
}

func registerHooks(lifecycle fx.Lifecycle, cluster *cluster.Cluster, server *restapi.Server) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				cluster.StartMember()
				go func() {
					err := server.Serve()
					if err != nil {
						panic(err)
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				cluster.Shutdown(true)
				err := server.Shutdown()
				if err != nil {
					log.Default().Printf("http shutdown error: %s", err.Error())
				}
				return nil
			},
		},
	)
}
