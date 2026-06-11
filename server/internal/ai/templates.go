package ai

import (
	"context"
	"encoding/json"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/yongkl/vibe-pokeface/internal/model"
)

var templateVarPattern = regexp.MustCompile(`\$\{([A-Za-z0-9_]+)\}`)

func RenderTemplate(content string, vars map[string]string) string {
	return templateVarPattern.ReplaceAllStringFunc(content, func(match string) string {
		key := strings.TrimSuffix(strings.TrimPrefix(match, "${"), "}")
		if value, ok := vars[key]; ok {
			return value
		}
		return match
	})
}

func TemplateVariables(content string) []string {
	matches := templateVarPattern.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool, len(matches))
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 || seen[match[1]] {
			continue
		}
		seen[match[1]] = true
		result = append(result, match[1])
	}
	return result
}

func (a *AIAgent) renderPublishedPrompt(phase string, templateType string, defaultContent string) string {
	if a.aiStore == nil {
		return defaultContent
	}
	tmpl, err := a.aiStore.GetPublishedPromptTemplate(context.Background(), a.gameTypeSnapshot(), phase, templateType)
	if err != nil || tmpl == nil {
		return defaultContent
	}
	return RenderTemplate(tmpl.Content, a.templateVars(phase, defaultContent))
}

func (a *AIAgent) templateVars(phase string, defaultContent string) map[string]string {
	name := "AI玩家"
	personality := "冷静分析，稳健出牌"
	playStyle := "balanced"
	if a.Character != nil {
		if a.Character.Name != "" {
			name = a.Character.Name
		}
		if a.Character.Personality != nil && *a.Character.Personality != "" {
			personality = *a.Character.Personality
		}
		if a.Character.PlayStyle != "" {
			playStyle = a.Character.PlayStyle
		}
	}
	return map[string]string{
		"game_type":       a.gameTypeSnapshot(),
		"phase":           phase,
		"seat":            strconv.Itoa(a.Seat),
		"user_id":         a.UserID,
		"room_id":         a.roomID,
		"character_name":  name,
		"personality":     personality,
		"play_style":      playStyle,
		"default_content": defaultContent,
	}
}

func (a *AIAgent) toolSchemasForPhase(phase string) []ToolSchema {
	defaults := GetToolSchemasForGame(a.gameTypeSnapshot(), phase)
	if a.aiStore == nil {
		return defaults
	}
	templates, err := a.aiStore.ListPublishedToolTemplates(context.Background(), a.gameTypeSnapshot(), phase)
	if err != nil || len(templates) == 0 {
		return defaults
	}
	byName := make(map[string]model.AIToolTemplate, len(templates))
	for _, tmpl := range templates {
		byName[tmpl.ToolName] = tmpl
	}
	result := make([]ToolSchema, 0, len(defaults))
	for _, schema := range defaults {
		tmpl, ok := byName[schema.Function.Name]
		if !ok {
			result = append(result, schema)
			continue
		}
		if !tmpl.Enabled {
			continue
		}
		if tmpl.Description != nil {
			schema.Function.Description = *tmpl.Description
		}
		if tmpl.ParametersJSON != nil && *tmpl.ParametersJSON != "" {
			var params ParamSchema
			if err := json.Unmarshal([]byte(*tmpl.ParametersJSON), &params); err == nil && params.Type != "" {
				schema.Function.Parameters = params
			} else if err != nil {
				log.Printf("[AI:%s] invalid tool template parameters for %s: %v", a.UserID, tmpl.ToolName, err)
			}
		}
		result = append(result, schema)
	}
	return result
}
