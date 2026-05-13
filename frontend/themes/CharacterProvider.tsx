"use client";

import { createContext, useContext, ReactNode } from "react";
import { CharacterStyle } from "./types";
import { getCharacterStyle, DEFAULT_CHARACTER_ID } from "./registry";

const CharacterContext = createContext<CharacterStyle | null>(null);

export function useCharacterStyle(): CharacterStyle {
  const ctx = useContext(CharacterContext);
  if (!ctx) {
    if (process.env.NODE_ENV === "development") {
      console.warn("useCharacterStyle used outside CharacterProvider, falling back to default");
    }
    return getCharacterStyle(DEFAULT_CHARACTER_ID);
  }
  return ctx;
}

interface CharacterProviderProps {
  characterId: string;
  children: ReactNode;
}

export function CharacterProvider({ characterId, children }: CharacterProviderProps) {
  const style = getCharacterStyle(characterId);
  return (
    <CharacterContext.Provider value={style}>
      {children}
    </CharacterContext.Provider>
  );
}
