# Mobile & PWA Design Adaptation

## Overview

This document extends the Starbucks-inspired DESIGN.md with mobile-first and PWA considerations for the PokeFace game app. Unlike Starbucks' content-heavy site, PokeFace is a real-time multiplayer card game, so mobile UX priorities differ: **one-handed play, chat as overlay not sidebar, and instant launch from home screen.**

## PWA Configuration

- **Theme Color**: `#1E3932` (House Green) — matches the admin sidebar and dark brand surfaces
- **Background Color**: `#f2f0eb` (Cream) — matches the page canvas
- **Display**: `standalone` — full-screen app-like experience
- **Icon**: Playing-card SVG with brand green background
- **Start URL**: `/auth/login` (redirects authenticated users via token check)

## Breakpoint Strategy

Adapting the DESIGN.md breakpoints to our game app:

| Name | Width | Layout |
|------|-------|--------|
| **Phone** | < 640px | Single column, full-width cards, bottom-sheet chat |
| **Tablet** | 640–1023px | Two-column game layout, side panel stacks |
| **Desktop** | 1024px+ | Full game table + side chat panel |

## Key Mobile Adaptations

### 1. Safe Area Support

All major containers use `env(safe-area-inset-*)` padding to avoid notches, status bars, and home indicators on modern phones.

### 2. Auth Pages (Login / Register)

- Card goes full-width with `mx-4` margins rather than `max-w-md` centered
- Input fields increase touch target to minimum 44px height
- Buttons remain pill-shaped but with increased padding on mobile for tap comfort
- No changes to visual hierarchy — auth is already minimal and card-based

### 3. Game Room (the critical mobile surface)

- **Desktop**: Current layout — game table left, chat panel `w-80` right
- **Mobile**: Game table fills screen. Chat collapses to a **slide-up bottom sheet** triggered by a chat FAB button
- The bottom sheet follows the same card design language (`12px` radius top, white bg, cream border)
- Voice button shrinks to a compact `40px` circle in the game controls area

### 4. Lobby Page

- Header collapses to single-line brand + sign-out icon
- "Create Room" button stays prominent as a full-width pill below the title on phone
- Empty state icon scales down slightly but keeps the centered card treatment

### 5. Admin Pages

- Sidebar collapses to a **hamburger drawer** on phone (< 768px)
- Drawer overlays the content with a backdrop, matching the House-Green sidebar aesthetic
- Table views become horizontally scrollable on narrow screens
- Stat cards go to single-column grid on phone

### 6. Touch & Interaction

- All interactive targets minimum `44px` height (WCAG compliance)
- `active:scale-[0.95]` preserved as the signature micro-interaction
- No hover-dependent states (hover doesn't exist on touch)
- Increased tap padding on pill buttons: `10px 20px` vs `7px 16px` on mobile

### 7. Theme Color Meta

- Status bar matches House Green (`#1E3932`) on mobile browsers
- PWA splash screen uses Cream (`#f2f0eb`) background with green icon

## Implementation Checklist

- [ ] PWA manifest.json with icons
- [ ] SVG app icons (multi-resolution)
- [ ] Viewport meta tags for PWA
- [ ] Safe-area-inset CSS variables
- [ ] Apple-touch-icon and msapplication meta
- [ ] Service worker registration (optional offline support)
- [ ] Admin sidebar mobile collapse
- [ ] Room page chat bottom sheet
- [ ] Auth pages touch target sizing
- [ ] Lobby header responsive
- [ ] Admin tables horizontal scroll on mobile

## Icon Specification

Single SVG icon (playing card + brand) serves as the source for all resolutions:

- **192x192**: Standard PWA icon
- **512x512**: Splash screen icon
- **180x180**: Apple touch icon
- **favicon.ico**: 32x32 derived from the SVG

The icon is a white playing card with a spade/card symbol centered on a House Green (`#1E3932`) rounded rectangle.
