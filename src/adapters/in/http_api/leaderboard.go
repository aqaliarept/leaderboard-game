package httpapi

import (
	"errors"
	"net/http"

	"github.com/Aqaliarept/leaderboard-game/application"
	"github.com/Aqaliarept/leaderboard-game/application/services"
	"github.com/Aqaliarept/leaderboard-game/domain/competition"
	"github.com/Aqaliarept/leaderboard-game/domain/player"
	generated "github.com/Aqaliarept/leaderboard-game/generated/cluster"
	"github.com/Aqaliarept/leaderboard-game/generated/server/models"
	"github.com/Aqaliarept/leaderboard-game/generated/server/restapi/operations"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/samber/lo"
)

type LeaderboardApiImpl struct {
	service *services.LeaderboardService
}

func NewLeaderboardApiImpl(service *services.LeaderboardService) *LeaderboardApiImpl {
	return &LeaderboardApiImpl{service}
}

func (a *LeaderboardApiImpl) SetupHandlers(api *operations.LeaderboardAPIAPI) {
	api.AddScoresHandler = operations.AddScoresHandlerFunc(a.addScores)
	api.GetLeaderboardHandler = operations.GetLeaderboardHandlerFunc(a.getLeaderboard)
	api.GetPlayerLeaderboardHandler = operations.GetPlayerLeaderboardHandlerFunc(a.getPlayerLeaderboard)
	api.JoinHandler = operations.JoinHandlerFunc(a.join)
}

func (s *LeaderboardApiImpl) addScores(params operations.AddScoresParams) middleware.Responder {
	err := s.service.AddSrores(player.PlayerId(*params.Score.PlayerID), player.Scores(*params.Score.Score))
	if generated.IsPlayerNotPlaying(err) {
		return &operations.AddScoresConflict{}
	}
	if err != nil {
		return middleware.Error(http.StatusInternalServerError, nil)
	}
	return operations.NewAddScoresOK()
}

func (s *LeaderboardApiImpl) getLeaderboard(params operations.GetLeaderboardParams) middleware.Responder {
	info, err := s.service.GetLeaderboard(player.CompetitionId(params.LeaderboardID))
	if errors.Is(err, application.ErrNotFound) {
		return operations.NewGetLeaderboardNotFound()
	}
	if err != nil {
		return middleware.Error(http.StatusInternalServerError, nil)
	}
	return &operations.GetLeaderboardOK{mapLeaderboard(info)}
}

type EmptyJson struct {
}

// WriteResponse implements middleware.Responder.
func (e *EmptyJson) WriteResponse(w http.ResponseWriter, p runtime.Producer) {
	w.WriteHeader(200)
	w.Write([]byte("{}"))
}

func (s *LeaderboardApiImpl) getPlayerLeaderboard(params operations.GetPlayerLeaderboardParams) middleware.Responder {
	info, err := s.service.GetPlayer(player.PlayerId(params.PlayerID))
	if errors.Is(err, application.ErrNotFound) {
		return &EmptyJson{}
	}
	if err != nil {
		return middleware.Error(http.StatusInternalServerError, nil)
	}
	return &operations.GetPlayerLeaderboardOK{mapLeaderboard(info)}
}

func (s *LeaderboardApiImpl) join(params operations.JoinParams) middleware.Responder {
	err := s.service.Join(player.PlayerId(params.PlayerID))
	if generated.IsPlayerAlreadyPlaying(err) {
		return &operations.JoinConflict{}
	}
	if err != nil {
		return middleware.Error(http.StatusInternalServerError, nil)
	}
	return &operations.JoinAccepted{}
}

func mapLeaderboard(info *competition.CompetitionInfo) *models.LeaderboardResponse {
	return &models.LeaderboardResponse{
		strfmt.DateTime(info.EndsAt),
		lo.Map(info.Players, func(c competition.PlayerInfo, _ int) *models.PlayerScore {
			id := string(c.Id)
			scores := int64(c.Scores)
			return &models.PlayerScore{
				&id,
				&scores,
			}
		}),
		string(info.Id),
	}
}
