# Frontend Rules

## Tech Stack

- **Framework**: Next.js 16 with App Router (`app/` directory)
- **Rendering**: React 19 Server Components by default; `"use client"` only when needed
- **Styling**: Tailwind CSS v4 with Starbucks-inspired design system
- **Language**: TypeScript (strict)

## Component Architecture

```
page.tsx (Server Component, data fetching)
  └── <ClientShell> ("use client", WebSocket context)
        ├── <GameTable> (card layout, game state)
        ├── <HandArea> (player's cards, selection)
        ├── <ActionBar> (bid/pass/play buttons)
        ├── <PlayerInfo> (other players, cards count)
        └── <ChatPanel> (chat, emoji, voice)
```

## State Management

- **Game state**: Received via WebSocket `state_update` events. Single source of truth from server.
- **Local UI state**: `useState` — card selection, animation triggers, modal open/close
- **Auth state**: JWT token in localStorage, managed by `lib/api-client.ts`
- **No global state library**: WebSocket events + React context are sufficient

## Styling

- Tailwind utility classes only. No custom CSS files.
- Design tokens: colors from Tailwind config (green palette, warm neutrals), rounded corners, shadows
- Mobile-first responsive: `sm:` → `md:` → `lg:` breakpoints
- Animation: CSS transitions for card moves, CSS keyframes for deal/shuffle effects

## Performance

- Server Components where possible (no client-side JS for static content)
- `React.memo` on card components that re-render frequently during game
- `useCallback` for handler functions passed as props to memoized children
- No unnecessary `useEffect` — prefer event handlers and derived state
