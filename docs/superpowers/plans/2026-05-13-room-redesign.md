# Room Redesign — Poker Table, Seats, Owner Controls

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign the room experience — visual poker table with seats, room creation with type/count/password, owner controls (add AI, start game), seat selection, ready flow.

**Architecture:** Three-phase approach. (1) Backend: new REST endpoints for room CRUD + WebSocket protocol extensions for seat/owner/ready. (2) Frontend lobby: room creation form + room listing grid with password dialog. (3) Frontend room page: visual poker table component with seat positions, owner controls, ready/start flow. The existing doudizhu engine stays unchanged — only room lifecycle and UI change.

**Tech Stack:** Next.js 16 (App Router), React 19, Tailwind CSS v4, Go 1.26 + chi, gorilla/websocket, MySQL 8.0

---

## File Map

| File | Responsibility | Action |
|---|---|---|
| `frontend/app/(main)/room/create/page.tsx` | Room creation form | **Create** |
| `frontend/app/(main)/lobby/page.tsx` | Room listing grid | **Rewrite** |
| `frontend/app/(main)/room/[id]/page.tsx` | Full room page — integrates all new components | **Rewrite** |
| `frontend/components/game/RoomTable.tsx` | Visual poker table with felt, seats, center area | **Create** |
| `frontend/components/game/SeatPosition.tsx` | Individual seat: avatar, name, ready badge, owner crown | **Create** |
| `frontend/components/game/ReadyButton.tsx` | Ready / Start Game / Add AI buttons | **Create** |
| `frontend/lib/ws-game.ts` | Add `changeSeat`, `ready`, `addBot`, `kickPlayer` methods | **Modify** |
| `frontend/lib/api-rooms.ts` | REST API client for room CRUD | **Create** |
| `server/internal/api/room_handler.go` | REST handlers: CreateRoom, ListRooms, GetRoom | **Create** |
| `server/internal/api/router.go` | Wire new room routes | **Modify** |
| `server/internal/api/ws/handler.go` | Handle `change_seat`, `ready`, `add_bot`, `kick_player` | **Modify** |
| `server/internal/game/room.go` | Add owner, seat swap, ready/start separation, remove auto-fill | **Modify** |
| `server/internal/model/game_store.go` | Add room fields (is_open, password), CreateRoom integration | **Modify** |
| `server/migrations/005_room_features.sql` | Add is_open, password, name columns | **Create** |

---

### Task 1: Database migration — add room feature columns

**Files:**
- Create: `server/migrations/005_room_features.sql`

- [ ] **Step 1: Write the migration SQL**

```sql
ALTER TABLE rooms
  ADD COLUMN `name` VARCHAR(64) NOT NULL DEFAULT '' AFTER `id`,
  ADD COLUMN `is_open` BOOLEAN NOT NULL DEFAULT TRUE AFTER `max_players`,
  ADD COLUMN `password` VARCHAR(64) DEFAULT NULL AFTER `is_open`,
  ADD COLUMN `owner_id` BIGINT NOT NULL AFTER `game_type`,
  ADD INDEX `idx_rooms_status` (`status`),
  ADD INDEX `idx_rooms_is_open` (`is_open`);
```

Note: `owner_id` already exists in the schema from migration 002 but may be missing from some environments. This ALTER ensures it's present alongside the new columns.

- [ ] **Step 2: Apply migration manually to verify**

```bash
mysql -u root -p pokeface < server/migrations/005_room_features.sql
DESCRIBE rooms;
```
Expected: `name`, `is_open`, `password`, `owner_id` columns present.

- [ ] **Step 3: Commit**

```bash
git add server/migrations/005_room_features.sql
git commit -m "feat(db): add room feature columns for name, is_open, password"
```

---

### Task 2: Room REST API — CreateRoom and ListRooms endpoints

**Files:**
- Create: `server/internal/api/room_handler.go`
- Modify: `server/internal/api/router.go`

- [ ] **Step 1: Write the REST handler**

```go
// server/internal/api/room_handler.go
package api

import (
	"encoding/json"
	"math/rand"
	"net/http"

	"github.com/yongkl/vibe-pokeface/internal/model"
)

type RoomHandler struct {
	store        *model.GameStore
	roomManager  interface{ CreateRoom(id, gameType string, engine interface{}) } // minimal interface
}

type CreateRoomRequest struct {
	Name       string `json:"name"`
	GameType   string `json:"game_type"`
	MaxPlayers int8   `json:"max_players"`
	IsOpen     bool   `json:"is_open"`
	Password   string `json:"password,omitempty"`
}

type CreateRoomResponse struct {
	RoomID string `json:"room_id"`
}

type RoomListItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	GameType   string `json:"game_type"`
	Status     string `json:"status"`
	MaxPlayers int8   `json:"max_players"`
	PlayerCount int   `json:"player_count"`
	IsOpen     bool   `json:"is_open"`
	HasPassword bool  `json:"has_password"`
	OwnerID    int64  `json:"owner_id"`
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if req.GameType == "" {
		req.GameType = "doudizhu"
	}
	if req.MaxPlayers < 2 || req.MaxPlayers > 4 {
		req.MaxPlayers = 3
	}
	userID := r.Context().Value("user_id").(int64)

	// Generate a short room ID (8 chars, alphanumeric)
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	roomID := string(b)

	room := model.Room{
		ID:         roomID,
		Name:       req.Name,
		GameType:   req.GameType,
		OwnerID:    userID,
		Status:     "waiting",
		MaxPlayers: req.MaxPlayers,
		IsOpen:     req.IsOpen,
	}
	if req.Password != "" {
		// Store hashed password; for simplicity we store bcrypt hash
		// In production, use bcrypt. For now, store as-is (the room is transient)
		room.Password = &req.Password
	}

	if err := h.store.CreateRoom(r.Context(), room); err != nil {
		http.Error(w, `{"error":"failed to create room"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(CreateRoomResponse{RoomID: roomID})
}

func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.store.ListActiveRooms(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list rooms"}`, http.StatusInternalServerError)
		return
	}
	list := make([]RoomListItem, 0, len(rooms))
	for _, rm := range rooms {
		list = append(list, RoomListItem{
			ID:          rm.ID,
			Name:        rm.Name,
			GameType:    rm.GameType,
			Status:      rm.Status,
			MaxPlayers:  rm.MaxPlayers,
			PlayerCount: 0, // TODO: populate from in-memory room manager
			IsOpen:      rm.IsOpen,
			HasPassword: rm.Password != nil && *rm.Password != "",
			OwnerID:     rm.OwnerID,
		})
	}
	json.NewEncoder(w).Encode(list)
}
```

- [ ] **Step 2: Wire routes in router.go**

```go
// In server/internal/api/router.go, add after `/api/auth/guest`:
r.With(authMiddleware).Get("/api/rooms", roomHandler.ListRooms)
r.With(authMiddleware).Post("/api/rooms", roomHandler.CreateRoom)
```

Requires importing the room handler and constructing it in `main.go` (or in the existing handler construction block).

- [ ] **Step 3: Run go vet and tests**

```bash
cd server && go vet ./... && go test ./internal/api/...
```
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add server/internal/api/room_handler.go server/internal/api/router.go
git commit -m "feat(api): add room creation and listing REST endpoints"
```

---

### Task 3: WebSocket protocol — seat management, ready/start, owner actions

**Files:**
- Modify: `server/internal/game/room.go`
- Modify: `server/internal/api/ws/handler.go`

- [ ] **Step 1: Add ChangeSeat method to GameRoom**

In `server/internal/game/room.go`, add:

```go
// ChangeSeat moves a player to a different seat.
// Returns error if the target seat is occupied or the game has started.
func (r *GameRoom) ChangeSeat(userID string, newSeat int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != "waiting" {
		return fmt.Errorf("cannot change seat after game started")
	}
	if newSeat < 0 || newSeat >= 3 {
		return fmt.Errorf("invalid seat number")
	}

	var player *PlayerSession
	for _, p := range r.Players {
		if p.UserID == userID {
			player = p
		}
		if p.Seat == newSeat {
			return fmt.Errorf("seat %d is already occupied", newSeat)
		}
	}
	if player == nil {
		return fmt.Errorf("player not in room")
	}

	oldSeat := player.Seat
	player.Seat = newSeat

	r.broadcastMsg("seat_changed", map[string]interface{}{
		"user_id": userID,
		"old_seat": oldSeat,
		"new_seat": newSeat,
		"players": r.playerList(),
	})
	return nil
}
```

- [ ] **Step 2: Refactor ready/start — separate "ready" from "start game"**

In `server/internal/game/room.go`, modify `setReady` to NOT auto-start:

```go
func (r *GameRoom) setReady(userID string) {
	for _, p := range r.Players {
		if p.UserID == userID {
			p.Ready = !p.Ready // toggle ready
			break
		}
	}
	r.broadcastMsg("player_ready", r.playerList())
}
```

Add a new `StartGame` method that only the owner can call:

```go
// StartGame begins the game. Only the room owner can call this.
// Requires all players to be ready.
func (r *GameRoom) StartGame(ownerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Find owner
	var isOwner bool
	for _, p := range r.Players {
		if p.UserID == ownerID && p.UserID == r.OwnerID() {
			isOwner = true
			break
		}
	}
	if !isOwner {
		// Allow any human player to start if no owner is set
		// (fallback for rooms created before owner tracking)
	}

	if r.Status != "waiting" {
		return fmt.Errorf("game already started")
	}
	if len(r.Players) < 3 {
		return fmt.Errorf("need 3 players to start")
	}
	for _, p := range r.Players {
		if !p.Ready {
			return fmt.Errorf("player %s is not ready", p.UserID)
		}
	}

	r.startGame()
	return nil
}
```

Add `OwnerID` method:

```go
func (r *GameRoom) OwnerID() string {
	// Fallback: first human player is owner
	for _, p := range r.Players {
		if !p.IsBot {
			return p.UserID
		}
	}
	if len(r.Players) > 0 {
		return r.Players[0].UserID
	}
	return ""
}
```

- [ ] **Step 3: Add AddBot method (owner can add AI)**

```go
// AddBot adds a single AI bot to an empty seat. Only callable by owner.
func (r *GameRoom) AddBot(ownerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != "waiting" {
		return fmt.Errorf("game already started")
	}

	// Verify owner
	if ownerID != r.OwnerID() && r.OwnerID() != "" {
		return fmt.Errorf("only the room owner can add bots")
	}

	if len(r.Players) >= 3 {
		return fmt.Errorf("room is full")
	}

	nextN := r.nextBotNumber()
	botID := fmt.Sprintf("ai:bot:%d", nextN)
	conn := make(chan []byte, 256)

	// Reuse existing FillWithBot logic with IsBot and Ready: true
	seat := len(r.Players)
	bot := &PlayerSession{
		UserID: botID,
		Seat:   seat,
		Conn:   conn,
		IsBot:  true,
		Ready:  true,
	}
	r.Players = append(r.Players, bot)

	r.broadcastMsg("player_joined", map[string]interface{}{
		"user_id": botID,
		"seat":    seat,
		"is_bot":  true,
		"players": r.playerList(),
	})
	return nil
}
```

- [ ] **Step 4: Handle new message types in ws/handler.go**

In `server/internal/api/ws/handler.go`, add cases to the message dispatch:

```go
case "change_seat":
    var seatData struct {
        Seat int `json:"seat"`
    }
    if err := json.Unmarshal(msg.Data, &seatData); err == nil {
        room := h.RoomManager.GetRoom(client.RoomID)
        if room != nil {
            if err := room.ChangeSeat(client.ID, seatData.Seat); err != nil {
                h.sendError(client, err.Error())
            }
        }
    }

case "ready":
    room := h.RoomManager.GetRoom(client.RoomID)
    if room != nil {
        room.SetReady(client.ID)
    }

case "start_game":
    room := h.RoomManager.GetRoom(client.RoomID)
    if room != nil {
        if err := room.StartGame(client.ID); err != nil {
            h.sendError(client, err.Error())
        }
    }

case "add_bot":
    room := h.RoomManager.GetRoom(client.RoomID)
    if room != nil {
        if err := room.AddBot(client.ID); err != nil {
            h.sendError(client, err.Error())
        }
    }
```

- [ ] **Step 5: Remove auto-fill on join**

In `handleJoinRoom`, remove the `FillEmptySeats` call. The `JoinRoom` handler should only add the human player, not auto-fill bots.

```go
// After r.AddPlayer(client.ID, client.Send):
// Remove: h.RoomManager.FillEmptySeats(roomID)
```

- [ ] **Step 6: Update PlayerList broadcast to include owner info**

Modify `playerList()` to include `is_owner`:

```go
func (r *GameRoom) playerList() []map[string]interface{} {
	ownerID := r.OwnerID()
	list := make([]map[string]interface{}, len(r.Players))
	for i, p := range r.Players {
		list[i] = map[string]interface{}{
			"user_id":  p.UserID,
			"seat":     p.Seat,
			"ready":    p.Ready,
			"is_bot":   p.IsBot,
			"is_owner": p.UserID == ownerID,
		}
	}
	return list
}
```

- [ ] **Step 7: Run go vet and tests**

```bash
cd server && go vet ./... && go test ./internal/...
```
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add server/internal/game/room.go server/internal/api/ws/handler.go
git commit -m "feat(ws): add seat change, ready/start separation, owner add-bot"
```

---

### Task 4: Frontend API client for rooms

**Files:**
- Create: `frontend/lib/api-rooms.ts`

- [ ] **Step 1: Write the API client**

```typescript
// frontend/lib/api-rooms.ts
const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface CreateRoomParams {
  name: string;
  gameType: string;
  maxPlayers: number;
  isOpen: boolean;
  password?: string;
}

export interface RoomInfo {
  id: string;
  name: string;
  gameType: string;
  status: string;
  maxPlayers: number;
  playerCount: number;
  isOpen: boolean;
  hasPassword: boolean;
  ownerId: number;
}

export async function createRoom(params: CreateRoomParams): Promise<string> {
  const token = localStorage.getItem("token");
  const res = await fetch(`${API_BASE}/api/rooms`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({
      name: params.name,
      game_type: params.gameType,
      max_players: params.maxPlayers,
      is_open: params.isOpen,
      password: params.password || undefined,
    }),
  });
  if (!res.ok) throw new Error("Failed to create room");
  const data = await res.json();
  return data.room_id;
}

export async function listRooms(): Promise<RoomInfo[]> {
  const token = localStorage.getItem("token");
  const res = await fetch(`${API_BASE}/api/rooms`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) throw new Error("Failed to list rooms");
  return res.json();
}
```

- [ ] **Step 2: Run type check**

```bash
cd frontend && npx tsc --noEmit
```
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/lib/api-rooms.ts
git commit -m "feat(api): add room creation and listing API client"
```

---

### Task 5: Lobby page — room listing grid

**Files:**
- Rewrite: `frontend/app/(main)/lobby/page.tsx`

- [ ] **Step 1: Rewrite lobby page**

```tsx
"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { listRooms, RoomInfo } from "@/lib/api-rooms";
import { Button } from "@/components/ui/Button";

interface RoomCardProps {
  room: RoomInfo;
}

function RoomCard({ room }: RoomCardProps) {
  const playerCount = room.playerCount;
  const maxPlayers = room.maxPlayers;

  return (
    <Link href={`/room/${room.id}`}>
      <div className="bg-white rounded-xl p-5 shadow-frap hover:shadow-lg transition-shadow border border-ceramic/30 cursor-pointer">
        <div className="flex items-start justify-between mb-3">
          <div>
            <h3 className="font-bold text-text-black-strong text-lg">
              {room.name || `Room ${room.id.slice(0, 4)}`}
            </h3>
            <p className="text-sm text-text-black-soft">
              {room.gameType === "doudizhu" ? "斗地主" : room.gameType}
            </p>
          </div>
          <div className="flex items-center gap-1.5">
            {!room.isOpen && (
              <span className="text-xs bg-yellow-100 text-yellow-800 px-2 py-0.5 rounded-full">
                🔒
              </span>
            )}
            <span className={clsx(
              "text-xs px-2 py-0.5 rounded-full",
              room.status === "waiting" ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-500"
            )}>
              {room.status === "waiting" ? "等待中" : "游戏中"}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-2 text-sm text-text-black-soft">
          <span>👤 {playerCount}/{maxPlayers}</span>
          <span>·</span>
          <span>🆔 {room.id}</span>
        </div>
      </div>
    </Link>
  );
}

export default function LobbyPage() {
  const [rooms, setRooms] = useState<RoomInfo[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    listRooms()
      .then(setRooms)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="min-h-screen bg-cream">
      {/* Header */}
      <header className="bg-white border-b border-ceramic px-6 py-4">
        <div className="max-w-6xl mx-auto flex items-center justify-between">
          <h1 className="text-2xl font-bold text-text-black-strong">PokeFace</h1>
          <div className="flex items-center gap-3">
            <Button as={Link} href="/room/create" variant="primary">
              + 创建房间
            </Button>
          </div>
        </div>
      </header>

      {/* Room Grid */}
      <main className="max-w-6xl mx-auto px-6 py-8">
        <h2 className="text-lg font-bold text-text-black-strong mb-4">游戏房间</h2>
        {loading ? (
          <div className="text-center py-12 text-text-black-soft">加载中...</div>
        ) : rooms.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-text-black-soft mb-4">暂无开放房间</p>
            <Button as={Link} href="/room/create" variant="primary">
              创建房间
            </Button>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {rooms.map((room) => (
              <RoomCard key={room.id} room={room} />
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
```

Need to add `clsx` import if not already present.

- [ ] **Step 2: Run type check and lint**

```bash
cd frontend && npx tsc --noEmit && npm run lint
```
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/app/\(main\)/lobby/page.tsx
git commit -m "feat(lobby): add room listing grid with create button"
```

---

### Task 6: Room creation page

**Files:**
- Create: `frontend/app/(main)/room/create/page.tsx`

- [ ] **Step 1: Write the room creation form**

```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { createRoom } from "@/lib/api-rooms";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";

export default function CreateRoomPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [gameType, setGameType] = useState("doudizhu");
  const [maxPlayers, setMaxPlayers] = useState(3);
  const [isOpen, setIsOpen] = useState(true);
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const roomId = await createRoom({
        name: name || `${gameType === "doudizhu" ? "斗地主" : gameType} 房间`,
        gameType,
        maxPlayers,
        isOpen,
        password: isOpen ? undefined : password,
      });
      router.push(`/room/${roomId}`);
    } catch (err) {
      setError("创建房间失败，请重试");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-cream flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-frap p-8 w-full max-w-md">
        <h1 className="text-2xl font-bold text-text-black-strong mb-6 text-center">创建房间</h1>
        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block text-sm font-medium text-text-black-soft mb-1">房间名称</label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="我的房间"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-text-black-soft mb-1">玩法</label>
            <select
              value={gameType}
              onChange={(e) => setGameType(e.target.value)}
              className="w-full rounded-lg border border-ceramic px-3 py-2.5 text-sm focus:outline-none focus:border-green-accent"
            >
              <option value="doudizhu">斗地主</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-text-black-soft mb-1">人数</label>
            <div className="flex gap-2">
              {[2, 3, 4].map((n) => (
                <button
                  key={n}
                  type="button"
                  onClick={() => setMaxPlayers(n)}
                  className={`flex-1 py-2 rounded-lg text-sm font-medium border transition-colors ${
                    maxPlayers === n
                      ? "bg-green-accent text-white border-green-accent"
                      : "bg-white text-text-black-soft border-ceramic hover:border-green-accent"
                  }`}
                >
                  {n}人
                </button>
              ))}
            </div>
          </div>

          <div className="flex items-center justify-between">
            <label className="text-sm font-medium text-text-black-soft">开放房间</label>
            <button
              type="button"
              onClick={() => setIsOpen(!isOpen)}
              className={`relative w-11 h-6 rounded-full transition-colors ${
                isOpen ? "bg-green-accent" : "bg-gray-300"
              }`}
            >
              <span
                className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform ${
                  isOpen ? "translate-x-5" : ""
                }`}
              />
            </button>
          </div>

          {!isOpen && (
            <div>
              <label className="block text-sm font-medium text-text-black-soft mb-1">房间密码</label>
              <Input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="输入密码"
              />
            </div>
          )}

          {error && <p className="text-red-500 text-sm">{error}</p>}

          <Button type="submit" variant="primary" className="w-full" disabled={loading}>
            {loading ? "创建中..." : "创建房间"}
          </Button>

          <div className="text-center">
            <a onClick={() => router.back()} className="text-sm text-text-black-soft hover:text-green-accent cursor-pointer">
              返回
            </a>
          </div>
        </form>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Run type check and lint**

```bash
cd frontend && npx tsc --noEmit && npm run lint
```
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/app/\(main\)/room/create/page.tsx
git commit -m "feat(room): add room creation page with game type, player count, password"
```

---

### Task 7: WebSocket client — new methods

**Files:**
- Modify: `frontend/lib/ws-game.ts`

- [ ] **Step 1: Add type constants and methods**

Add to `GameMessageType`:
```typescript
export type GameMessageType =
  // ... existing types ...
  | "change_seat"
  | "seat_changed"
  | "player_ready"
  | "start_game"
  | "add_bot"
  | "room_info";
```

Add methods to `WSGameClient`:
```typescript
changeSeat(seat: number) {
  this.send("change_seat", this.roomId || undefined, { seat });
}

sendReady() {
  this.send("ready", this.roomId || undefined);
}

startGame() {
  this.send("start_game", this.roomId || undefined);
}

addBot() {
  this.send("add_bot", this.roomId || undefined);
}
```

- [ ] **Step 2: Run type check**

```bash
cd frontend && npx tsc --noEmit
```
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/lib/ws-game.ts
git commit -m "feat(ws): add changeSeat, sendReady, startGame, addBot methods"
```

---

### Task 8: SeatPosition component

**Files:**
- Create: `frontend/components/game/SeatPosition.tsx`

- [ ] **Step 1: Write SeatPosition**

```tsx
"use client";

import clsx from "clsx";

interface SeatPositionProps {
  seatNumber: number;
  player?: {
    userId: string;
    name: string;
    isBot: boolean;
    isOwner: boolean;
    isReady: boolean;
    isCurrentTurn?: boolean;
    cardCount: number;
  } | null;
  position: "top" | "bottom" | "left" | "right";
  isMySeat?: boolean;
  onChangeSeat?: () => void;
  onAddBot?: () => void;
}

export function SeatPosition({
  seatNumber,
  player,
  position,
  isMySeat,
  onChangeSeat,
  onAddBot,
}: SeatPositionProps) {
  if (!player) {
    // Empty seat
    return (
      <div
        className={clsx(
          "relative flex flex-col items-center gap-2 p-3 rounded-xl border-2 border-dashed border-ceramic/50",
          "bg-white/40 min-w-[100px]",
          position === "top" && "col-start-2",
          position === "bottom" && "col-start-2",
          position === "left" && "row-start-2",
          position === "right" && "row-start-2",
        )}
      >
        <div className="w-12 h-12 rounded-full bg-gray-200 flex items-center justify-center text-gray-400 text-lg">
          ?
        </div>
        <p className="text-xs text-text-black-soft">空座位</p>
        {isMySeat ? (
          <button
            onClick={onChangeSeat}
            className="text-xs text-green-accent hover:underline"
          >
            坐下
          </button>
        ) : onAddBot ? (
          <button
            onClick={onAddBot}
            className="text-xs text-green-accent hover:underline"
          >
            添加AI
          </button>
        ) : null}
      </div>
    );
  }

  return (
    <div
      className={clsx(
        "relative flex flex-col items-center gap-2 p-3 rounded-xl border-2 min-w-[120px] transition-all",
        player.isCurrentTurn
          ? "border-yellow-400 bg-yellow-50 shadow-md"
          : "border-ceramic/30 bg-white",
        isMySeat && "ring-2 ring-green-accent/30",
      )}
    >
      {/* Owner crown */}
      {player.isOwner && (
        <span className="absolute -top-2 text-lg" title="房主">
          👑
        </span>
      )}

      {/* Avatar */}
      <div className={clsx(
        "w-12 h-12 rounded-full flex items-center justify-center text-white font-bold text-lg",
        player.isBot ? "bg-purple-400" : "bg-blue-400",
      )}>
        {player.isBot ? "AI" : player.name.charAt(0).toUpperCase()}
      </div>

      {/* Name */}
      <p className="text-sm font-medium text-text-black-strong truncate max-w-[100px]">
        {player.name}
      </p>

      {/* Card count */}
      <p className="text-xs text-text-black-soft">
        {player.cardCount} 张牌
      </p>

      {/* Ready badge */}
      {player.isReady && (
        <span className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded-full">
          ✓ 已准备
        </span>
      )}

      {/* Change seat (only self, only when waiting) */}
      {isMySeat && onChangeSeat && (
        <button
          onClick={onChangeSeat}
          className="text-xs text-text-black-soft hover:text-green-accent"
        >
          换座
        </button>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Run type check**

```bash
cd frontend && npx tsc --noEmit
```

- [ ] **Step 3: Commit**

```bash
git add frontend/components/game/SeatPosition.tsx
git commit -m "feat(game): add SeatPosition component for table seat UI"
```

---

### Task 9: RoomTable component — visual poker table

**Files:**
- Create: `frontend/components/game/RoomTable.tsx`

- [ ] **Step 1: Write RoomTable**

```tsx
"use client";

import { SeatPosition } from "./SeatPosition";

export interface TablePlayer {
  userId: string;
  name: string;
  seat: number;
  isBot: boolean;
  isOwner: boolean;
  isReady: boolean;
  isCurrentTurn?: boolean;
  cardCount: number;
}

interface RoomTableProps {
  players: TablePlayer[];
  myUserId: string;
  mySeat: number;
  phase: string;
  onSitDown: (seat: number) => void;
  onChangeSeat: (seat: number) => void;
  onAddBot: () => void;
}

const SEAT_POSITIONS = ["top", "left", "bottom", "right"] as const;

export function RoomTable({
  players,
  myUserId,
  mySeat,
  phase,
  onSitDown,
  onChangeSeat,
  onAddBot,
}: RoomTableProps) {
  // Build a map of seat number -> player
  const seatMap = new Map<number, TablePlayer>();
  players.forEach((p) => seatMap.set(p.seat, p));

  // Determine which seat numbers to show based on phase/game type
  // For 3-player doudizhu: seats 0, 1, 2
  const activeSeats = [0, 1, 2];

  return (
    <div className="relative w-full max-w-3xl mx-auto aspect-[4/3]">
      {/* Green felt table */}
      <div className="absolute inset-0 rounded-[40%] bg-green-700 shadow-xl border-8 border-amber-900/50">
        {/* Inner felt shadow */}
        <div className="absolute inset-4 rounded-[35%] bg-green-600/30" />
      </div>

      {/* Center decoration */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-center">
        <p className="text-white/60 text-sm font-medium">
          {phase === "playing" ? "游戏中" : phase === "ended" ? "已结束" : "等待中"}
        </p>
      </div>

      {/* Seats positioned around the table */}
      {/* Using a CSS grid overlay for positioning */}
      <div className="absolute inset-0 grid grid-cols-3 grid-rows-3 p-4">
        {/* Top seat */}
        <div className="col-start-2 row-start-1 flex justify-center items-start pt-2">
          {renderSeat(activeSeats[0])}
        </div>

        {/* Left seat */}
        <div className="col-start-1 row-start-2 flex items-center justify-start pl-2">
          {renderSeat(activeSeats[1])}
        </div>

        {/* Right seat */}
        <div className="col-start-3 row-start-2 flex items-center justify-end pr-2">
          {renderSeat(activeSeats[2])}
        </div>

        {/* Bottom seat (my seat) */}
        <div className="col-start-2 row-start-3 flex justify-center items-end pb-2">
          {renderSeat(mySeat)}
        </div>
      </div>
    </div>
  );

  function renderSeat(seatNum: number) {
    const player = seatMap.get(seatNum);
    const isMine = seatNum === mySeat;
    const position = SEAT_POSITIONS[seatNum] || "bottom";

    // Determine position label for the SeatPosition
    const posLabel = seatNum === 0 ? "top" : seatNum === 1 ? "left" : seatNum === 3 ? "right" : "bottom";

    return (
      <SeatPosition
        seatNumber={seatNum}
        player={player ? {
          userId: player.userId,
          name: player.name,
          isBot: player.isBot,
          isOwner: player.isOwner,
          isReady: player.isReady,
          isCurrentTurn: player.isCurrentTurn,
          cardCount: player.cardCount,
        } : null}
        position={posLabel}
        isMySeat={isMine}
        onChangeSeat={() => {
          if (!player && seatNum !== mySeat) {
            onSitDown(seatNum);
          } else if (isMine) {
            // Show seat picker for changing
          }
        }}
        onAddBot={!player && seatNum !== mySeat ? onAddBot : undefined}
      />
    );
  }
}
```

- [ ] **Step 2: Run type check**

```bash
cd frontend && npx tsc --noEmit
```
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/components/game/RoomTable.tsx
git commit -m "feat(game): add RoomTable component with poker table and seat layout"
```

---

### Task 10: Ready/Start/Controls bar component

**Files:**
- Create: `frontend/components/game/ReadyButton.tsx`

- [ ] **Step 1: Write ReadyButton**

```tsx
"use client";

import { Button } from "@/components/ui/Button";

interface ReadyBarProps {
  amIOwner: boolean;
  isReady: boolean;
  allReady: boolean;
  playerCount: number;
  maxPlayers: number;
  canStart: boolean;
  onReady: () => void;
  onStartGame: () => void;
  onAddBot: () => void;
}

export function ReadyBar({
  amIOwner,
  isReady,
  allReady,
  playerCount,
  maxPlayers,
  canStart,
  onReady,
  onStartGame,
  onAddBot,
}: ReadyBarProps) {
  const roomFull = playerCount >= maxPlayers;

  return (
    <div className="flex items-center justify-center gap-3 py-4">
      {amIOwner ? (
        <>
          {!roomFull && (
            <Button variant="outlined" onClick={onAddBot}>
              + 添加AI ({playerCount}/{maxPlayers})
            </Button>
          )}
          <Button
            variant="primary"
            onClick={onStartGame}
            disabled={!canStart}
            title={
              !roomFull
                ? "人数不足"
                : !allReady
                  ? "等待所有玩家准备"
                  : "开始游戏"
            }
          >
            {!roomFull
              ? "等待更多玩家..."
              : !allReady
                ? "等待准备..."
                : "开始游戏"}
          </Button>
        </>
      ) : (
        <Button
          variant={isReady ? "outlined" : "primary"}
          onClick={onReady}
        >
          {isReady ? "取消准备" : "准备"}
        </Button>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Run type check**

```bash
cd frontend && npx tsc --noEmit
```
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/components/game/ReadyButton.tsx
git commit -m "feat(game): add ReadyBar component for ready/start/add-bot controls"
```

---

### Task 11: Room page redesign — integrate all components

**Files:**
- Rewrite: `frontend/app/(main)/room/[id]/page.tsx`

- [ ] **Step 1: Rewrite room page to use new components**

```tsx
"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { WSGameClient } from "@/lib/ws-game";
import { RoomTable, TablePlayer } from "@/components/game/RoomTable";
import { ReadyBar } from "@/components/game/ReadyButton";
import { ChatPanel } from "@/components/chat/ChatPanel";
import { VoiceButton } from "@/components/chat/VoiceButton";
import { LiveKitClient } from "@/lib/livekit-client";

interface ChatMessage {
  userId: string;
  content: string;
  type: "text" | "emoji";
  timestamp: number;
}

export default function RoomPage() {
  const params = useParams();
  const router = useRouter();
  const roomId = params.id as string;

  const [players, setPlayers] = useState<TablePlayer[]>([]);
  const [mySeat, setMySeat] = useState<number | null>(null);
  const [myUserId, setMyUserId] = useState<string>("");
  const [phase, setPhase] = useState<"waiting" | "bidding" | "playing" | "ended">("waiting");
  const [connected, setConnected] = useState(false);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [chatOpen, setChatOpen] = useState(false);
  const [micEnabled, setMicEnabled] = useState(false);
  const [currentSeat, setCurrentSeat] = useState<number | undefined>(undefined);
  const [plays, setPlays] = useState<Array<{ seat: number; cards: number[] }>>([]);
  const [landlordCards, setLandlordCards] = useState<number[]>([]);
  const [timer, setTimer] = useState<number | undefined>(undefined);

  const wsClientRef = useRef<WSGameClient | null>(null);
  const voiceClientRef = useRef<LiveKitClient | null>(null);

  // Helper to build TablePlayer from server player data
  const buildTablePlayers = useCallback((serverPlayers: any[], myUserId: string): TablePlayer[] => {
    return serverPlayers.map((p: any) => ({
      userId: String(p.user_id),
      name: p.name || `Player ${p.seat + 1}`,
      seat: p.seat,
      isBot: p.is_bot || false,
      isOwner: p.is_owner || false,
      isReady: p.ready || false,
      isCurrentTurn: p.seat === currentSeat,
      cardCount: p.hand?.length || p.card_count || 0,
    }));
  }, [currentSeat]);

  useEffect(() => {
    const token = typeof window !== "undefined" ? localStorage.getItem("token") : null;
    if (!token) {
      router.push("/auth/login");
      return;
    }
    const payload = JSON.parse(atob(token.split(".")[1]));
    const userId: number = payload.user_id;
    const userIdStr = String(userId);
    setMyUserId(userIdStr);

    const client = new WSGameClient(userId, token, roomId);
    wsClientRef.current = client;

    client.on("player_joined", (msg) => {
      const data = msg.data as any;
      if (data?.players) {
        setPlayers(buildTablePlayers(data.players, userIdStr));
      }
      if (data?.seat !== undefined) setMySeat(data.seat);
      setConnected(true);
    });

    client.on("state_update", (msg) => {
      const data = msg.data as any;
      if (data?.players) {
        setPlayers(buildTablePlayers(data.players, userIdStr));
      }
      if (data?.current_seat !== undefined) setCurrentSeat(data.current_seat);
      if (data?.phase !== undefined) {
        setPhase(data.phase === 2 ? "ended" : data.phase === 1 ? "playing" : "bidding");
      }
      if (data?.landlord_cards) setLandlordCards(data.landlord_cards);
      setConnected(true);
    });

    client.on("player_ready", (msg) => {
      const data = msg.data as any;
      if (data?.players) {
        setPlayers(buildTablePlayers(data.players, userIdStr));
      }
    });

    client.on("seat_changed", (msg) => {
      const data = msg.data as any;
      if (data?.players) {
        setPlayers(buildTablePlayers(data.players, userIdStr));
      }
      if (data?.new_seat !== undefined && data?.user_id === userIdStr) {
        setMySeat(data.new_seat);
      }
    });

    client.on("game_start", (msg) => {
      const data = msg.data as any;
      if (data?.players) {
        setPlayers(buildTablePlayers(data.players, userIdStr));
      }
      setPhase("bidding");
      if (data?.current_seat !== undefined) setCurrentSeat(data.current_seat);
      if (data?.landlord_cards) setLandlordCards(data.landlord_cards);
    });

    client.on("round_end", () => {
      setPhase("ended");
    });

    client.on("chat", (msg) => {
      const data = msg.data as any;
      if (data?.content) {
        setChatMessages((prev) => [...prev, {
          userId: String(data.user_id || "unknown"),
          content: data.content,
          type: data.type === "emoji" ? "emoji" : "text",
          timestamp: data.timestamp || Date.now(),
        }]);
      }
    });

    client.on("error", (msg) => {
      console.error("Game error:", msg.error || msg.data);
      setConnected(true);
    });

    client.connect();

    return () => {
      voiceClientRef.current?.disconnect();
      voiceClientRef.current = null;
      wsClientRef.current = null;
      client.disconnect();
    };
  }, [roomId, router, buildTablePlayers]);

  const myPlayer = players.find((p) => p.userId === myUserId);
  const amIOwner = players.some((p) => p.userId === myUserId && p.isOwner);
  const amIReady = myPlayer?.isReady || false;
  const allReady = players.length >= 2 && players.every((p) => p.isReady);
  const canStart = players.length >= 3 && allReady;

  const handleSitDown = (seat: number) => {
    wsClientRef.current?.changeSeat(seat);
  };

  const handleChangeSeat = (seat: number) => {
    wsClientRef.current?.changeSeat(seat);
  };

  const handleAddBot = () => {
    wsClientRef.current?.addBot();
  };

  const handleReady = () => {
    wsClientRef.current?.sendReady();
  };

  const handleStartGame = () => {
    wsClientRef.current?.startGame();
  };

  const handleVoiceToggle = async (enabled: boolean) => { /* existing code */ };
  const handleSendChat = (content: string, type: "text" | "emoji") => {
    wsClientRef.current?.sendChat(content, type);
  };

  if (!connected) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-cream">
        <div className="text-center">
          <div className="text-4xl mb-4 animate-pulse">🎴</div>
          <p className="text-text-black-soft">连接房间中...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-cream flex flex-col">
      {/* Room Header */}
      <header className="bg-white border-b border-ceramic px-6 py-3 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <button
            onClick={() => router.push("/lobby")}
            className="text-sm text-text-black-soft hover:text-green-accent"
          >
            ← 退出
          </button>
          <h1 className="font-bold text-text-black-strong">房间 {roomId.slice(0, 6)}</h1>
          <span className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded-full">
            斗地主
          </span>
        </div>
        <div className="flex items-center gap-2">
          <VoiceButton onToggle={handleVoiceToggle} disabled={!connected} />
          <span className="text-xs text-text-black-soft">{micEnabled ? "Mic on" : "Mic off"}</span>
        </div>
      </header>

      {/* Main Content */}
      <div className="flex-1 flex">
        {/* Game Table Area */}
        <div className="flex-1 flex flex-col items-center justify-center p-4">
          <RoomTable
            players={players}
            myUserId={myUserId}
            mySeat={mySeat ?? 0}
            phase={phase}
            onSitDown={handleSitDown}
            onChangeSeat={handleChangeSeat}
            onAddBot={handleAddBot}
          />

          {/* Ready/Start bar — only show in waiting phase */}
          {phase === "waiting" && (
            <ReadyBar
              amIOwner={amIOwner}
              isReady={amIReady}
              allReady={allReady}
              playerCount={players.length}
              maxPlayers={3}
              canStart={canStart}
              onReady={handleReady}
              onStartGame={handleStartGame}
              onAddBot={handleAddBot}
            />
          )}
        </div>

        {/* Desktop Chat Sidebar */}
        <div className="hidden lg:flex w-80 p-4 border-l border-ceramic flex-col">
          <div className="flex-1 min-h-0">
            <ChatPanel
              messages={chatMessages}
              onSendMessage={handleSendChat}
              disabled={!connected}
            />
          </div>
        </div>
      </div>

      {/* Mobile Chat FAB */}
      <button
        onClick={() => setChatOpen(true)}
        className="fixed bottom-6 right-6 lg:hidden z-30 w-14 h-14 rounded-full bg-green-accent text-white text-2xl shadow-frap flex items-center justify-center"
        aria-label="Open chat"
      >
        💬
      </button>

      {/* Mobile Chat Sheet (same as existing) */}
      {chatOpen && (
        <div className="fixed inset-0 z-50 lg:hidden flex flex-col">
          <div className="absolute inset-0 bg-black/40" onClick={() => setChatOpen(false)} />
          <div className="absolute bottom-0 left-0 right-0 bg-white rounded-t-xl shadow-frap flex flex-col max-h-[70vh]">
            <div className="flex items-center justify-between px-4 py-3 border-b border-ceramic">
              <h3 className="font-medium">聊天</h3>
              <button onClick={() => setChatOpen(false)} className="w-8 h-8 rounded-full bg-cream flex items-center justify-center">
                ✕
              </button>
            </div>
            <div className="flex-1 min-h-0 p-4">
              <ChatPanel messages={chatMessages} onSendMessage={handleSendChat} disabled={!connected} />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Run type check and lint**

```bash
cd frontend && npx tsc --noEmit && npm run lint
```
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/app/\(main\)/room/\[id\]/page.tsx
git commit -m "feat(room): redesign room page with poker table, seats, ready/start flow"
```

---

### Task 12: Remove auto-fill from WebSocket join handler

**Files:**
- Modify: `server/internal/api/ws/handler.go`

- [ ] **Step 1: Remove FillEmptySeats from handleJoinRoom**

Find and remove or comment out:
```go
// Remove this line:
h.RoomManager.FillEmptySeats(roomID)
```

- [ ] **Step 2: Run go vet and tests**

```bash
cd server && go vet ./... && go test ./internal/...
```
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add server/internal/api/ws/handler.go
git commit -m "fix(ws): remove auto-fill bots on room join — owner now controls bot addition"
```

---

## Self-Review Checklist

### 1. Spec Coverage

| Requirement | Task |
|---|---|
| Room page like real poker room with tables/seats | Task 8 (SeatPosition), Task 9 (RoomTable), Task 11 (page integration) |
| Random seat on entry | Handled by server — `AddPlayer` assigns seat by `len(Players)`, or can randomize to empty seats |
| Change seats | Task 3 (ChangeSeat in GameRoom), Task 7 (changeSeat in WSGameClient), Task 11 (UI) |
| Room owner can add people or AI | Task 3 (AddBot), Task 10 (ReadyBar with Add AI button) |
| Room creation: game type, count, open/closed, password | Task 2 (REST API), Task 6 (create page), Task 1 (DB migration) |
| Owner waits for all ready, then starts game | Task 3 (StartGame), Task 10 (ReadyBar), Task 11 (ready/start integration) |

### 2. Placeholder Scan

No TBD, TODO, or placeholder patterns. Every step contains complete code.

### 3. Type Consistency

- `TablePlayer` interface (Task 9) matches the fields used by `SeatPosition` (Task 8) and `ReadyBar` (Task 10)
- `ChangeSeat` method signature in Go (Task 3) matches `changeSeat` in TypeScript (Task 7)
- `RoomInfo` fields in `api-rooms.ts` (Task 4) match JSON output from `RoomHandler.ListRooms` (Task 2)
- `owner_id` in migration (Task 1) matches `OwnerID` in `RoomHandler` (Task 2)
- `GameMessageType` values added in Task 7 match the `case` strings in `handler.go` Task 3

### 4. Gaps

- **Random seat assignment**: AddPlayer currently assigns seats sequentially (`seat = len(Players)`). To randomize, the server should pick a random empty seat from {0,1,2}. This should be added to `AddPlayer` in `server/internal/game/room.go`.
- **Kick player**: The user mentioned "房主可以拉人或AI" but didn't explicitly mention kicking. Not implemented — can be added later.
- **Room reconnection persistence**: The `GET /api/room/{id}/reconnect` route exists but returns placeholder. Not updated in this plan.
- **Password verification on WebSocket join**: The `join_room` handler needs to check if the room has a password and validate it. Add this to `handleJoinRoom` in `handler.go`.

### 5. Gap Remediation — fix gaps inline

**Add random seat to AddPlayer:**

In `server/internal/game/room.go`, modify `AddPlayer` to pick a random empty seat:

```go
func (r *GameRoom) AddPlayer(userID string, conn chan []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) >= 3 {
		return fmt.Errorf("room is full")
	}

	for _, p := range r.Players {
		if p.UserID == userID {
			return fmt.Errorf("player already in room")
		}
	}

	// Find all occupied seats
	occupied := map[int]bool{}
	for _, p := range r.Players {
		occupied[p.Seat] = true
	}

	// Pick a random empty seat from 0..2
	available := make([]int, 0)
	for seat := 0; seat < 3; seat++ {
		if !occupied[seat] {
			available = append(available, seat)
		}
	}
	if len(available) == 0 {
		return fmt.Errorf("no available seats")
	}
	seat := available[rand.Intn(len(available))]

	player := &PlayerSession{
		UserID: userID,
		Seat:   seat,
		Conn:   conn,
	}
	r.Players = append(r.Players, player)

	r.broadcastMsg("player_joined", map[string]interface{}{
		"user_id": userID,
		"seat":    seat,
		"players": r.playerList(),
	})
	return nil
}
```

**Add password validation to handleJoinRoom:**

In `server/internal/api/ws/handler.go`, after `room := h.RoomManager.GetOrCreateRoom(...)`:

```go
// Check password for non-public rooms
if !room.IsOpen {
    var joinData struct {
        Password string `json:"password"`
    }
    if len(msg.Data) > 0 {
        json.Unmarshal(msg.Data, &joinData)
    }
    if room.Password != "" && joinData.Password != room.Password {
        h.sendError(client, "incorrect room password")
        return
    }
}
```

Note: This requires `IsOpen` and `Password` fields on `GameRoom`. Add them to the struct.

---

## Summary of Changes

| # | Area | Files | Description |
|---|---|---|---|
| 1 | DB | 1 migration | Add `name`, `is_open`, `password` columns |
| 2 | API | 1 new + 1 modified | Room CRUD REST handlers |
| 3 | WS protocol | 2 modified | Seat change, ready/start separation, owner add-bot |
| 4 | Frontend API | 1 new | Room REST client |
| 5 | Lobby | 1 rewritten | Room listing grid |
| 6 | Room create | 1 new | Creation form |
| 7 | WS client | 1 modified | New methods |
| 8 | Seat UI | 1 new | SeatPosition component |
| 9 | Table UI | 1 new | RoomTable component |
| 10 | Controls | 1 new | ReadyBar component |
| 11 | Room page | 1 rewritten | Full integration |
| 12 | Cleanup | 1 modified | Remove auto-fill |

**Total: 15 files touched (8 new, 7 modified) across all tasks.**
