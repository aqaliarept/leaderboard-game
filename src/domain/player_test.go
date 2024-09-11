package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPlayer(t *testing.T) {
	player := NewPlayer("id", 1, "-")
	player.Join(time.Now())
	require.Equal(t, PlayerStatusWaiting, player.State().Status)
}
