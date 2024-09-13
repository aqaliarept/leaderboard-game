package application

import (
	"errors"

	"github.com/Aqaliarept/leaderboard-game/domain/competition"
	"github.com/Aqaliarept/leaderboard-game/domain/player"
)

var ErrNotFound = errors.New("not found")

type LeaderBoardStorage interface {
	Start()
	Save(competition *competition.CompetitionInfo) error
	Get(id player.CompetitionId) (*competition.CompetitionInfo, error)
	GetPlayer(id player.PlayerId) (*competition.CompetitionInfo, error)
}

type PlayerInfo struct {
	Id    player.PlayerId
	Level player.Level
}

type PlayersRepo interface {
	Get(id player.PlayerId) *PlayerInfo
}
