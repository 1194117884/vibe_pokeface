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
