package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Aqaliarept/leaderboard-game/application"
	"github.com/Aqaliarept/leaderboard-game/domain/competition"
	"github.com/Aqaliarept/leaderboard-game/domain/player"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
)

type redisStorage struct {
	redis *redis.Client
}

func NewRedisStorage(config *application.Config) application.LeaderBoardStorage {
	opts, err := redis.ParseURL(config.RedisConnection)
	if err != nil {
		panic(err)
	}
	return &redisStorage{redis.NewClient(opts)}
}

func (m *redisStorage) Start() {
}

func competitionKey(id player.CompetitionId) string {
	return fmt.Sprintf("leaderboard:%s", id)
}
func playerKey(id player.PlayerId) string {
	return fmt.Sprintf("player:%s", id)
}

func (m *redisStorage) Save(info *competition.CompetitionInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	err = m.redis.Set(context.Background(), competitionKey(info.Id), data, 0).Err()
	if err != nil {
		return err
	}
	fmt.Printf("updated competition [%s]\n", info.Id)
	if info.IsCompleted {
		delPlayers := lo.Map(info.Players, func(pi competition.PlayerInfo, _ int) string {
			return playerKey(pi.Id)
		})
		m.redis.Del(context.Background(), delPlayers...)
		fmt.Printf("players deleted [%#v]\n", delPlayers)
	} else {
		for _, p := range info.Players {
			err = m.redis.Set(context.Background(), playerKey(p.Id), string(info.Id), 0).Err()
			if err != nil {
				return err
			}
			fmt.Printf("player added [%s]\n", p.Id)
		}
	}
	return nil
}

func (m *redisStorage) Get(id player.CompetitionId) (*competition.CompetitionInfo, error) {
	data, err := m.redis.Get(context.Background(), competitionKey(id)).Result()
	if err != nil {
		return nil, err
	}
	if err == redis.Nil {
		return nil, fmt.Errorf("%w competition id: [%s]", application.ErrNotFound, id)
	}
	res := competition.CompetitionInfo{}
	err = json.Unmarshal([]byte(data), &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (m *redisStorage) GetPlayer(id player.PlayerId) (*competition.CompetitionInfo, error) {
	data, err := m.redis.Get(context.Background(), playerKey(id)).Result()
	if err != nil {
		return nil, err
	}
	if err == redis.Nil {
		return nil, fmt.Errorf("%w player id: [%s]", application.ErrNotFound, id)
	}
	return m.Get(player.CompetitionId(data))
}
