# Security Rules

## Authentication

- JWT tokens signed with HS256 using `JWT_SECRET` from environment — never hardcoded
- Tokens expire server-side. Client must handle 401 by clearing token and redirecting to login.
- WebSocket connections authenticated via `?token=` query parameter
- Admin endpoints require JWT + admin role check

## Secrets & Environment

- All secrets in `.env` (gitignored). `.env.example` contains placeholder values only.
- Never log tokens, passwords, or secrets. Redact sensitive fields from logs.
- Database credentials must not appear in code, config files, or commit history.

## Input Validation

- **Server validates all client input** — game actions, chat messages, room IDs. Never trust the client.
- Validate types, lengths, and ranges before processing.
- SQL injection: All queries use parameterized statements via sqlx. Never string-concatenate SQL.
- XSS: React's JSX auto-escapes by default. For chat messages, sanitize any HTML before rendering.

## WebSocket

- Validate JWT on WebSocket upgrade before establishing connection
- Rate limit connections per IP
- Validate message type and payload structure before routing
- Close connections that send malformed messages (potential attack)

## Dependencies

- No direct DOM manipulation or `dangerouslySetInnerHTML` without explicit review
- Audit new dependencies before adding (known vulnerabilities, maintenance status)
- Keep Go modules and npm packages updated for security patches

## Data Privacy

- Passwords hashed with bcrypt (not stored plaintext)
- User data accessible only by the owning user (except admin)
- Chat messages stored server-side only for current session
- Admin panel access requires authenticated admin user
