import { RoomTheme } from "../types";
import { registerRoomTheme } from "../registry";

const theme: RoomTheme = {
  id: "imperial-emerald",
  name: "Imperial Emerald Arena",
  background: {
    image: "radial-gradient(circle, #1a7452 0%, #063a27 100%)",
    color: "#111415",
    overlay: undefined,
  },
  table: {
    feltColor: "radial-gradient(circle, #1a7452 0%, #063a27 100%)",
    borderColor: "#111415",
    borderWidth: "0",
    decoration: "",
    shadow: "none",
  },
  ambient: {
    enabled: false,
  },
  cardStyleId: "classic",
};

registerRoomTheme(theme);
