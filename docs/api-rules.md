# API Rules & Conventions

## REST Endpoints

### URL Pattern
```
/api/{resource}
/api/{resource}/{id}
/api/admin/{resource}
```

### Standard Response Format

**Success (single):**
```json
{ "id": 1, "nickname": "player1", ... }
```

**Success (list with pagination):**
```json
{
  "data": [...],
  "total": 42,
  "page": 1,
  "pageSize": 20
}
```

**Error:**
```json
{ "error": "description of what went wrong" }
```

### HTTP Status Codes
- `200` — OK (GET, PUT, PATCH)
- `201` — Created (POST)
- `204` — No Content (DELETE)
- `400` — Bad Request (invalid input)
- `401` — Unauthorized (missing/invalid token)
- `403` — Forbidden (valid token but insufficient permissions)
- `404` — Not Found
- `409` — Conflict (duplicate, state conflict)
- `429` — Too Many Requests (rate limited)
- `500` — Internal Server Error

## WebSocket Messages

### Message Format
```json
{
  "type": "message_type",
  "payload": { ... }
}
```

### Message Types (Client → Server)
| type | payload | description |
|---|---|---|
| `join_room` | `{ room_id }` | Join an existing room |
| `create_room` | `{ }` | Create a new room |
| `leave_room` | `{ }` | Leave current room |
| `room_action` | `{ action, data }` | Game action (play cards, bid, pass) |
| `chat_message` | `{ content }` | Send chat message |
| `ready` | `{ }` | Signal ready to start |

### Message Types (Server → Client)
| type | payload | description |
|---|---|---|
| `state_update` | `{ room_id, state }` | Full game state after any change |
| `error` | `{ message }` | Error notification |
| `chat_message` | `{ sender, content, timestamp }` | Chat broadcast |
| `player_joined` | `{ player }` | Player entered room |
| `player_left` | `{ player_id }` | Player left room |
| `game_start` | `{ initial_state }` | Game started |
| `game_over` | `{ winner, scores }` | Game ended |

## Authentication

- **REST**: `Authorization: Bearer <jwt_token>` header
- **WebSocket**: JWT token passed as `?token=<jwt_token>` query parameter on connection
- **Storage**: Client stores token in localStorage under `pokeface_token`
- **Expiry**: Token expires server-side; client must re-login on 401

## Pagination

### Request
```
GET /api/resource?page=1&pageSize=20
```

### Defaults
- `page`: 1 (1-indexed)
- `pageSize`: 20

### Response
```json
{
  "data": [...],
  "total": 100,
  "page": 1,
  "pageSize": 20
}
```

## Rate Limiting

- **Auth endpoints** (`/api/auth/*`): 10 requests/min per IP
- **Admin endpoints** (`/api/admin/*`): 60 requests/min per IP
- **Game endpoints** (WebSocket): n/a (connection-based)
- Response on limit: HTTP 429 + `Retry-After` header
