package main

// import (
// 	"time"

// 	"github.com/Aqaliarept/leaderboard-game/actors"
// 	console "github.com/asynkron/goconsole"
// 	"github.com/asynkron/protoactor-go/actor"
// )

// type clock struct {
// }

// func (clock) Now() time.Time {
// 	return time.Now()
// }

// func main() {
// 	system := actor.NewActorSystem()

// 	competitionCoordinatorProps := actor.PropsFromProducer(func() actor.Actor { return actors.NewCompetitionCoordinatorActor() })
// 	ccPid, err := system.Root.SpawnNamed(competitionCoordinatorProps, "competition-coordinator")
// 	if err != nil {
// 		panic(err)
// 	}
// 	gatekeeperProps := actor.PropsFromProducer(func() actor.Actor { return actors.NewGatekeeperActor(ccPid, &clock{}) })
// 	gkPid, err := system.Root.SpawnNamed(gatekeeperProps, "gatekeeper")
// 	if err != nil {
// 		panic(err)
// 	}
// 	props := actor.PropsFromProducer(func() actor.Actor { return actors.NewLeaderboard(gkPid) })
// 	leader, err := system.Root.SpawnNamed(props, "leaderboard")
// 	if err != nil {
// 		panic(err)
// 	}

// 	system.Root.RequestFuture(leader, &actors.Join{"one"}, 1*time.Second).Result()

// 	// actors.Request(system.Root, leader, actors.Join{"one"})

// 	// for i := range lo.Range(9) {
// 	// system.Root.RequestFuture(gkPid, &actors.JoinWaiting{nil, actors.PlayerId(fmt.Sprintf("%d", i)), 10, "country"}, 1*time.Second).Result()
// 	// 	// fmt.Printf("RES: %#v", res)
// 	// 	// <-time.After(100 * time.Millisecond)
// 	// }

// 	// go func() {
// 	// 	counter := 1
// 	// 	for {
// 	// 		res, _ := system.Root.RequestFuture(gkPid, &actors.JoinWaiting{nil, actors.PlayerId(fmt.Sprintf("%d", counter)), 10, "country"}, 1*time.Second).Result()
// 	// 		fmt.Printf("RES: %#v", res)
// 	// 		counter++
// 	// 		<-time.After(100 * time.Millisecond)
// 	// 	}
// 	// }()
// 	// fmt.Printf("%#v", res)
// 	_, _ = console.ReadLine()
// }
