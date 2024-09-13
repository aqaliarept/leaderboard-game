package services

import (
	"github.com/Aqaliarept/leaderboard-game/application"
	"github.com/Aqaliarept/leaderboard-game/domain/competition"
	"github.com/Aqaliarept/leaderboard-game/domain/player"
	generated "github.com/Aqaliarept/leaderboard-game/generated/cluster"
	"github.com/asynkron/protoactor-go/cluster"
)

type LeaderboardService struct {
	cluster *cluster.Cluster
	storage application.LeaderBoardStorage
}

func NewLeaderboardService(cluster *cluster.Cluster, storage application.LeaderBoardStorage) *LeaderboardService {
	return &LeaderboardService{cluster, storage}
}

func (l *LeaderboardService) Join(id player.PlayerId) error {
	client := generated.GetPlayerGrainClient(l.cluster, string(id))
	_, err := client.Join(&generated.JoinRequest{})
	return err
}

func (l *LeaderboardService) AddSrores(id player.PlayerId, scores player.Scores) error {
	client := generated.GetPlayerGrainClient(l.cluster, string(id))
	_, err := client.AddScores(&generated.AddScoresRequest{Scores: int32(scores)})
	return err
}

func (l *LeaderboardService) GetLeaderboard(id player.CompetitionId) (*competition.CompetitionInfo, error) {
	return l.storage.Get(player.CompetitionId(id))
}

func (l *LeaderboardService) GetPlayer(id player.PlayerId) (*competition.CompetitionInfo, error) {
	return l.storage.GetPlayer(player.PlayerId(id))
}
