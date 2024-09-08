package domain

// import (
// 	"errors"
// 	"fmt"

// 	"github.com/gr1nd3rz/go-fast-ddd/core"
// )

// var _ core.AggregateState = CompetitionState{}

// type CompetitionState struct {
// }

// // Apply implements core.AggregateState.
// func (c CompetitionState) Apply(event core.Event) core.AggregateState {
// 	switch e := event.(type) {
// 	case PlayerCreated:
// 		return p
// 	default:
// 		panic(fmt.Errorf("%w event [%T]", errors.ErrUnsupported, event))
// 	}
// }

// type CompetitionCreated struct {
// }
