package storage

import (
	"fmt"

	"github.com/Aqaliarept/leaderboard-game/application"
	"github.com/Aqaliarept/leaderboard-game/domain/competition"
	"github.com/Aqaliarept/leaderboard-game/domain/player"
)

type InMemStore struct {
	leaderboard map[player.CompetitionId]*competition.CompetitionInfo
	players     map[player.PlayerId]*competition.CompetitionInfo
	input       chan *competition.CompetitionInfo
}

func NewMemStrore() application.LeaderBoardStorage {
	return &InMemStore{
		make(map[player.CompetitionId]*competition.CompetitionInfo),
		make(map[player.PlayerId]*competition.CompetitionInfo),
		make(chan *competition.CompetitionInfo, 100)}
}

func (m *InMemStore) Start() {
	go func() {
		for v := range m.input {
			m.leaderboard[v.Id] = v
			fmt.Printf("updated competition [%s]\n", v.Id)
			if v.IsCompleted {
				for _, p := range v.Players {
					delete(m.players, p.Id)
					fmt.Printf("player removed [%s]\n", p.Id)
				}
			} else {
				for _, p := range v.Players {
					m.players[p.Id] = v
					fmt.Printf("player added [%s]\n", p.Id)
				}
			}
		}
	}()
}
func (m *InMemStore) Save(competition *competition.CompetitionInfo) {
	m.input <- competition
}

func (m *InMemStore) Get(id player.CompetitionId) (*competition.CompetitionInfo, error) {
	info, ok := m.leaderboard[id]
	if !ok {
		return nil, fmt.Errorf("%w leaderboard id: [%s]", application.ErrNotFound, id)
	}
	return info, nil
}

func (m *InMemStore) GetPlayer(id player.PlayerId) (*competition.CompetitionInfo, error) {
	info, ok := m.players[id]
	if !ok {
		return nil, fmt.Errorf("%w player id: [%s]", application.ErrNotFound, id)
	}
	return info, nil
}
