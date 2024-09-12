package httpapi

import (
	"github.com/Aqaliarept/leaderboard-game/application/services"
	"github.com/Aqaliarept/leaderboard-game/generated/server/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
)

type LeaderboardApiImpl struct {
	service *services.LeaderboardService
}

func NewApiImpl(service *services.LeaderboardService) *LeaderboardApiImpl {
	return &LeaderboardApiImpl{service}
}

func (a *LeaderboardApiImpl) SetupHandlers(api *operations.LeaderboardAPIAPI) {
	api.AddScoresHandler = operations.AddScoresHandlerFunc(a.addScores)
	api.GetLeaderboardHandler = operations.GetLeaderboardHandlerFunc(a.getLeaderboard)
	api.GetPlayerLeaderboardHandler = operations.GetPlayerLeaderboardHandlerFunc(a.getPlayerLeaderboard)
	api.JoinHandler = operations.JoinHandlerFunc(a.join)
}

func (a *LeaderboardApiImpl) addScores(params operations.AddScoresParams) middleware.Responder {
	return middleware.NotImplemented("operation operations.AddScores has not yet been implemented")
}

func (s *LeaderboardApiImpl) getLeaderboard(params operations.GetLeaderboardParams) middleware.Responder {
	return middleware.NotImplemented("operation operations.AddScores has not yet been implemented")
}

func (s *LeaderboardApiImpl) getPlayerLeaderboard(params operations.GetPlayerLeaderboardParams) middleware.Responder {
	return middleware.NotImplemented("operation operations.AddScores has not yet been implemented")
}

func (s *LeaderboardApiImpl) join(params operations.JoinParams) middleware.Responder {
	err := s.service.Join(params.PlayerID)
	if err != nil {
		return &operations.JoinConflict{}
	}
	return &operations.JoinAccepted{}
}
