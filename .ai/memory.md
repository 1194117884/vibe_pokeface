# AI Session Memory

Cross-session persistent knowledge about the project and its development.

## Key Decisions

- MySQL chosen over SQLite for concurrency (see `docs/decisions/README.md` ADR-001)
- 54-card encoding: 0-51 standard, 52=small joker, 53=big joker (ADR-002)
- LLM-based AI opponents, not rule-based (ADR-003)
- Starbucks-inspired design system (ADR-004)
- Chi router over Gin for net/http compatibility (ADR-005)

## Behavior Preferences

- User wants plans confirmed before code execution
- Reports required after each logical step (not just at the end)
- One task at a time — no multitasking
- Conventional commits: `type(scope): description`

## Important Conventions

- Game logic changes must be tested thoroughly (core correctness)
- No new dependencies without explicit approval
- Max 3 files per change step
- Do not modify config files, docker-compose, CI/CD, or migration files without instruction

## Recurring Patterns

<!-- Updated as patterns emerge across sessions -->

## Known Pitfalls

<!-- Updated when bugs or issues recur -->
