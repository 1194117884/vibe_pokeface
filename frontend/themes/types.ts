export interface RoomTheme {
  id: string;
  name: string;
  background: {
    image: string;
    color: string;
    overlay?: string;
  };
  table: {
    feltColor: string;
    feltTexture?: string;
    borderColor: string;
    borderWidth: string;
    decoration: string;
    shadow: string;
  };
  ambient: {
    enabled: boolean;
    npcSprites?: string[];
    npcCount?: number;
  };
  cardStyleId: string;
}

export interface CharacterStyle {
  id: string;
  name: string;
  emoji: string;
  backgroundColor: string;
  borderColor: string;
}

export interface CardStyle {
  id: string;
  name: string;
  backColor: string;
  backPattern?: string;
  suitColors: {
    hearts: string;
    diamonds: string;
    clubs: string;
    spades: string;
  };
}
