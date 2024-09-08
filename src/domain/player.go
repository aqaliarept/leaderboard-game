package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/gr1nd3rz/go-fast-ddd/core"
)

type PlayerStatus string

const PlayerStatusIdle = PlayerStatus("")
const PlayerStatusWaiting = PlayerStatus("waiting")
const PlayerStatusPlaying = PlayerStatus("playing")

type Level int
type Country string
type BucketId int
type CompetitionId string

const CompetitionIdUnset = CompetitionId("")

// errors
var ErrAlreadyPlaying = errors.New("already playing")

// events
type PlayerCreated struct {
	Level   Level
	Country Country
}

type WaitingStarted struct {
	StartedAt time.Time
	BucketId  BucketId
}

type CompetitionJoined struct {
	CompetitionId CompetitionId
}

type CompetitionCompleted struct {
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
	case CompetitionJoined:
		p.Status = PlayerStatusPlaying
		p.CompetitionId = e.CompetitionId
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

type BucketSelector struct {
}

func (BucketSelector) GetBucket(level Level, coutry Country) (BucketId, error) {
	if level <= 10 {
		return 1, nil
	}
	if level > 10 && level <= 20 {
		return 2, nil
	}
	return 0, fmt.Errorf("%w level: [%d]", errors.ErrUnsupported, level)
}

func (p *Player) Join(now time.Time, bucketSelector BucketSelector) (core.EventPack, error) {
	return p.ProcessCommand(func(ps *PlayerState, er core.EventRaiser) error {
		switch ps.Status {
		case PlayerStatusPlaying:
			return ErrAlreadyPlaying
		case PlayerStatusIdle:
			bucket, err := bucketSelector.GetBucket(ps.Level, ps.Country)
			if err != nil {
				return err
			}
			er.Raise(WaitingStarted{now, bucket})
		case PlayerStatusWaiting:
		default:
			panic(fmt.Errorf("%w status [%s]", errors.ErrUnsupported, ps.Status))
		}
		return nil
	})
}

func (p *Player) StartCompetition(competitionId CompetitionId) (core.EventPack, error) {
	return p.ProcessCommand(func(ps *PlayerState, er core.EventRaiser) error {
		switch ps.Status {
		case PlayerStatusWaiting:
			er.Raise(CompetitionJoined{competitionId})
		case PlayerStatusIdle:
			fallthrough
		case PlayerStatusPlaying:
			return fmt.Errorf("wrong status [%s]", ps.Status)
		default:
			panic(fmt.Errorf("%w status [%s]", errors.ErrUnsupported, ps.Status))
		}
		return nil
	})
}

func (p *Player) EndCompetition(competitionId CompetitionId) (core.EventPack, error) {
	return p.ProcessCommand(func(ps *PlayerState, er core.EventRaiser) error {
		switch ps.Status {
		case PlayerStatusPlaying:
			er.Raise(CompetitionCompleted{})
		case PlayerStatusIdle:
			fallthrough
		case PlayerStatusWaiting:
			return fmt.Errorf("wrong status [%s]", ps.Status)
		default:
			panic(fmt.Errorf("%w status [%s]", errors.ErrUnsupported, ps.Status))
		}
		return nil
	})
}
