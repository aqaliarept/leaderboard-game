package storage

import (
	"strconv"

	"github.com/Aqaliarept/leaderboard-game/application"
	"github.com/Aqaliarept/leaderboard-game/domain/player"
	"github.com/samber/lo"
)

type inMemPlayersRepo struct {
	players map[player.PlayerId]*application.PlayerInfo
}

// Get implements application.PlayersRepo.
func (i *inMemPlayersRepo) Get(id player.PlayerId) *application.PlayerInfo {
	p, ok := i.players[id]
	if ok {
		return p
	}
	return &application.PlayerInfo{id, 1}
}

func NewTestPlayerRepo() application.PlayersRepo {
	repo := inMemPlayersRepo{}
	repo.players = make(map[player.PlayerId]*application.PlayerInfo)
	for i := range lo.Range(30) {
		id := player.PlayerId(strconv.Itoa(i + 1))
		repo.players[id] = &application.PlayerInfo{id, player.Level(i + 1)}
	}
	return &repo
}
