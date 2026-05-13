"use client";

import { useEffect, useState } from "react";
import { useRoomTheme } from "@/themes";

interface NPC {
  id: number;
  emoji: string;
  direction: "left" | "right";
  y: number;
  speed: number;
  delay: number;
}

export function NPCWalker() {
  const theme = useRoomTheme();
  const [npcs, setNpcs] = useState<NPC[]>([]);

  useEffect(() => {
    if (!theme.ambient.enabled || !theme.ambient.npcSprites?.length) {
      setNpcs([]);
      return;
    }

    const count = theme.ambient.npcCount || 2;
    const sprites = theme.ambient.npcSprites;

    const generated: NPC[] = Array.from({ length: count }, (_, i) => ({
      id: i,
      emoji: sprites[i % sprites.length],
      direction: Math.random() > 0.5 ? "left" : "right",
      y: 20 + Math.random() * 60,
      speed: 10 + Math.random() * 8,
      delay: i * (4 + Math.random() * 4),
    }));

    setNpcs(generated);
  }, [theme]);

  if (npcs.length === 0) return null;

  return (
    <div className="fixed inset-0 pointer-events-none overflow-hidden z-10">
      {npcs.map((npc) => (
        <div
          key={npc.id}
          className="absolute text-2xl opacity-20"
          style={{
            top: `${npc.y}%`,
            animation: `npc-walk-${npc.direction} ${npc.speed}s ${npc.delay}s linear infinite`,
          }}
        >
          {npc.emoji}
        </div>
      ))}
    </div>
  );
}
