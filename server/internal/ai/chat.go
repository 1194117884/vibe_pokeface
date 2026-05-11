package ai

import "fmt"

// AIChat generates character-based chat messages during a game.
type AIChat struct {
	CharacterID int
	Name        string
	Personality string
	Catchphrase string
}

// NewAIChat creates a new AIChat for the given character.
func NewAIChat(charID int, name, personality, catchphrase string) *AIChat {
	return &AIChat{
		CharacterID: charID,
		Name:        name,
		Personality: personality,
		Catchphrase: catchphrase,
	}
}

// GenerateGreeting returns the character's greeting message.
func (c *AIChat) GenerateGreeting() string {
	if c.Catchphrase != "" {
		return fmt.Sprintf("%s 来啦！%s", c.Name, c.Catchphrase)
	}
	return fmt.Sprintf("%s 来了，开始吧！", c.Name)
}

// GenerateCatchphrase returns the character's catchphrase, or a fallback.
func (c *AIChat) GenerateCatchphrase() string {
	if c.Catchphrase == "" {
		return "嘿嘿嘿"
	}
	return c.Catchphrase
}

// GeneratePlayComment returns a comment after the character plays cards or passes.
func (c *AIChat) GeneratePlayComment(action string, cards []int) string {
	if action == "pass" {
		return "过过过"
	}
	return "接招！"
}
