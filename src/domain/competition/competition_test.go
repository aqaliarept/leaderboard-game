package competition

import (
	"testing"
	"time"

	"github.com/Aqaliarept/leaderboard-game/domain/player"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/zavitax/sortedset-go"
)

func mapPlayers(set *sortedset.SortedSet[player.PlayerId, player.Scores, player.PlayerId]) []player.PlayerId {
	return lo.Map(set.GetRangeByRank(1, -1, false), func(n *sortedset.SortedSetNode[player.PlayerId, player.Scores, player.PlayerId], _ int) player.PlayerId {
		return n.Key()
	})
}

type MyObject struct {
	ID   string
	Data string
}

func TestCompetition(t *testing.T) {
	t.Run(`Given comptition with 3 players
		When reporting the scores
		Then players should be ranged by their scores
	`, func(t *testing.T) {
		start := time.Now()
		c := New("test", []player.PlayerId{"1", "2", "3"}, start, 1*time.Hour)
		require.Equal(t, []player.PlayerId{"1", "2", "3"}, mapPlayers(c.State().leaderboard))
		c.ReportScores("1", 10)
		c.ReportScores("2", 20)
		c.ReportScores("3", 30)
		require.Equal(t, []player.PlayerId{"3", "2", "1"}, mapPlayers(c.State().leaderboard))
		c.ReportScores("2", 20)
		require.Equal(t, []player.PlayerId{"2", "3", "1"}, mapPlayers(c.State().leaderboard))
	})

	t.Run(`Players with equal scores are sorted aphpabetically by ID
`, func(t *testing.T) {
		start := time.Now()
		c := New("test", []player.PlayerId{"2", "1", "3"}, start, 1*time.Hour)
		require.Equal(t, []player.PlayerId{"1", "2", "3"}, mapPlayers(c.State().leaderboard))

		c.ReportScores("3", 30)
		c.ReportScores("2", 30)
		require.Equal(t, []player.PlayerId{"2", "3", "1"}, mapPlayers(c.State().leaderboard))

	})
}
