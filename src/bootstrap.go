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
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/k8s"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/test"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/go-openapi/loads"
	"github.com/jessevdk/go-flags"
	"k8s.io/utils/env"
)

type clock struct {
}

func (clock) Now() time.Time {
	return time.Now()
}

func NewClock() grains.Clock {
	return &clock{}
}

func getHostInformation() (host string, port int, advertisedHost string) {
	host = env.GetString("PROTOHOST", "127.0.0.1")
	port, err := env.GetInt("PROTOPORT", 0)
	if err != nil {
		log.Panic(err)
	}
	advertisedHost = env.GetString("PROTOADVERTISEDHOST", "")
	log.Printf("host: %s, port: %d, advertisedHost: %s", host, port, advertisedHost)
	return
}

func getClusterProvider() (cluster.ClusterProvider, *remote.Config) {
	if os.Getenv("KUBE_ENV") != "" {
		host, port, advertisedHost := getHostInformation()
		config := remote.Configure(host, port, remote.WithAdvertisedHost(advertisedHost))
		provider, err := k8s.New()
		if err != nil {
			log.Panic(err)
		}
		return provider, config
	} else {
		config := remote.Configure("localhost", 0)
		provider := test.NewTestProvider(test.NewInMemAgent())
		return provider, config
	}
}

func NewCluster(
	playerFactory *grains.PlayerGrainFactory,
	competitionFactory *grains.CompetitionGrainFactory,
) *cluster.Cluster {
	system := actor.NewActorSystem()
	lookup := disthash.New()
	playerKind := generated.NewPlayerKind(playerFactory.New, 0)
	compKind := generated.NewCompetitionKind(competitionFactory.New, 0)
	gatekeeperKind := generated.NewGatekeeperKind(grains.NewGatekeeper, 0)
	provider, config := getClusterProvider()
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
