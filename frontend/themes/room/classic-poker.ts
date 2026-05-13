import { RoomTheme } from "../types";
import { registerRoomTheme } from "../registry";

const theme: RoomTheme = {
  id: "classic-poker",
  name: "Classic Poker Room",
  background: {
    image: "radial-gradient(ellipse at center, #2d1b0e 0%, #1a1a2e 100%)",
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
    npcSprites: ["🚶", "🚶‍♂️", "🚶‍♀️"],
    npcCount: 2,
  },
  cardStyleId: "classic",
};

registerRoomTheme(theme);
