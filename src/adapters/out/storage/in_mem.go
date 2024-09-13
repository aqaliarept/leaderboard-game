package storage

import (
	"fmt"

	"sync"

	"github.com/Aqaliarept/leaderboard-game/application"
	"github.com/Aqaliarept/leaderboard-game/domain/competition"
	"github.com/Aqaliarept/leaderboard-game/domain/player"
)

type inMemStore struct {
	leaderboard *sync.Map
	players     *sync.Map
}

func NewTestStrore() application.LeaderBoardStorage {
	return &inMemStore{
		&sync.Map{},
		&sync.Map{}}
}

func (m *inMemStore) Start() {
	go func() {
	}()
}
func (m *inMemStore) Save(c *competition.CompetitionInfo) error {
	m.leaderboard.Store(c.Id, c)
	fmt.Printf("updated competition [%s]\n", c.Id)
	if c.IsCompleted {
		for _, p := range c.Players {
			m.players.Delete(p.Id)
			fmt.Printf("player removed [%s]\n", p.Id)
		}
	} else {
		for _, p := range c.Players {
			m.players.Store(p.Id, c)
			fmt.Printf("player added [%s]\n", p.Id)
		}
	}
	return nil
}

func (m *inMemStore) Get(id player.CompetitionId) (*competition.CompetitionInfo, error) {
	info, ok := m.leaderboard.Load(id)
	if !ok {
		return nil, fmt.Errorf("%w leaderboard id: [%s]", application.ErrNotFound, id)
	}
	return info.(*competition.CompetitionInfo), nil
}

func (m *inMemStore) GetPlayer(id player.PlayerId) (*competition.CompetitionInfo, error) {
	info, ok := m.players.Load(id)
	if !ok {
		return nil, fmt.Errorf("%w player id: [%s]", application.ErrNotFound, id)
	}
	return info.(*competition.CompetitionInfo), nil
}
