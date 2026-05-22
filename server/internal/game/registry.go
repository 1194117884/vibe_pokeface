package game

import "fmt"

var engineFactories = map[string]func() GameEngine{}

// RegisterEngine registers a GameEngine factory for a given game type.
// Called from init() in each engine subpackage.
func RegisterEngine(gameType string, factory func() GameEngine) {
	engineFactories[gameType] = factory
}

// NewEngine creates a GameEngine for the given game type.
func NewEngine(gameType string) (GameEngine, error) {
	factory, ok := engineFactories[gameType]
	if !ok {
		return nil, fmt.Errorf("unknown game type: %s", gameType)
	}
	return factory(), nil
}

// ValidGameTypes returns the list of supported game types.
func ValidGameTypes() []string {
	types := make([]string, 0, len(engineFactories))
	for t := range engineFactories {
		types = append(types, t)
	}
	return types
}
