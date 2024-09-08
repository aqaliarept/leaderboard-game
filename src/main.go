package main

import (
	"time"

	"github.com/Aqaliarept/leaderboard-game/actors"
	console "github.com/asynkron/goconsole"
	"github.com/asynkron/protoactor-go/actor"
)

type clock struct {
}

func (clock) Now() time.Time {
	return time.Now()
}

func main() {
	system := actor.NewActorSystem()

	competitionCoordinatorProps := actor.PropsFromProducer(func() actor.Actor { return actors.NewCompetitionCoordinatorActor() })
	ccPid, err := system.Root.SpawnNamed(competitionCoordinatorProps, "competition-coordinator")
	if err != nil {
		panic(err)
	}
	gatekeeperProps := actor.PropsFromProducer(func() actor.Actor { return actors.NewGatekeeperActor(ccPid, &clock{}) })
	gkPid, err := system.Root.SpawnNamed(gatekeeperProps, "gatekeeper")
	if err != nil {
		panic(err)
	}
	props := actor.PropsFromProducer(func() actor.Actor { return actors.NewLeaderboard(gkPid) })
	_, err = system.Root.SpawnNamed(props, "leaderboard")
	if err != nil {
		panic(err)
	}
	// fmt.Printf("%#v", res)
	_, _ = console.ReadLine()
}
