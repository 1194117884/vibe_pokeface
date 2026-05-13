export * from "./types";
export * from "./registry";
export { RoomThemeProvider, useRoomTheme } from "./RoomThemeProvider";
export { CharacterProvider, useCharacterStyle } from "./CharacterProvider";

// Side-effect imports must come AFTER all exports so registry.ts
// fully evaluates before theme configs try to call register*().
import "./room/classic-poker";
import "./room/teahouse";
import "./room/modern-lounge";
import "./character/avatars";
import "./card/classic";
