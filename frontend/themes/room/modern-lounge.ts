import { RoomTheme } from "../types";
import { registerRoomTheme } from "../registry";

const theme: RoomTheme = {
  id: "modern-lounge",
  name: "Modern Lounge",
  background: {
    image: "radial-gradient(ellipse at center, #0f0f23 0%, #060612 100%)",
    color: "#060612",
    overlay: "linear-gradient(rgba(0,0,0,0.4), rgba(0,0,0,0.2))",
  },
  table: {
    feltColor: "#0d47a1",
    borderColor: "#1a237e",
    borderWidth: "8px",
    decoration: "♦",
    shadow: "0 20px 60px rgba(0,0,0,0.7)",
  },
  ambient: {
    enabled: true,
    npcSprites: ["🚶", "🚶‍♂️", "🚶‍♀️"],
    npcCount: 2,
  },
  cardStyleId: "classic",
};

registerRoomTheme(theme);
