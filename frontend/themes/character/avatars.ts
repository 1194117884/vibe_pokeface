import { CharacterStyle } from "../types";
import { registerCharacterStyle } from "../registry";

const avatars: CharacterStyle[] = [
  { id: "panda",    name: "Panda",    emoji: "🐼", backgroundColor: "#374151", borderColor: "#4B5563" },
  { id: "fox",      name: "Fox",      emoji: "🦊", backgroundColor: "#991B1B", borderColor: "#DC2626" },
  { id: "tiger",    name: "Tiger",    emoji: "🐯", backgroundColor: "#92400E", borderColor: "#D97706" },
  { id: "rabbit",   name: "Rabbit",   emoji: "🐰", backgroundColor: "#6B21A8", borderColor: "#9333EA" },
  { id: "phoenix",  name: "Phoenix",  emoji: "🦅", backgroundColor: "#1E40AF", borderColor: "#2563EB" },
  { id: "dragon",   name: "Dragon",   emoji: "🐉", backgroundColor: "#047857", borderColor: "#059669" },
];

avatars.forEach((a) => registerCharacterStyle(a));
