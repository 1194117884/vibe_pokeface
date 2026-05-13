"use client";

import { createContext, useContext, useEffect, ReactNode } from "react";
import { RoomTheme } from "./types";
import { getRoomTheme } from "./registry";

const RoomThemeContext = createContext<RoomTheme | null>(null);

export function useRoomTheme(): RoomTheme {
  const ctx = useContext(RoomThemeContext);
  if (!ctx) throw new Error("useRoomTheme must be used within RoomThemeProvider");
  return ctx;
}

interface RoomThemeProviderProps {
  themeId: string;
  children: ReactNode;
}

export function RoomThemeProvider({ themeId, children }: RoomThemeProviderProps) {
  const theme = getRoomTheme(themeId);

  useEffect(() => {
    const root = document.documentElement;
    root.style.setProperty("--bg-image", `url(${theme.background.image})`);
    root.style.setProperty("--bg-color", theme.background.color);
    root.style.setProperty("--bg-overlay", theme.background.overlay || "none");
    root.style.setProperty("--felt-color", theme.table.feltColor);
    root.style.setProperty("--felt-shadow", theme.table.shadow);
    root.style.setProperty("--table-border-color", theme.table.borderColor);
    root.style.setProperty("--table-border-width", theme.table.borderWidth);
    root.style.setProperty("--table-decoration", theme.table.decoration);
    return () => {
      root.style.removeProperty("--bg-image");
      root.style.removeProperty("--bg-color");
      root.style.removeProperty("--bg-overlay");
      root.style.removeProperty("--felt-color");
      root.style.removeProperty("--felt-shadow");
      root.style.removeProperty("--table-border-color");
      root.style.removeProperty("--table-border-width");
      root.style.removeProperty("--table-decoration");
    };
  }, [theme]);

  return (
    <RoomThemeContext.Provider value={theme}>
      {children}
    </RoomThemeContext.Provider>
  );
}
