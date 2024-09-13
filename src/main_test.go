package main

import (
	"strconv"
	"testing"
	"time"

	"github.com/Aqaliarept/leaderboard-game/adapters/out/storage"
	"github.com/Aqaliarept/leaderboard-game/application"
	"github.com/Aqaliarept/leaderboard-game/application/grains"
	"github.com/Aqaliarept/leaderboard-game/application/services"
	"github.com/Aqaliarept/leaderboard-game/domain/player"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func Test_application_container_deps_configuration(t *testing.T) {
	err := fx.ValidateApp(configureContainer())
	require.NoError(t, err)
}

func Test_application_tests(t *testing.T) {
	config := application.Config{
		4 * time.Second,
		10 * time.Second,
		10,
		1,
	}
	clock := NewClock()
	store := storage.NewMemStrore()
	store.Start()
	cluster := NewCluster(
		grains.NewPlayerGrainFactory(clock),
		grains.NewCompetitionGrainFactory(&config, clock, store),
		grains.NewGatekeeperFactory(clock, &config),
	)
	cluster.StartMember()
	cluster.Shutdown(false)
	svc := services.NewLeaderboardService(cluster, store)

	players := lo.Map(lo.Range(21), func(i int, _ int) player.PlayerId {
		return player.PlayerId(strconv.Itoa(i))
	})

	// enque
	_, _, done := lo.WaitFor(func(i int) bool {
		err := svc.Join(player.PlayerId(players[0]))
		return err == nil
	}, 5*time.Second, 1*time.Second)
	require.True(t, done)

	for i := 1; i < len(players); i++ {
		id := player.PlayerId(strconv.Itoa(i))
		err := svc.Join(id)
		require.NoError(t, err)
	}
	time.Sleep(5 * time.Second)

	// first 2 groups should have 10 players
	for i := 0; i < 20; i++ {
		leaderboard, err := svc.GetPlayer(players[i])
		require.NoError(t, err)
		require.Equal(t, 10, len(leaderboard.Players))
	}

	// the last player should be in group with 1 player
	leaderboard, err := svc.GetPlayer(players[20])
	require.NoError(t, err)
	require.Equal(t, 1, len(leaderboard.Players))
}
