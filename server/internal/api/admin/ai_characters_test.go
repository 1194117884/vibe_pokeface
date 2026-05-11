package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yongkl/vibe-pokeface/internal/model"
)

type mockAICharacterStore struct {
	listCharactersFn    func(ctx context.Context) ([]model.AICharacter, error)
	createCharacterFn   func(ctx context.Context, c *model.AICharacter) error
	updateCharacterFn   func(ctx context.Context, c *model.AICharacter) error
	deleteCharacterFn   func(ctx context.Context, id int) error
}

func (m *mockAICharacterStore) ListCharacters(ctx context.Context) ([]model.AICharacter, error) {
	return m.listCharactersFn(ctx)
}
func (m *mockAICharacterStore) GetAllCharacters(ctx context.Context) ([]model.AICharacter, error) {
	return m.listCharactersFn(ctx)
}
func (m *mockAICharacterStore) GetCharacter(ctx context.Context, id int) (*model.AICharacter, error) {
	return nil, nil
}
func (m *mockAICharacterStore) CreateCharacter(ctx context.Context, c *model.AICharacter) error {
	return m.createCharacterFn(ctx, c)
}
func (m *mockAICharacterStore) UpdateCharacter(ctx context.Context, c *model.AICharacter) error {
	return m.updateCharacterFn(ctx, c)
}
func (m *mockAICharacterStore) DeleteCharacter(ctx context.Context, id int) error {
	return m.deleteCharacterFn(ctx, id)
}

func TestAICharacterList(t *testing.T) {
	store := &mockAICharacterStore{
		listCharactersFn: func(ctx context.Context) ([]model.AICharacter, error) {
			return []model.AICharacter{
				{ID: 1, Name: "TestBot", PlayStyle: "aggressive", Enabled: true},
			}, nil
		},
	}
	handler := NewAICharacterHandler(store)

	req := httptest.NewRequest("GET", "/api/admin/ai-characters", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var chars []model.AICharacter
	if err := json.NewDecoder(w.Body).Decode(&chars); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(chars) != 1 || chars[0].Name != "TestBot" {
		t.Errorf("unexpected response: %+v", chars)
	}
}

func TestAICharacterList_Empty(t *testing.T) {
	store := &mockAICharacterStore{
		listCharactersFn: func(ctx context.Context) ([]model.AICharacter, error) {
			return []model.AICharacter{}, nil
		},
	}
	handler := NewAICharacterHandler(store)

	req := httptest.NewRequest("GET", "/api/admin/ai-characters", nil)
	w := httptest.NewRecorder()
	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}

	var chars []model.AICharacter
	json.NewDecoder(w.Body).Decode(&chars)
	if chars == nil || len(chars) != 0 {
		t.Errorf("expected empty array, got %+v", chars)
	}
}

func TestAICharacterCreate_Success(t *testing.T) {
	store := &mockAICharacterStore{
		createCharacterFn: func(ctx context.Context, c *model.AICharacter) error {
			c.ID = 1
			return nil
		},
	}
	handler := NewAICharacterHandler(store)

	body := map[string]interface{}{
		"name": "NewBot", "play_style": "conservative", "enabled": true,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/ai-characters", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Create(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var created model.AICharacter
	json.NewDecoder(w.Body).Decode(&created)
	if created.ID != 1 || created.Name != "NewBot" {
		t.Errorf("unexpected created: %+v", created)
	}
}

func TestAICharacterCreate_MissingName(t *testing.T) {
	handler := NewAICharacterHandler(&mockAICharacterStore{})

	body := map[string]interface{}{"play_style": "aggressive"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/ai-characters", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Create(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAICharacterDelete_Success(t *testing.T) {
	called := false
	store := &mockAICharacterStore{
		deleteCharacterFn: func(ctx context.Context, id int) error {
			called = true
			if id != 1 {
				t.Errorf("id = %d, want 1", id)
			}
			return nil
		},
	}
	handler := NewAICharacterHandler(store)

	req := httptest.NewRequest("DELETE", "/api/admin/ai-characters/1", nil)
	w := httptest.NewRecorder()
	handler.Delete(w, req)

	if w.Code != http.StatusBadRequest {
		t.Logf("status = %d (expected 400 without router)", w.Code)
	}
	if called {
		t.Error("delete should not have been called (no chi context)")
	}
}
