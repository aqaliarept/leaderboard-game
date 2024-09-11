package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/gr1nd3rz/go-fast-ddd/core"
)

type PlayerId core.AggregateId
type Scores int
type PlayerStatus string

const PlayerStatusIdle = PlayerStatus("")
const PlayerStatusWaiting = PlayerStatus("waiting")
const PlayerStatusPlaying = PlayerStatus("playing")

type Level int
type Country string
type CompetitionId string

const CompetitionIdUnset = CompetitionId("")

// errors
var ErrNotWaiting = errors.New("not waiting")
var ErrNotPlaying = errors.New("not playing")
var ErrAlreadyPlaying = errors.New("already playing")

// events
type PlayerCreated struct {
	Level   Level
	Country Country
}

type WaitingStarted struct {
	StartedAt time.Time
}

type WaitingExpired struct {
	StartedAt time.Time
}

type CompetitionJoined struct {
	CompetitionId CompetitionId
}

type CompetitionCompleted struct {
}

type ScoresAdded struct {
	Competition CompetitionId
	Scores      Scores
}

// state
type PlayerState struct {
	Country          Country
	Level            Level
	Status           PlayerStatus
	WaitingStartedAt time.Time
	CompetitionId    CompetitionId
}

// Apply implements core.AggregateState.
func (p PlayerState) Apply(event core.Event) core.AggregateState {
	switch e := event.(type) {
	case PlayerCreated:
		p.Country = e.Country
		p.Level = e.Level
		p.Status = PlayerStatusIdle
		return p
	case WaitingStarted:
		p.Status = PlayerStatusWaiting
		p.WaitingStartedAt = e.StartedAt
		return p
	case WaitingExpired:
		p.Status = PlayerStatusIdle
		return p
	case CompetitionJoined:
		p.Status = PlayerStatusPlaying
		p.CompetitionId = e.CompetitionId
		return p
	case ScoresAdded:
		return p
	case CompetitionCompleted:
		p.Status = PlayerStatusIdle
		p.CompetitionId = CompetitionIdUnset
		return p
	default:
		panic(fmt.Errorf("%w event [%T]", errors.ErrUnsupported, event))
	}
}

type Player struct {
	core.Aggregate[PlayerState]
}

func NewPlayer(id core.AggregateId, level Level, country Country) *Player {
	player := Player{}
	player.Initialize(id, PlayerCreated{level, country})
	return &player
}

func (p *Player) Join(now time.Time) (core.EventPack, error) {
	if p.Aggregate.State().Status == PlayerStatusPlaying {
		return nil, ErrAlreadyPlaying
	}
	return p.ProcessCommand(func(ps *PlayerState, er core.EventRaiser) error {
		er.RaiseNotEqual(PlayerStatusWaiting, ps.Status, WaitingStarted{now})
		return nil
	})
}

func (p *Player) StartCompetition(competitionId CompetitionId) (core.EventPack, error) {
	if p.Aggregate.State().Status != PlayerStatusWaiting {
		return nil, ErrNotWaiting
	}
	return p.ProcessCommand(func(ps *PlayerState, er core.EventRaiser) error {
		er.Raise(CompetitionJoined{competitionId})
		return nil
	})
}

func (p *Player) CompleteCompetition() (core.EventPack, error) {
	if p.Aggregate.State().Status != PlayerStatusPlaying {
		return nil, ErrNotPlaying
	}
	return p.ProcessCommand(func(ps *PlayerState, er core.EventRaiser) error {
		er.Raise(CompetitionCompleted{})
		return nil
	})
}

func (p *Player) WaitingExpired() (core.EventPack, error) {
	return p.ProcessCommand(func(ps *PlayerState, er core.EventRaiser) error {
		er.RaiseTrue(ps.Status == PlayerStatusWaiting, WaitingExpired{})
		return nil
	})
}

func (p *Player) AddScores(scores Scores) (core.EventPack, error) {
	if p.Aggregate.State().Status != PlayerStatusPlaying {
		return nil, ErrNotPlaying
	}
	return p.ProcessCommand(func(ps *PlayerState, er core.EventRaiser) error {
		er.Raise(ScoresAdded{ps.CompetitionId, scores})
		return nil
	})
}
