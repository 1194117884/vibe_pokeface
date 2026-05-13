"use client";

import { createContext, useContext, ReactNode } from "react";
import { CharacterStyle } from "./types";
import { getCharacterStyle } from "./registry";

const CharacterContext = createContext<CharacterStyle | null>(null);

export function useCharacterStyle(): CharacterStyle {
  const ctx = useContext(CharacterContext);
  if (!ctx) return getCharacterStyle("panda");
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
