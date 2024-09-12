package player

import (
	"testing"
	"time"

	"github.com/Aqaliarept/leaderboard-game/domain"
	"github.com/stretchr/testify/require"
)

func TestPlayer(t *testing.T) {
	player := New("id", 1, "-")
	_, err := player.Join(time.Now())
	require.NoError(t, err)
	require.Equal(t, PlayerStatusWaiting, player.State().Status)

	pack, err := player.StartCompetition("comp")
	require.NoError(t, err)
	_, err = domain.EventOfType[CompetitionJoined](pack)
	require.NoError(t, err)

	_, err = player.Join(time.Now())
	require.ErrorIs(t, err, ErrAlreadyPlaying)
}
