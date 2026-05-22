# Coding Rules

## TypeScript (Frontend)

- **Strict mode**: `strict: true` in tsconfig — no `any`, no implicit returns
- **Component files**: One component per file. Max 200 lines per component — extract sub-components or hooks when exceeded.
- **Naming**: PascalCase for components, camelCase for functions/variables, kebab-case for file names
- **Props**: Always use explicit interface types (never inline `{ a: string }`)
- **Hooks**: Extract reusable logic into custom hooks in `hooks/`. Hooks must start with `use`.
- **State**: Server state via WebSocket events, not polling. Local UI state via `useState`/`useReducer`.
- **Imports**: Library imports first, then `@/` aliases, then relative. Group with blank lines between.

## Go (Server)

- **Error handling**: Never ignore errors. Always wrap with context: `fmt.Errorf("doing X: %w", err)`
- **Naming**: Exported identifiers in PascalCase, unexported in camelCase. No underscores in names.
- **Package naming**: Single word, lowercase, no underscores. Package name should match directory name.
- **Interfaces**: Define interfaces where they're consumed, not where they're implemented.
- **Structs**: Use field tags (`json:"" db:""`) consistently. Constructor functions `NewX()` preferred over direct struct literals.
- **Concurrency**: All goroutines must have a clear lifecycle (start + stop). Use context for cancellation.
- **HTTP handlers**: Handler → validate input → call service/store → return response. No business logic in handlers.

## Both

- **Tests**: Write tests for new logic. Game logic must have tests covering edge cases.
- **Comments**: Explain WHY, not WHAT. No docstring noise. Remove dead/commented-out code.
- **No magic numbers**: Define constants for all numeric values with semantic meaning.
- **Dependencies**: No new dependencies without explicit approval.
