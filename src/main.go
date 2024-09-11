package main

import (
	"fmt"
	"time"

	generated "github.com/Aqaliarept/leaderboard-game/cluster"
	"github.com/Aqaliarept/leaderboard-game/grains"

	console "github.com/asynkron/goconsole"
	actor "github.com/asynkron/protoactor-go/actor"
	cluster "github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/test"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
)

type clock struct {
}

func (clock) Now() time.Time {
	return time.Now()
}

func main() {
	system := actor.NewActorSystem()
	provider := test.NewTestProvider(test.NewInMemAgent())
	lookup := disthash.New()
	config := remote.Configure("localhost", 0)
	playerKind := generated.NewPlayerKind(grains.NewPlayerGrain, 0)
	compKind := generated.NewCompetitionKind(grains.NewCompetitionGrain, 0)
	gatekeeperKind := generated.NewGatekeeperKind(grains.NewGatekeeper, 0)
	clusterConfig := cluster.Configure("test", provider, lookup, config, cluster.WithKinds(playerKind, compKind, gatekeeperKind))
	cst := cluster.New(system, clusterConfig)
	cst.StartMember()
	client := generated.GetPlayerGrainClient(cst, "test")
	_, err := client.Join(&generated.JoinRequest{Name: "player-1"})
	if err != nil {
		if generated.IsUserNotFound(err) {
			fmt.Println("user not found")
		} else {
			fmt.Printf("unknown error: %v\n", err)
		}
	}

	// client := generated.GetGatekeeperGrainClient(cst, "gatekeeper")
	// r := &generated.EnqueueRequest{
	// 	PlayerId: "aaa",
	// 	Level:    1,
	// }
	// _, err := client.Enqueue(r)
	// if err != nil {
	// 	fmt.Printf("unknown error: %v\n", err)
	// }

	_, _ = console.ReadLine()
}
