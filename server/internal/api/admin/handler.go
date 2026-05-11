package admin

import (
	"github.com/yongkl/vibe-pokeface/internal/api/ws"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

type Handler struct {
	Dashboard *DashboardHandler
	Users     *AdminUserHandler
	Rooms     *AdminRoomHandler
	AIChars   *AICharacterHandler
	LLMConfig *LLMConfigHandler
	Scores    *ScoreHandler
}

func NewHandler(userStore *model.UserDB, gameStore *model.GameStore, aiStore *model.AIStore, hub *ws.Hub) *Handler {
	return &Handler{
		Dashboard: NewDashboardHandler(userStore, gameStore, hub),
		Users:     NewAdminUserHandler(userStore),
		Rooms:     NewAdminRoomHandler(gameStore),
		AIChars:   NewAICharacterHandler(aiStore),
		LLMConfig: NewLLMConfigHandler(aiStore, aiStore),
		Scores:    NewScoreHandler(gameStore),
	}
}
