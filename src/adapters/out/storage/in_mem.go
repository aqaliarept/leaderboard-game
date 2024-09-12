package storage

import (
	"fmt"

	"github.com/Aqaliarept/leaderboard-game/application"
	"github.com/Aqaliarept/leaderboard-game/domain/competition"
	"github.com/Aqaliarept/leaderboard-game/domain/player"
	"github.com/samber/lo"
)

type InMemStore struct {
	leaderboard map[player.CompetitionId]*competition.CompetitionInfo
	players     map[player.PlayerId]*competition.CompetitionInfo
	input       chan lo.Tuple2[*competition.CompetitionInfo, bool]
}

func NewMemStrore() application.LeaderBoardStorage {
	return &InMemStore{
		make(map[player.CompetitionId]*competition.CompetitionInfo),
		make(map[player.PlayerId]*competition.CompetitionInfo),
		make(chan lo.Tuple2[*competition.CompetitionInfo, bool], 100)}
}

func (m *InMemStore) Start() {
	go func() {
		for v := range m.input {
			m.leaderboard[v.A.Id] = v.A
			if v.B {
				for _, p := range v.A.Players {
					delete(m.players, p.Id)
				}
			} else {
				for _, p := range v.A.Players {
					m.players[p.Id] = v.A
				}
			}
		}
	}()
}
func (m *InMemStore) Save(competition *competition.CompetitionInfo, isCompleted bool) {
	m.input <- lo.T2(competition, isCompleted)
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
