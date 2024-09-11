package actors

import (
	"fmt"
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

func Request(context actor.Context, target *actor.PID, message any) {
	req, err := context.RequestFuture(target, message, 1*time.Second).Result()
	msg := fmt.Sprintf("sending %T", message)
	if err != nil {
		context.Logger().Error(msg, "msg", err.Error())
	}
	e, ok := req.(Error)
	if ok {
		context.Logger().Error(msg, "msg", e.message)
	}
}
