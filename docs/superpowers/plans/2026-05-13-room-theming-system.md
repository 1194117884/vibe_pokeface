# Room Theming System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add theme system (room themes, character avatars, card styles) + polish room page with background, NPC walking, and proper nickname display.

**Architecture:** Config-driven theming via React context + CSS variables. Room themes are plain config objects; RoomThemeProvider applies them as CSS custom properties. Character styles map user IDs to avatars. Server exposes nickname + character_id in player data.

**Tech Stack:** Next.js 16, React 19, Tailwind CSS v4, TypeScript, Go 1.26

---

### Task 1: Theme types + registry + RoomThemeProvider

**Files:**
- Create: `frontend/themes/types.ts`
- Create: `frontend/themes/registry.ts`
- Create: `frontend/themes/RoomThemeProvider.tsx`

- [ ] **Step 1: Create types.ts**

```typescript
export interface RoomTheme {
  id: string;
  name: string;
  background: {
    image: string;
    color: string;
    overlay?: string;
  };
  table: {
    feltColor: string;
    feltTexture?: string;
    borderColor: string;
    borderWidth: string;
    decoration: string;
    shadow: string;
  };
  ambient: {
    enabled: boolean;
    npcSprites?: string[];
    npcCount?: number;
  };
  cardStyleId: string;
}

export interface CharacterStyle {
  id: string;
  name: string;
  emoji: string;
  backgroundColor: string;
  borderColor: string;
}

export interface CardStyle {
  id: string;
  name: string;
  backColor: string;
  backPattern?: string;
  suitColors: {
    hearts: string;
    diamonds: string;
    clubs: string;
    spades: string;
  };
}
```

- [ ] **Step 2: Create registry.ts**

```typescript
import { RoomTheme, CharacterStyle, CardStyle } from "./types";

export const roomThemes: Record<string, RoomTheme> = {};
export const characterStyles: Record<string, CharacterStyle> = {};
export const cardStyles: Record<string, CardStyle> = {};

export function registerRoomTheme(theme: RoomTheme) {
  roomThemes[theme.id] = theme;
}

export function registerCharacterStyle(style: CharacterStyle) {
  characterStyles[style.id] = style;
}

export function registerCardStyle(style: CardStyle) {
  cardStyles[style.id] = style;
}

export function getRoomTheme(id: string): RoomTheme {
  return roomThemes[id] || roomThemes["classic-poker"];
}

export function getCharacterStyle(id: string): CharacterStyle {
  return characterStyles[id] || characterStyles["panda"];
}

export function getCardStyle(id: string): CardStyle {
  return cardStyles[id] || cardStyles["classic"];
}
```

- [ ] **Step 3: Create RoomThemeProvider.tsx**

```typescript
"use client";

import { createContext, useContext, useEffect, ReactNode } from "react";
import { RoomTheme } from "./types";
import { getRoomTheme } from "./registry";

const RoomThemeContext = createContext<RoomTheme | null>(null);

export function useRoomTheme(): RoomTheme {
  const ctx = useContext(RoomThemeContext);
  if (!ctx) throw new Error("useRoomTheme must be used within RoomThemeProvider");
  return ctx;
}

interface RoomThemeProviderProps {
  themeId: string;
  children: ReactNode;
}

export function RoomThemeProvider({ themeId, children }: RoomThemeProviderProps) {
  const theme = getRoomTheme(themeId);

  useEffect(() => {
    const root = document.documentElement;
    root.style.setProperty("--bg-image", `url(${theme.background.image})`);
    root.style.setProperty("--bg-color", theme.background.color);
    root.style.setProperty("--bg-overlay", theme.background.overlay || "none");
    root.style.setProperty("--felt-color", theme.table.feltColor);
    root.style.setProperty("--felt-shadow", theme.table.shadow);
    root.style.setProperty("--table-border-color", theme.table.borderColor);
    root.style.setProperty("--table-border-width", theme.table.borderWidth);
    root.style.setProperty("--table-decoration", theme.table.decoration);
    return () => {
      root.style.removeProperty("--bg-image");
      root.style.removeProperty("--bg-color");
      root.style.removeProperty("--bg-overlay");
      root.style.removeProperty("--felt-color");
      root.style.removeProperty("--felt-shadow");
      root.style.removeProperty("--table-border-color");
      root.style.removeProperty("--table-border-width");
      root.style.removeProperty("--table-decoration");
    };
  }, [theme]);

  return (
    <RoomThemeContext.Provider value={theme}>
      {children}
    </RoomThemeContext.Provider>
  );
}
```

- [ ] **Step 4: Create CharacterProvider.tsx**

```typescript
"use client";

import { createContext, useContext, ReactNode } from "react";
import { CharacterStyle } from "./types";
import { getCharacterStyle } from "./registry";

const CharacterContext = createContext<CharacterStyle | null>(null);

export function useCharacterStyle(): CharacterStyle {
  const ctx = useContext(CharacterContext);
  if (!ctx) return getCharacterStyle("panda");
  return ctx;
}

interface CharacterProviderProps {
  characterId: string;
  children: ReactNode;
}

export function CharacterProvider({ characterId, children }: CharacterProviderProps) {
  const style = getCharacterStyle(characterId);
  return (
    <CharacterContext.Provider value={style}>
      {children}
    </CharacterContext.Provider>
  );
}
```

- [ ] **Step 5: Create barrel export**

**Create:** `frontend/themes/index.ts`
```typescript
export * from "./types";
export * from "./registry";
export { RoomThemeProvider, useRoomTheme } from "./RoomThemeProvider";
export { CharacterProvider, useCharacterStyle } from "./CharacterProvider";
```

- [ ] **Step 6: Commit**

```bash
git add frontend/themes/ && git commit -m "feat(theme): add theme types, registry, and React providers"
```

---

### Task 2: Create room theme configs

**Files:**
- Create: `frontend/themes/room/classic-poker.ts`
- Create: `frontend/themes/room/teahouse.ts`
- Create: `frontend/themes/room/modern-lounge.ts`

- [ ] **Step 1: Create classic-poker.ts**

```typescript
import { RoomTheme } from "../types";
import { registerRoomTheme } from "../registry";

const theme: RoomTheme = {
  id: "classic-poker",
  name: "Classic Poker Room",
  background: {
    image: "/images/themes/classic-poker/bg.jpg",
    color: "#1a1a2e",
    overlay: "linear-gradient(rgba(0,0,0,0.6), rgba(0,0,0,0.4))",
  },
  table: {
    feltColor: "#1B5E20",
    borderColor: "#8B4513",
    borderWidth: "8px",
    decoration: "🃏",
    shadow: "0 20px 60px rgba(0,0,0,0.5)",
  },
  ambient: {
    enabled: true,
    npcSprites: ["/images/themes/classic-poker/npc-waiter.png"],
    npcCount: 2,
  },
  cardStyleId: "classic",
};

registerRoomTheme(theme);
```

- [ ] **Step 2: Create teahouse.ts**

```typescript
import { RoomTheme } from "../types";
import { registerRoomTheme } from "../registry";

const theme: RoomTheme = {
  id: "teahouse",
  name: "Chinese Teahouse",
  background: {
    image: "/images/themes/teahouse/bg.jpg",
    color: "#2D1B0E",
    overlay: "linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.3))",
  },
  table: {
    feltColor: "#1a4731",
    borderColor: "#8B6914",
    borderWidth: "8px",
    decoration: "🏮",
    shadow: "0 20px 60px rgba(0,0,0,0.5)",
  },
  ambient: {
    enabled: true,
    npcSprites: ["/images/themes/teahouse/npc-server.png"],
    npcCount: 2,
  },
  cardStyleId: "classic",
};

registerRoomTheme(theme);
```

- [ ] **Step 3: Create modern-lounge.ts**

```typescript
import { RoomTheme } from "../types";
import { registerRoomTheme } from "../registry";

const theme: RoomTheme = {
  id: "modern-lounge",
  name: "Modern Lounge",
  background: {
    image: "/images/themes/modern-lounge/bg.jpg",
    color: "#0f0f23",
    overlay: "linear-gradient(rgba(0,0,0,0.5), rgba(15,15,35,0.4))",
  },
  table: {
    feltColor: "#0d47a1",
    borderColor: "#1a237e",
    borderWidth: "6px",
    decoration: "♦",
    shadow: "0 20px 60px rgba(0,0,0,0.5), 0 0 30px rgba(13,71,161,0.3)",
  },
  ambient: {
    enabled: true,
    npcSprites: ["/images/themes/modern-lounge/npc-host.png"],
    npcCount: 1,
  },
  cardStyleId: "classic",
};

registerRoomTheme(theme);
```

- [ ] **Step 4: Import themes in registry**

Edit `frontend/themes/registry.ts` — add imports at top:
```typescript
import "./room/classic-poker";
import "./room/teahouse";
import "./room/modern-lounge";
```

- [ ] **Step 5: Commit**

```bash
git add frontend/themes/ && git commit -m "feat(theme): add 3 room theme configs"
```

---

### Task 3: Create avatar character styles

**Files:**
- Create: `frontend/themes/character/avatars.ts`

- [ ] **Step 1: Create avatars.ts**

```typescript
import { CharacterStyle } from "../types";
import { registerCharacterStyle } from "../registry";

const avatars: CharacterStyle[] = [
  { id: "panda",    name: "Panda",    emoji: "🐼", backgroundColor: "#374151", borderColor: "#4B5563" },
  { id: "fox",      name: "Fox",      emoji: "🦊", backgroundColor: "#991B1B", borderColor: "#DC2626" },
  { id: "tiger",    name: "Tiger",    emoji: "🐯", backgroundColor: "#92400E", borderColor: "#D97706" },
  { id: "rabbit",   name: "Rabbit",   emoji: "🐰", backgroundColor: "#6B21A8", borderColor: "#9333EA" },
  { id: "phoenix",  name: "Phoenix",  emoji: "🦅", backgroundColor: "#1E40AF", borderColor: "#2563EB" },
  { id: "dragon",   name: "Dragon",   emoji: "🐉", backgroundColor: "#047857", borderColor: "#059669" },
];

avatars.forEach((a) => registerCharacterStyle(a));
```

- [ ] **Step 2: Import in registry.ts**

Add to `frontend/themes/registry.ts`:
```typescript
import "./character/avatars";
```

- [ ] **Step 3: Commit**

```bash
git add frontend/themes/ && git commit -m "feat(theme): add avatar character styles"
```

---

### Task 4: Create card style configs

**Files:**
- Create: `frontend/themes/card/classic.ts`

- [ ] **Step 1: Create classic.ts**

```typescript
import { CardStyle } from "../types";
import { registerCardStyle } from "../registry";

registerCardStyle({
  id: "classic",
  name: "Classic",
  backColor: "#1E3932",
  suitColors: {
    hearts: "#DC2626",
    diamonds: "#DC2626",
    clubs: "#111827",
    spades: "#111827",
  },
});
```

- [ ] **Step 2: Import in registry.ts**

Add to `frontend/themes/registry.ts`:
```typescript
import "./card/classic";
```

- [ ] **Step 3: Commit**

```bash
git add frontend/themes/ && git commit -m "feat(theme): add card style configs"
```

---

### Task 5: Server — add nickname + character_id to player data

**Files:**
- Modify: `server/internal/game/room.go` (playerList, PlayerSession, AddPlayer, FillWithBot)

- [ ] **Step 1: Add Nickname and CharacterID to PlayerSession**

```go
type PlayerSession struct {
    UserID      string
    PlayerID    int64
    Seat        int
    Conn        chan []byte
    Ready       bool
    IsBot       bool
    Nickname    string // display name
    CharacterID string // avatar character ID
}
```

- [ ] **Step 2: Update playerList() to include nickname + character_id**

```go
func (r *GameRoom) playerList() []map[string]interface{} {
    ownerID := r.OwnerID()
    list := make([]map[string]interface{}, len(r.Players))
    for i, p := range r.Players {
        entry := map[string]interface{}{
            "user_id":  p.UserID,
            "seat":     p.Seat,
            "ready":    p.Ready,
            "is_bot":   p.IsBot,
            "is_owner": p.UserID == ownerID,
            "nickname": p.Nickname,
        }
        if !p.IsBot {
            entry["character_id"] = p.CharacterID
        }
        list[i] = entry
    }
    return list
}
```

- [ ] **Step 3: Set nickname/character on AddPlayer**

The Hub needs to pass user info when adding a player. The cleanest path: pass a callback or store reference to RoomManager.

**Add to GameRoom struct:**
```go
type GameRoom struct {
    // ...existing fields...
    userLookup func(userID string) (nickname string, characterID string)
}
```

**In handleJoinRoom (ws/handler.go), pass user data when joining:**
```go
// Fetch user data
var userNickname string
var userCharacterID string
if userID, err := strconv.ParseInt(client.ID, 10, 64); err == nil {
    if user, err := h.UserStore.FindByID(context.Background(), userID); err == nil {
        userNickname = user.Nickname
        if user.CharacterID != nil {
            userCharacterID = *user.CharacterID
        }
    }
}
room.AddPlayerWithInfo(client.ID, userNickname, userCharacterID, client.Send)
```

- [ ] **Step 4: Add AddPlayerWithInfo method**

```go
func (r *GameRoom) AddPlayerWithInfo(userID string, nickname string, characterID string, conn chan []byte) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    // ... same logic as AddPlayer but set Nickname + CharacterID...
    player := &PlayerSession{
        UserID:      userID,
        Seat:        seat,
        Conn:        conn,
        Nickname:    nickname,
        CharacterID: characterID,
    }
    // ...rest same...
}
```

- [ ] **Step 5: Update FillWithBot / AddBot to set bot nickname**

```go
// In FillWithBot:
bot := &PlayerSession{
    UserID:   botID,
    Seat:     seat,
    Conn:     conn,
    IsBot:    true,
    Ready:    true,
    Nickname: fmt.Sprintf("AI %d", nextN), // or get from character config
}
```

- [ ] **Step 6: Update ws/handler.go to pass user info on join**

- [ ] **Step 7: Commit**

```bash
git add server/internal/game/room.go server/internal/api/ws/handler.go && git commit -m "feat(server): add nickname and character_id to player data"
```

---

### Task 6: Server — add theme selection to rooms

**Files:**
- Modify: `server/internal/game/room.go`

- [ ] **Step 1: Add Theme field to GameRoom**

```go
type GameRoom struct {
    // ...existing fields...
    Theme string `json:"theme"` // room theme ID
}
```

Default to "classic-poker" in `NewGameRoom`.

- [ ] **Step 2: Add theme to join response**

In the `player_joined` broadcast, include the room theme:
```go
r.broadcastMsg("player_joined", map[string]interface{}{
    "user_id": userID,
    "seat":    seat,
    "players": r.playerList(),
    "theme":   r.Theme,
})
```

- [ ] **Step 3: Add SetTheme handler in ws/handler.go**

```go
case "set_theme":
    h.handleSetTheme(client, msg)
```

```go
func (h *Hub) handleSetTheme(client *Client, msg C2SMessage) {
    var themeData struct {
        Theme string `json:"theme"`
    }
    if err := json.Unmarshal(msg.Data, &themeData); err != nil {
        return
    }
    room := h.RoomManager.GetRoom(client.RoomID)
    if room != nil {
        if err := room.SetTheme(client.ID, themeData.Theme); err != nil {
            errMsg, _ := json.Marshal(S2CMessage{Type: "error", Data: err.Error()})
            select {
            case client.Send <- errMsg:
            default:
            }
        }
    }
}
```

- [ ] **Step 4: Add SetTheme method to GameRoom**

```go
func (r *GameRoom) SetTheme(userID string, themeID string) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    if r.OwnerID() != userID {
        return fmt.Errorf("only the room owner can change the theme")
    }
    r.Theme = themeID
    r.broadcastMsg("theme_changed", map[string]interface{}{
        "theme": themeID,
    })
    return nil
}
```

- [ ] **Step 5: Commit**

```bash
git add server/internal/game/room.go server/internal/api/ws/handler.go && git commit -m "feat(server): add room theme selection"
```

---

### Task 7: Update RoomTable to use CSS variables

**Files:**
- Modify: `frontend/components/game/RoomTable.tsx`

- [ ] **Step 1: Replace hardcoded colors with CSS variables**

```typescript
export function RoomTable({ players, myUserId, mySeat, phase, onSitDown, onChangeSeat, onAddBot }: RoomTableProps) {
  return (
    <div className="relative w-full max-w-3xl mx-auto aspect-[4/3]">
      {/* Theme-aware felt table */}
      <div
        className="absolute inset-0 rounded-[40%] shadow-xl"
        style={{
          backgroundColor: "var(--felt-color, #1B5E20)",
          boxShadow: "var(--felt-shadow, 0 20px 60px rgba(0,0,0,0.5))",
          borderWidth: "var(--table-border-width, 8px)",
          borderStyle: "solid",
          borderColor: "var(--table-border-color, #8B4513)",
        }}
      >
        <div
          className="absolute inset-4 rounded-[35%]"
          style={{ backgroundColor: "var(--felt-color, #1B5E20)", opacity: 0.3 }}
        />
      </div>

      {/* Center decoration */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-center">
        <span className="text-4xl" style={{ filter: "drop-shadow(0 2px 4px rgba(0,0,0,0.5))" }}>
          {themeDecoration}
        </span>
        <p className="text-white/60 text-sm font-medium mt-1">
          {phase === "playing" ? "游戏中" : phase === "ended" ? "已结束" : "等待中"}
        </p>
      </div>
      {/* ...seats same layout... */}
    </div>
  );
}
```

- [ ] **Step 2: Use useRoomTheme for decoration**

```typescript
import { useRoomTheme } from "@/themes";

// Inside component:
const theme = useRoomTheme();
// Then use theme.table.decoration for the center
```

- [ ] **Step 3: Commit**

```bash
git add frontend/components/game/RoomTable.tsx && git commit -m "feat(theme): make RoomTable use CSS variables from theme"
```

---

### Task 8: Add NPC walking animation component

**Files:**
- Create: `frontend/components/game/NPCWalker.tsx`

- [ ] **Step 1: Create NPCWalker component**

```typescript
"use client";

import { useEffect, useState } from "react";
import { useRoomTheme } from "@/themes";

interface NPC {
  id: number;
  sprite: string;
  direction: "left" | "right";
  y: number; // vertical position (%)
  speed: number; // animation duration (s)
  delay: number; // start delay (s)
}

export function NPCWalker() {
  const theme = useRoomTheme();
  const [npcs, setNpcs] = useState<NPC[]>([]);

  useEffect(() => {
    if (!theme.ambient.enabled || !theme.ambient.npcSprites?.length) return;

    const count = theme.ambient.npcCount || 2;
    const sprites = theme.ambient.npcSprites;

    const generated: NPC[] = Array.from({ length: count }, (_, i) => ({
      id: i,
      sprite: sprites[i % sprites.length],
      direction: Math.random() > 0.5 ? "left" : "right",
      y: 20 + Math.random() * 60,
      speed: 12 + Math.random() * 8,
      delay: i * (5 + Math.random() * 5),
    }));

    setNpcs(generated);
  }, [theme]);

  return (
    <div className="fixed inset-0 pointer-events-none overflow-hidden z-10">
      {npcs.map((npc) => (
        <div
          key={npc.id}
          className="absolute opacity-30"
          style={{
            top: `${npc.y}%`,
            animation: `npc-walk-${npc.direction} ${npc.speed}s ${npc.delay}s linear infinite`,
            fontSize: "2rem",
          }}
        >
          🚶
        </div>
      ))}
    </div>
  );
}
```

- [ ] **Step 2: Add keyframes to globals.css**

Add to `frontend/app/globals.css`:
```css
@keyframes npc-walk-left {
  0% {
    transform: translateX(110vw) scaleX(1);
    opacity: 0;
  }
  10% {
    opacity: 0.3;
  }
  90% {
    opacity: 0.3;
  }
  100% {
    transform: translateX(-10vw) scaleX(1);
    opacity: 0;
  }
}

@keyframes npc-walk-right {
  0% {
    transform: translateX(-10vw) scaleX(-1);
    opacity: 0;
  }
  10% {
    opacity: 0.3;
  }
  90% {
    opacity: 0.3;
  }
  100% {
    transform: translateX(110vw) scaleX(-1);
    opacity: 0;
  }
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/components/game/NPCWalker.tsx frontend/app/globals.css && git commit -m "feat(theme): add NPC walking animation component"
```

---

### Task 9: Update SeatPosition to show avatar + nickname

**Files:**
- Modify: `frontend/components/game/SeatPosition.tsx`
- Modify: `frontend/components/game/RoomTable.tsx` (TablePlayer type)

- [ ] **Step 1: Extend TablePlayer with nickname + characterId**

In `RoomTable.tsx`:
```typescript
export interface TablePlayer {
  userId: string;
  name: string;
  nickname: string;      // <-- add
  characterId: string;   // <-- add
  seat: number;
  isBot: boolean;
  isOwner: boolean;
  isReady: boolean;
  isCurrentTurn?: boolean;
  cardCount: number;
}
```

- [ ] **Step 2: Update SeatPosition to show avatar + nickname**

Replace the avatar circle and name display:
```typescript
// In the player card section:
<div
  className="w-12 h-12 rounded-full flex items-center justify-center text-2xl"
  style={{
    backgroundColor: playerCharStyle.backgroundColor,
    borderColor: playerCharStyle.borderColor,
    borderWidth: "2px",
    borderStyle: "solid",
  }}
>
  {playerCharStyle.emoji}
</div>

<p className="text-sm font-medium text-text-black-strong truncate max-w-[100px]">
  {player.nickname || player.name}
</p>
```

- [ ] **Step 3: Use CharacterProvider in seat rendering**

In RoomTable's `renderSeat`, wrap each seat's character display in a CharacterProvider. Or simpler: just use the character registry directly.

Actually simplest: import `getCharacterStyle` in SeatPosition and use the `characterId` from the player object.

- [ ] **Step 4: Commit**

```bash
git add frontend/components/game/ && git commit -m "feat(theme): update seats to show avatar emoji and nickname"
```

---

### Task 10: Wire everything together in room page

**Files:**
- Modify: `frontend/app/(main)/room/[id]/page.tsx`
- Modify: `frontend/components/game/RoomTable.tsx` (update to pass new fields)

- [ ] **Step 1: Update toTablePlayer to include nickname + characterId**

```typescript
function toTablePlayer(p: ServerPlayer): TablePlayer {
  return {
    userId: String(p.user_id ?? p.userId ?? ""),
    name: String(p.user_id ?? p.userId ?? ""),
    nickname: p.nickname || String(p.user_id ?? p.userId ?? ""),
    characterId: p.character_id || "panda",
    seat: p.seat ?? 0,
    isBot: p.is_bot ?? p.isBot ?? false,
    isOwner: p.is_owner ?? p.isOwner ?? false,
    isReady: p.ready ?? p.isReady ?? false,
    cardCount: Array.isArray(p.hand) ? p.hand.length : (p.card_count ?? p.cardCount ?? 0),
  };
}
```

- [ ] **Step 2: Add theme state to room page**

```typescript
const [roomTheme, setRoomTheme] = useState("classic-poker");
```

Listen for `theme_changed` event:
```typescript
client.on("theme_changed", (msg) => {
  const data = msg.data as ServerData;
  if (data?.theme) setRoomTheme(data.theme);
});
```

Get theme from `player_joined`:
```typescript
client.on("player_joined", (msg) => {
  const data = msg.data as ServerData;
  if (data?.theme) setRoomTheme(data.theme);
  // ...rest
});
```

- [ ] **Step 3: Wrap room content with RoomThemeProvider and NPCWalker**

```typescript
return (
  <RoomThemeProvider themeId={roomTheme}>
    <NPCWalker />
    <div className="min-h-screen flex flex-col relative z-20"
         style={{
           backgroundImage: "var(--bg-image)",
           backgroundColor: "var(--bg-color, #f2f0eb)",
           backgroundSize: "cover",
           backgroundPosition: "center",
         }}>
      {/* add overlay */}
      <div className="fixed inset-0 pointer-events-none"
           style={{ background: "var(--bg-overlay, none)" }} />
      {/* ...rest of page content... */}
    </div>
  </RoomThemeProvider>
);
```

- [ ] **Step 4: Update ServerPlayer interface to include new fields**

```typescript
interface ServerPlayer {
  // ...existing fields...
  nickname?: string;
  character_id?: string;
}
```

- [ ] **Step 5: Commit**

```bash
git add frontend/app/ frontend/components/ && git commit -m "feat(theme): wire room page with theme system, background, and NPCs"
```

---

### Task 11: Verify — run type checks and tests

- [ ] **Step 1: Frontend type check**

```bash
cd frontend && npx tsc --noEmit
```

- [ ] **Step 2: Server build and vet**

```bash
cd server && go vet ./... && go build ./cmd/server/
```

- [ ] **Step 3: Fix any issues**

- [ ] **Step 4: Commit any fixes**

```bash
git commit -m "fix: address type check and build issues"
```
