# Landlord Visual Indicator Enhancement — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the landlord (地主) visually unmistakable by applying a dark red + gold themed seat card, swapping room owner emoji from 👑 to ⭐, and adding a current-turn bottom badge.

**Architecture:** Pure frontend CSS/Tailwind change in two existing components. `SeatPosition` gets the landlord red theme, owner emoji swap, and turn badge. `PlayerInfo` gets enhanced landlord styling. No server changes — `isLandlord` is already correctly passed through.

**Tech Stack:** React 19, TypeScript, Tailwind CSS v4, clsx

---

## File Structure

| File | Role | Action |
|---|---|---|
| `frontend/components/game/SeatPosition.tsx` | Oval table seat card — landlord theme, owner emoji, turn badge, baodan reposition | **Modify** |
| `frontend/components/game/PlayerInfo.tsx` | Linear game view player info — enhanced landlord styling | **Modify** |

---

### Task 1: Landlord theme + owner emoji + turn badge in SeatPosition

**File:** `frontend/components/game/SeatPosition.tsx`

- [ ] **Step 1: Replace the container clsx with landlord-aware classes**

Replace lines 69-78 (the occupied seat container div's className):

```tsx
// Before (lines 70-77):
<div
  className={clsx(
    "relative flex flex-col items-center gap-2 p-3 rounded-xl border-2 min-w-[120px] transition-all",
    player.isCurrentTurn
      ? "border-yellow-400 bg-yellow-50 shadow-md"
      : "border-ceramic/30 bg-white",
    isMySeat && "ring-2 ring-green-accent/30",
  )}
>

// After:
<div
  className={clsx(
    "relative flex flex-col items-center gap-2 p-3 rounded-xl min-w-[120px] transition-all",
    // Landlord gets red theme; current turn just enhances glow
    player.isLandlord
      ? "border-[2.5px] border-amber-400 bg-gradient-to-br from-red-950 to-red-900 shadow-[0_0_20px_rgba(251,191,36,0.4)]"
      : clsx(
          "border-2",
          player.isCurrentTurn
            ? "border-yellow-400 bg-yellow-50 shadow-md"
            : "border-ceramic/30 bg-white"
        ),
    // Landlord + current turn = even stronger glow
    player.isLandlord && player.isCurrentTurn && "shadow-[0_0_28px_rgba(251,191,36,0.6)]",
    isMySeat && !player.isLandlord && "ring-2 ring-green-accent/30",
  )}
>
```

- [ ] **Step 2: Replace owner badge emoji from 👑 to ⭐**

Replace lines 79-83:

```tsx
// Before:
{player.isOwner && (
  <span className="absolute -top-2 -left-1 text-lg" title="房主">
    👑
  </span>
)}

// After:
{player.isOwner && (
  <span className="absolute -top-2 -left-1 text-base" title="房主">
    ⭐
  </span>
)}
```

- [ ] **Step 3: Replace landlord badge — from corner trophy to top-center gold pill**

Replace lines 84-88:

```tsx
// Before:
{player.isLandlord && (
  <span className="absolute -top-2 -right-1 text-lg" title="地主">
    🏆
  </span>
)}

// After:
{player.isLandlord && (
  <div className="absolute -top-3.5 left-1/2 -translate-x-1/2 bg-amber-400 text-red-950 text-[11px] font-extrabold px-3 py-0.5 rounded tracking-wider whitespace-nowrap">
    👑 地主
  </div>
)}
```

- [ ] **Step 4: Add current turn bottom badge (both landlord and farmer)**

Add after the landlord/owner badge section (after line 88 in original, now after the landlord pill), before the avatar:

```tsx
{player.isCurrentTurn && (
  <div className="absolute -bottom-2 left-1/2 -translate-x-1/2 bg-amber-400 text-amber-950 text-[10px] font-bold px-2.5 py-0.5 rounded-full whitespace-nowrap shadow-[0_0_8px_rgba(250,204,21,0.5)]">
    ⚡ 出牌中
  </div>
)}
```

- [ ] **Step 5: Style avatar for landlord — gold border ring**

Replace lines 90-106 (the avatar div) — only change the avatar wrapper, keeping the emoji/character logic. Wrap the existing avatar with landlord-aware ring:

```tsx
// Change the avatar wrapper div. Lines 90-106 currently:
{player.isBot ? (
  <div className="w-12 h-12 rounded-full bg-purple-400 flex items-center justify-center text-white font-bold text-lg">
    AI
  </div>
) : (
  <div
    className="w-12 h-12 rounded-full flex items-center justify-center text-2xl"
    style={{...}}
  >
    {charStyle?.emoji ?? "🐼"}
  </div>
)}

// Replace the bot avatar div with:
<div className={clsx(
  "w-12 h-12 rounded-full flex items-center justify-center text-white font-bold text-lg",
  player.isLandlord && "ring-[2.5px] ring-amber-400"
)}>
  AI
</div>

// Replace the character avatar div with:
<div
  className={clsx(
    "w-12 h-12 rounded-full flex items-center justify-center text-2xl",
    player.isLandlord && "ring-[2.5px] ring-amber-400"
  )}
  style={{...}}
>
  {charStyle?.emoji ?? "🐼"}
</div>
```

- [ ] **Step 6: Style name and card count — white/gold for landlord**

Replace lines 108-114:

```tsx
// Before:
<p className="text-sm font-medium text-text-black-strong truncate max-w-[100px]">
  {player.nickname || player.name}
</p>

<p className="text-xs text-text-black-soft">
  {player.cardCount} 张牌
</p>

// After:
<p className={clsx(
  "text-sm font-medium truncate max-w-[100px]",
  player.isLandlord ? "text-white" : "text-text-black-strong"
)}>
  {player.nickname || player.name}
</p>

<p className={clsx(
  "text-xs",
  player.isLandlord ? "text-amber-300" : "text-text-black-soft"
)}>
  {player.cardCount} 张牌
</p>
```

- [ ] **Step 7: Reposition baodan/baoshuang for landlord — bottom-left instead of top-right**

Replace lines 131-140:

```tsx
// Before:
{cardsLeft === "baodan" && (
  <span className="absolute -top-2 right-0 bg-red-500 text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold animate-pulse">
    报单
  </span>
)}
{cardsLeft === "baoshuang" && (
  <span className="absolute -top-2 right-0 bg-orange-400 text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold animate-pulse">
    报双
  </span>
)}

// After — move to bottom-left so it doesn't conflict with landlord top-center badge:
const baodanClass = player.isLandlord
  ? "absolute -bottom-2 left-0 bg-red-500 text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold animate-pulse"
  : "absolute -top-2 right-0 bg-red-500 text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold animate-pulse";

{cardsLeft === "baodan" && (
  <span className={baodanClass}>报单</span>
)}
{cardsLeft === "baoshuang" && (
  <span className={player.isLandlord
    ? "absolute -bottom-2 left-0 bg-orange-400 text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold animate-pulse"
    : "absolute -top-2 right-0 bg-orange-400 text-white text-[10px] px-1.5 py-0.5 rounded-full font-bold animate-pulse"
  }>报双</span>
)}
```

- [ ] **Step 8: Update the existing ready badge to be visible on landlord's dark background**

Replace lines 116-120 (ready badge):

```tsx
// Before:
{player.isReady && (
  <span className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded-full">
    ✓ 已准备
  </span>
)}

// After — use slightly different styling when on dark landlord background:
{player.isReady && (
  <span className={clsx(
    "text-xs px-2 py-0.5 rounded-full",
    player.isLandlord
      ? "bg-green-700 text-green-100"
      : "bg-green-100 text-green-700"
  )}>
    ✓ 已准备
  </span>
)}
```

- [ ] **Step 9: Type check + lint**

```bash
cd frontend && npx tsc --noEmit && npm run lint
```

- [ ] **Step 10: Commit**

```bash
git add frontend/components/game/SeatPosition.tsx
git commit -m "feat(ui): add red-theme landlord indicator to SeatPosition"
```

---

### Task 2: Enhanced landlord styling in PlayerInfo

**File:** `frontend/components/game/PlayerInfo.tsx`

- [ ] **Step 1: Add gold border + dark red avatar for landlord opponent; full red theme for self-landlord**

Replace the container div (lines 22-28) and avatar section (lines 29-38):

```tsx
// Before container (lines 22-27):
<div
  className={clsx(
    "flex items-center gap-2 px-3 py-2 rounded-[12px] transition-all duration-200",
    isCurrentTurn && "bg-green-accent/10 ring-2 ring-green-accent shadow-card",
    isSelf ? "flex-row" : "flex-row-reverse"
  )}
>

// After — landlord styling depends on isSelf:
<div
  className={clsx(
    "flex items-center gap-2 px-3 py-2 rounded-[12px] transition-all duration-200",
    // Self + landlord: full red theme
    isSelf && isLandlord && "bg-gradient-to-br from-red-950 to-red-900 border-2 border-amber-400 shadow-[0_0_16px_rgba(251,191,36,0.4)]",
    // Opponent landlord: gold border on white bg
    !isSelf && isLandlord && "bg-white border-2 border-amber-400 shadow-[0_0_8px_rgba(251,191,36,0.2)]",
    // Normal current turn
    !isLandlord && isCurrentTurn && "bg-green-accent/10 ring-2 ring-green-accent shadow-card",
    isSelf ? "flex-row" : "flex-row-reverse"
  )}
>
```

- [ ] **Step 2: Avatar styling — dark red for landlord, keep normal otherwise**

Replace the avatar div (lines 29-38):

```tsx
// Before:
<div className="relative">
  <div className={clsx(
    "w-10 h-10 rounded-full flex items-center justify-center text-sm font-bold text-white shadow-sm",
    isLandlord ? "bg-house-green" : "bg-green-accent"
  )}>
    {name[0]}
  </div>
  {isLandlord && (
    <span className="absolute -top-1 -right-1 text-xs drop-shadow-sm">👑</span>
  )}
</div>

// After — dark red avatar + gold ring for landlord:
<div className="relative">
  <div className={clsx(
    "w-10 h-10 rounded-full flex items-center justify-center text-sm font-bold text-white shadow-sm",
    isLandlord
      ? "bg-gradient-to-br from-red-950 to-red-900 ring-[2px] ring-amber-400"
      : "bg-green-accent"
  )}>
    {name[0]}
  </div>
  {isLandlord && (
    <span className="absolute -top-1 -right-1 text-xs drop-shadow-sm">👑</span>
  )}
</div>
```

- [ ] **Step 3: Name + card count text — white/gold for self-landlord, gold tag for opponent landlord**

Replace the text block (lines 40-45):

```tsx
// Before:
<div className="text-sm">
  <p className="font-semibold text-text-black tracking-tight">{name}</p>
  <p className="text-xs text-text-black-soft tracking-tight">
    {cardCount} cards {isLandlord ? "· Landlord" : ""}
  </p>
</div>

// After:
<div className="text-sm">
  <p className={clsx(
    "font-semibold tracking-tight",
    isSelf && isLandlord ? "text-white" : "text-text-black"
  )}>
    {name}
  </p>
  <p className={clsx(
    "text-xs tracking-tight",
    isSelf && isLandlord
      ? "text-amber-300 font-semibold"
      : isLandlord
        ? "text-amber-700 font-semibold"
        : "text-text-black-soft"
  )}>
    {cardCount} cards
    {isLandlord && (
      <span className={clsx(
        "ml-1 px-1.5 py-0.5 rounded text-[10px] font-bold",
        isSelf
          ? "bg-amber-400 text-red-950"
          : "bg-amber-100 text-amber-800"
      )}>
        地主
      </span>
    )}
  </p>
</div>
```

- [ ] **Step 4: Type check + lint**

```bash
cd frontend && npx tsc --noEmit && npm run lint
```

- [ ] **Step 5: Commit**

```bash
git add frontend/components/game/PlayerInfo.tsx
git commit -m "feat(ui): add red-theme landlord styling to PlayerInfo"
```

---

### Task 3: Visual verification

- [ ] **Step 1: Start the dev server and visually verify**

```bash
cd frontend && npm run dev &
# Open browser to localhost:3000, join a game, verify:
# 1. Landlord seat is red + gold, clearly distinguishable
# 2. Room owner shows ⭐ not 👑
# 3. Current turn shows ⚡ 出牌中 badge
# 4. Landlord + current turn shows both
# 5. Baodan/baoshuang on landlord shows at bottom-left
```

- [ ] **Step 2: Run full frontend checks**

```bash
cd frontend && npx tsc --noEmit && npm run lint && npm test
```
