import { CardStyle } from "../types";
import { registerCardStyle } from "../registry";

registerCardStyle({
  id: "classic",
  name: "Classic",
  backColor: "#1E3932",
  suitColors: {
    hearts: "#DC2626",
    diamonds: "#DC2626",
    clubs: "#111827",
    spades: "#111827",
  },
});
