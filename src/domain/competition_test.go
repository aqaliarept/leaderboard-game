package domain

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/zavitax/sortedset-go"
)

func mapPlayers(set *sortedset.SortedSet[PlayerId, Scores, PlayerId]) []PlayerId {
	return lo.Map(set.GetRangeByRank(1, -1, false), func(n *sortedset.SortedSetNode[PlayerId, Scores, PlayerId], _ int) PlayerId {
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
		c := NewCompetition("test", []PlayerId{"1", "2", "3"})
		require.Equal(t, []PlayerId{"1", "2", "3"}, mapPlayers(c.State().leaderboard))
		c.ReportScores("1", 10)
		c.ReportScores("2", 20)
		c.ReportScores("3", 30)
		require.Equal(t, []PlayerId{"3", "2", "1"}, mapPlayers(c.State().leaderboard))
		c.ReportScores("2", 20)
		require.Equal(t, []PlayerId{"2", "3", "1"}, mapPlayers(c.State().leaderboard))
	})
}
