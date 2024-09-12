package services

import (
	generated "github.com/Aqaliarept/leaderboard-game/generated/cluster"
	"github.com/asynkron/protoactor-go/cluster"
)

type LeaderboardService struct {
	cluster *cluster.Cluster
}

func NewLeaderboardService(cluster *cluster.Cluster) *LeaderboardService {
	return &LeaderboardService{cluster}
}

func (l *LeaderboardService) Join(playerId string) error {
	client := generated.GetPlayerGrainClient(l.cluster, playerId)
	_, err := client.Join(&generated.JoinRequest{Name: playerId})
	return err
}
