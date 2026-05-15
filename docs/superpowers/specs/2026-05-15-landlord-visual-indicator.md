# Landlord Visual Indicator Enhancement

**Date:** 2026-05-15
**Status:** approved

## Problem

After game starts, players cannot easily identify who the landlord (地主) is. Current indicators are small emoji badges (🏆) identical in size/position to the room owner badge (👑), with no visual change to the seat card itself. The 👑 emoji is counterintuitively used for "room owner" rather than "landlord."

## Design

### Approach: Deep Red Landlord Theme + Gold Crown Label (Option B)

**Identity via color, turn via badge.** The landlord's seat uses a distinctive dark red background with gold border — visually impossible to miss. Current turn is indicated by a bottom badge and glow effect rather than competing with the identity color.

### SeatPosition.tsx Changes

**Landlord seat:**
- Background: `bg-gradient-to-br from-red-950 to-red-900` (dark red gradient)
- Border: `border-amber-400` (gold, 2.5px) with `shadow-[0_0_20px_rgba(251,191,36,0.4)]` glow
- Badge: `"👑 地 主"` in gold pill, positioned top-center (`-top-3.5`, translated)
- Avatar: gold border ring
- Name text: white
- Card count: gold text

**Farmer (non-landlord) seat:**
- Keep existing white background, ceramic border

**Room owner indicator:**
- Change from 👑 to ⭐ (star), same position (`-top-2 -left-1`)

**Current turn indicator (both landlord and farmer):**
- Add bottom badge: `"⚡ 出牌中"` in amber pill, positioned bottom-center
- Landlord: enhanced gold glow shadow (intensity ×1.5 when active)
- Farmer: keep existing yellow border + yellow-50 background, plus bottom badge

**Baodan/Baoshuang compatibility:**
- When landlord is baodan/baoshuang, the badge moves to bottom-left to avoid conflict with the 地主 top badge

### PlayerInfo.tsx Changes

**Landlord player (opponent):**
- Gold border around the player info card
- Dark red avatar background
- Crown emoji on avatar corner
- `"地主"` gold text tag inline with card count

**Landlord player (self):**
- Dark red gradient background on the entire PlayerInfo card
- Gold border
- Gold card count + 地主 tag
- Gold avatar ring

**Non-landlord:** Keep existing styling

**Current turn:**
- Landlord + current turn: enhanced gold glow + `"⚡ 出牌"` badge
- Farmer + current turn: keep existing green ring indicator

### Files Changed

| File | Scope |
|---|---|
| `frontend/components/game/SeatPosition.tsx` | Major: landlord theme, owner emoji swap, turn badge |
| `frontend/components/game/PlayerInfo.tsx` | Moderate: enhanced landlord styling |

No changes to RoomTable.tsx, GameTable.tsx, or server — `isLandlord` is already correctly passed through.

### Edge Cases

- Player is both room owner and landlord: ⭐ top-left (owner) + red theme (landlord) — no visual conflict
- Landlord is baodan/baoshuang: 报单/报双 badge moves to bottom-left
- Landlord AND current turn: red theme stays, glow intensifies, "⚡ 出牌中" badge at bottom
- Bidding phase: landlord not yet assigned — no themes applied until `game_start`
