import { RoomTheme } from "../types";
import { registerRoomTheme } from "../registry";

const theme: RoomTheme = {
  id: "classic-poker",
  name: "Classic Poker Room",
  background: {
    image: "radial-gradient(ellipse at center, #f2f0eb 0%, #edebe9 100%)",
    color: "#f2f0eb",
    overlay: undefined,
  },
  table: {
    feltColor: "#1E3932",
    borderColor: "#8B6914",
    borderWidth: "12px",
    decoration: "🃏",
    shadow:
      "0 0 6px rgba(0,0,0,0.24), 0 8px 12px rgba(0,0,0,0.14), 0 20px 60px rgba(0,0,0,0.3)",
  },
  ambient: {
    enabled: false,
  },
  cardStyleId: "classic",
};

registerRoomTheme(theme);
