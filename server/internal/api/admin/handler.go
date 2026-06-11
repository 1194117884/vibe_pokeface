package admin

import (
	"github.com/yongkl/vibe-pokeface/internal/api/ws"
	"github.com/yongkl/vibe-pokeface/internal/model"
)

type Handler struct {
	Dashboard         *DashboardHandler
	Users             *AdminUserHandler
	Rooms             *AdminRoomHandler
	AIChars           *AICharacterHandler
	LLMConfig         *LLMConfigHandler
	AIRuns            *AIRunHandler
	AITemplates       *AITemplateHandler
	Scores            *ScoreHandler
	RegistrationCodes *RegistrationCodesHandler
}

func NewHandler(userStore *model.UserDB, gameStore *model.GameStore, aiStore *model.AIStore, hub *ws.Hub, regCodeStore RegistrationCodeAdminStore) *Handler {
	return &Handler{
		Dashboard:         NewDashboardHandler(userStore, gameStore, hub),
		Users:             NewAdminUserHandler(userStore),
		Rooms:             NewAdminRoomHandler(gameStore),
		AIChars:           NewAICharacterHandler(aiStore),
		LLMConfig:         NewLLMConfigHandler(aiStore, aiStore),
		AIRuns:            NewAIRunHandler(aiStore),
		AITemplates:       NewAITemplateHandler(aiStore),
		Scores:            NewScoreHandler(gameStore),
		RegistrationCodes: NewRegistrationCodesHandler(regCodeStore),
	}
}
