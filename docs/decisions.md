# Architecture Decision Records

## ADR-001: Use MySQL rather than SQLite

- **Date**: 2026-05-11
- **Status**: Accepted

**Context**: Initial prototype used SQLite for simplicity. Multiplayer game requires concurrent write access, connection pooling, and future horizontal scaling.

**Decision**: Migrate to MySQL 8.0 via Docker Compose. Use sqlx for type-safe queries without ORM overhead.

**Consequences**: +production-grade concurrency, +familiar tooling, -requires Docker dependency for local development.

---

## ADR-002: 54-card encoding (0-53)

- **Date**: 2026-05-11
- **Status**: Accepted

**Context**: Need a compact, unambiguous card representation for network transfer and storage.

**Decision**: Encode each card as a single integer: 0-51 for standard cards (suit × 13 + rank, rank 0=3, ..., 11=A, 12=2), 52=small joker, 53=big joker.

**Consequences**: +compact (int per card), +sortable by value, -requires conversion layer for human display.

---

## ADR-003: LLM-based AI rather than rule-based

- **Date**: 2026-05-11
- **Status**: Accepted

**Context**: AI opponents need to play Dou Di Zhu at varying difficulty levels.

**Decision**: Use LLM (OpenAI-compatible API) for AI decision-making with a system prompt describing game rules and strategy. Provider abstraction allows switching models.

**Consequences**: +human-like play, +easy difficulty tuning via prompt, -latency per move, -API cost, -requires LLM provider setup.

---

## ADR-004: Starbucks-inspired design system

- **Date**: 2026-05-12
- **Status**: Accepted

**Context**: Need a distinctive visual identity that avoids generic card-game aesthetics.

**Decision**: Adopt Starbucks design language: deep greens, warm neutrals, rounded corners, premium feel. Applied via Tailwind CSS custom theme.

**Consequences**: +distinctive brand identity, -deviates from traditional card-game UI conventions.

---

## ADR-005: Chi router over Gin

- **Date**: 2026-05-11
- **Status**: Accepted

**Context**: Need HTTP routing for REST endpoints and middleware support.

**Decision**: Use chi because it's idiomatic (net/http compatible), lightweight, and has excellent middleware composition.

**Consequences**: +standard library compatibility, +explicit middleware chain, -fewer convenience helpers than Gin.
