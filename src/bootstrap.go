package main

import (
	"log"
	"os"
	"time"

	httpapi "github.com/Aqaliarept/leaderboard-game/adapters/in/http_api"
	"github.com/Aqaliarept/leaderboard-game/application/grains"
	generated "github.com/Aqaliarept/leaderboard-game/generated/cluster"
	"github.com/Aqaliarept/leaderboard-game/generated/server/restapi"
	"github.com/Aqaliarept/leaderboard-game/generated/server/restapi/operations"
	actor "github.com/asynkron/protoactor-go/actor"
	cluster "github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/test"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/go-openapi/loads"
	"github.com/jessevdk/go-flags"
)

type clock struct {
}

func (clock) Now() time.Time {
	return time.Now()
}

func NewClock() grains.Clock {
	return &clock{}
}

func NewCluster(playerFactory *grains.PlayerGrainFactory) *cluster.Cluster {
	system := actor.NewActorSystem()
	provider := test.NewTestProvider(test.NewInMemAgent())
	lookup := disthash.New()
	config := remote.Configure("localhost", 0)
	playerKind := generated.NewPlayerKind(playerFactory.New, 0)
	compKind := generated.NewCompetitionKind(grains.NewCompetitionGrain, 0)
	gatekeeperKind := generated.NewGatekeeperKind(grains.NewGatekeeper, 0)
	clusterConfig := cluster.Configure("test", provider, lookup, config, cluster.WithKinds(playerKind, compKind, gatekeeperKind))
	return cluster.New(system, clusterConfig)
}

func NewWebServer(apiImpl *httpapi.LeaderboardApiImpl) *restapi.Server {
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		panic(err)
	}
	api := operations.NewLeaderboardAPIAPI(swaggerSpec)
	api.UseSwaggerUI()
	apiImpl.SetupHandlers(api)

	server := restapi.NewServer(api)
	server.SetHandler(api.Serve(nil))

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Leaderboard API"
	parser.LongDescription = "API for managing leaderboard entries and scores."
	server.ConfigureFlags()
	for _, optsGroup := range api.CommandLineOptionsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			log.Fatalln(err)
		}
	}
	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}
	return server
}
