import type { RoomTheme, CharacterStyle, CardStyle } from "./types";

export const DEFAULT_ROOM_THEME_ID = "classic-poker";
export const DEFAULT_CHARACTER_ID = "panda";
export const DEFAULT_CARD_STYLE_ID = "classic";

const roomThemes: Record<string, RoomTheme> = {};
const characterStyles: Record<string, CharacterStyle> = {};
const cardStyles: Record<string, CardStyle> = {};

const DEFAULT_ROOM_THEME: RoomTheme = {
  id: DEFAULT_ROOM_THEME_ID,
  name: "Default",
  background: { image: "", color: "#f2f0eb" },
  table: {
    feltColor: "#1E3932",
    borderColor: "#8B6914",
    borderWidth: "12px",
    decoration: "",
    shadow:
      "0 0 6px rgba(0,0,0,0.24), 0 8px 12px rgba(0,0,0,0.14), 0 20px 60px rgba(0,0,0,0.3)",
  },
  ambient: { enabled: false },
  cardStyleId: DEFAULT_CARD_STYLE_ID,
};

const DEFAULT_CHARACTER_STYLE: CharacterStyle = {
  id: DEFAULT_CHARACTER_ID,
  name: "Panda",
  emoji: "🐼",
  backgroundColor: "#374151",
  borderColor: "#4B5563",
};

const DEFAULT_CARD_STYLE: CardStyle = {
  id: DEFAULT_CARD_STYLE_ID,
  name: "Classic",
  backColor: "#1E3932",
  suitColors: { hearts: "#DC2626", diamonds: "#DC2626", clubs: "#111827", spades: "#111827" },
};

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
  return roomThemes[id] || roomThemes[DEFAULT_ROOM_THEME_ID] || DEFAULT_ROOM_THEME;
}

export function getCharacterStyle(id: string): CharacterStyle {
  return characterStyles[id] || characterStyles[DEFAULT_CHARACTER_ID] || DEFAULT_CHARACTER_STYLE;
}

export function getCardStyle(id: string): CardStyle {
  return cardStyles[id] || cardStyles[DEFAULT_CARD_STYLE_ID] || DEFAULT_CARD_STYLE;
}
