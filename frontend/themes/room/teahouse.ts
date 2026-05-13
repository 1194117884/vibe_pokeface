import { RoomTheme } from "../types";
import { registerRoomTheme } from "../registry";

const theme: RoomTheme = {
  id: "teahouse",
  name: "Teahouse",
  background: {
    image: "radial-gradient(ellipse at center, #3d2010 0%, #1a0e05 100%)",
    color: "#1a0e05",
    overlay: "linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.3))",
  },
  table: {
    feltColor: "#1a4731",
    borderColor: "#8B6914",
    borderWidth: "8px",
    decoration: "🏮",
    shadow: "0 20px 60px rgba(0,0,0,0.6)",
  },
  ambient: {
    enabled: true,
    npcSprites: ["🚶", "🚶‍♂️", "🚶‍♀️"],
    npcCount: 2,
  },
  cardStyleId: "classic",
};

registerRoomTheme(theme);
