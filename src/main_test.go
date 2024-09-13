package main

import (
	"strconv"
	"testing"
	"time"

	"github.com/Aqaliarept/leaderboard-game/adapters/out/storage"
	"github.com/Aqaliarept/leaderboard-game/application"
	"github.com/Aqaliarept/leaderboard-game/application/grains"
	"github.com/Aqaliarept/leaderboard-game/application/services"
	"github.com/Aqaliarept/leaderboard-game/domain/competition"
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
		"",
		// "redis://localhost:6379",
		4 * time.Second,
		10 * time.Second,
		10,
		1,
	}
	clock := NewClock()
	// store := storage.NewRedisStorage(config)
	store := storage.NewTestStrore()
	playerRepo := storage.NewTestPlayerRepo()
	cluster := NewCluster(
		grains.NewPlayerGrainFactory(clock, playerRepo),
		grains.NewCompetitionGrainFactory(&config, clock, store),
		grains.NewGatekeeperFactory(clock, &config),
	)
	cluster.StartMember()
	cluster.Shutdown(false)
	svc := services.NewLeaderboardService(cluster, store)

	players := lo.Map(lo.Range(21), func(i int, _ int) player.PlayerId {
		return player.PlayerId(strconv.Itoa(i + 1))
	})

	// enque
	_, _, done := lo.WaitFor(func(i int) bool {
		err := svc.Join(player.PlayerId(players[0]))
		return err == nil
	}, 5*time.Second, 1*time.Second)
	require.True(t, done)

	for i := 1; i < len(players); i++ {
		go func() {
			err := svc.Join(players[i])
			if err != nil {
				panic(err)
			}
		}()
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

	for i := 0; i < 10; i++ {
		err := svc.AddSrores(players[i], player.Scores(i+1))
		require.NoError(t, err)
	}
	time.Sleep(2 * time.Second)
	leaderboard, err = svc.GetPlayer(players[0])
	require.NoError(t, err)
	require.NotEmpty(t, leaderboard.EndsAt)
	require.Equal(t, []competition.PlayerInfo{
		{"10", 10},
		{"9", 9},
		{"8", 8},
		{"7", 7},
		{"6", 6},
		{"5", 5},
		{"4", 4},
		{"3", 3},
		{"2", 2},
		{"1", 1},
	}, leaderboard.Players)
}
