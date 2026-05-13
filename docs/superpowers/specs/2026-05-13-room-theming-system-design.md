# Room Theming System Design

## Overview

Add a theming system to the room page covering:
1. **Room themes** — swappable backgrounds, table visuals, atmosphere
2. **Character styles** — per-user avatar + nickname display
3. **Card styles** — card backs/faces tied to themes

The immediate goal is to make the room page feel like a real game venue (background, table polish, ambient NPCs) with proper avatar+昵称 display on seats. The architecture is data-driven so new themes can be added as config files.

## 1. Theme Architecture (Frontend)

### File Structure

```
frontend/themes/
├── types.ts                 # All theme interfaces
├── registry.ts              # Theme registry + getter
├── RoomThemeProvider.tsx     # React context, applies CSS vars
├── CharacterProvider.tsx     # Character/avatar context
├── room/
│   ├── classic-poker.ts     # Classic poker room theme
│   ├── teahouse.ts          # Chinese teahouse theme
│   └── modern-lounge.ts     # Modern lounge theme
├── character/
│   └── avatars.ts           # Avatar definitions
└── card/
    └── classic.ts           # Classic card style
```

### RoomTheme Type

```typescript
interface RoomTheme {
  id: string;
  name: string;
  background: {
    image: string;       // CSS background-image URL
    color: string;       // fallback background color
    overlay?: string;    // gradient overlay for readability
  };
  table: {
    feltColor: string;
    feltTexture?: string;
    borderColor: string;
    borderWidth: string;
    decoration: string;   // center text/icon
    shadow: string;
  };
  ambient: {
    npcSprites: string[];  // walking NPC image paths
    npcCount: number;       // max NPCs on screen
  };
  cardStyleId: string;      // which card style to use
}
```

### How It Works

1. `RoomThemeProvider` reads a config and sets CSS custom properties on `:root`
2. All UI components use `var(--<name>)` via Tailwind or inline CSS
3. Adding a new theme = adding one config file + assets

CSS variables set by the provider:
- `--bg-image`, `--bg-color`, `--bg-overlay`
- `--felt-color`, `--felt-shadow`, `--table-border`, `--table-decoration`
- `--npc-sprites`, `--npc-count`

The room page stores the selected `themeId` (from room data or URL param). Provider looks it up from registry, applies variables.

## 2. Room Page Polish

### Background
- Full-viewport background image from theme, with a dark gradient overlay so text stays readable
- CSS `background-size: cover`, `background-position: center`
- Falls back to theme color if image hasn't loaded

### Table
- Uses CSS variables for felt, border, shadow instead of hardcoded `bg-green-700`
- Added subtle texture/shadow for depth
- Center decoration from theme config ("斗地主", icon, etc.)

### NPC Walking Animation
- Pure CSS keyframe animations — NPC sprites drift across the viewport
- NPCs fade in at one edge, walk across, fade out at the opposite edge
- Each NPC has random speed, delay, and Y position
- 1-3 NPCs visible at a time based on theme config
- Sprites defined per theme (casino waiters, teahouse servers, etc.)

### Seats
- Show character avatar (colored circle with emoji/icon) + nickname
- "Sit down" / "change seat" still works the same way
- Avatar colors from player's character style, not random

## 3. Character System

### Character Config

```typescript
interface CharacterStyle {
  id: string;
  name: string;           // display name
  emoji: string;          // avatar emoji
  backgroundColor: string;
  borderColor: string;
}
```

Default character set (~6 characters):
- Panda, Fox, Tiger, Rabbit, Phoenix, Dragon
- Each with distinct emoji + color scheme

Users pick a character when they register (defaults to first). Stored as `character_id` on the user record. Sent in player data so the frontend can render the right avatar.

### Server Changes

**playerList() in room.go** — add fields:
- `nickname` (string) — from User store for human players, "AI XXX" for bots
- `character_id` (string) — empty for bots

Need to add `CharacterID` to `PlayerSession` or fetch it from the User store when building the player list. Since the room has access to `model.GameStore`, we need a way to look up user data. Options:
- Pass a UserStore or lookup function to the room
- Or add a `nickname`/`characterId` field to `PlayerSession` that's set on join

**Room model** — add `Theme` field to `GameRoom`:
```go
type GameRoom struct {
    // ... existing fields
    Theme string `json:"theme"` // theme ID selected by room creator
}
```

## 4. Card Styles

```typescript
interface CardStyle {
  id: string;
  name: string;
  backImage: string;
  // Suit color overrides
  suitColors: {
    hearts: string;
    diamonds: string;
    clubs: string;
    spades: string;
  };
}
```

The Card component checks the current theme's `cardStyleId` and applies the matching style. Currently only "classic" exists — red hearts/diamonds, black clubs/spades, standard card back.

## 5. Non-Goals (out of scope)

- Payment/purchasing for themes or characters
- User-generated custom themes
- Animated/NFT card styles
- Theme shop UI
- Avatar customization beyond preset characters

## 6. Implementation Order

1. Create theme types + registry + `RoomThemeProvider`
2. Create 3 room theme configs (classic-poker, teahouse, modern-lounge)
3. Add NPC walking animation component
4. Update `RoomTable` to use CSS variables from theme
5. Create character avatar system + `CharacterProvider`
6. Update `SeatPosition` to show avatar + nickname
7. Server changes: add nickname/character_id to playerList()
8. Server changes: add theme field to room
9. Update room page to wire everything together
10. Create card style configs + update Card component
