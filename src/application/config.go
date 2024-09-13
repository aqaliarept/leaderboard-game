package application

import (
	"fmt"
	"time"

	"k8s.io/utils/env"
)

type Config struct {
	QueueWaitingTimeout time.Duration
	CompetitionDuration time.Duration
	CompetitionSize     uint
	MinCompetitionSize  uint
}

func NewConfig() *Config {
	conf := Config{}
	var err error

	sec, err := env.GetInt("COMPETITION_DURATION", 3600)
	if err != nil {
		panic(err)
	}
	conf.CompetitionDuration = time.Duration(sec) * time.Second

	sec, err = env.GetInt("QUEUE_WAITING_TIMEOUT", 30)
	if err != nil {
		panic(err)
	}
	conf.QueueWaitingTimeout = time.Duration(sec) * time.Second

	sec, err = env.GetInt("COMPETITION_SIZE", 10)
	if err != nil {
		panic(err)
	}
	conf.CompetitionSize = uint(sec)

	sec, err = env.GetInt("MIN_COMPETITION_SIZE", 2)
	if err != nil {
		panic(err)
	}
	conf.MinCompetitionSize = uint(sec)
	fmt.Printf("Config: [%#v]", conf)
	return &conf
}
