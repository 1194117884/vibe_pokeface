"use client";

import { useState } from "react";
import clsx from "clsx";

interface GamePhasePromptProps {
  title: string;
  detail?: string;
  tone?: "action" | "waiting" | "error";
  position?: "top" | "tableCenter";
  compact?: boolean;
}

export function GamePhasePrompt({ title, detail, tone = "waiting", position = "top", compact = false }: GamePhasePromptProps) {
  const promptKey = `${position}:${tone}:${title}:${detail ?? ""}`;
  const [dismissedKey, setDismissedKey] = useState<string | null>(null);

  if (dismissedKey === promptKey) return null;

  return (
    <div
      className={clsx(
        "pointer-events-auto fixed left-3 right-3 z-40 mx-auto max-w-[430px]",
        position === "top" && "top-[calc(var(--safe-area-top)+3rem)]",
        position === "tableCenter" && "top-[39vh]",
        compact ? "max-w-[min(390px,34vw)] rounded-full px-10 py-2" : "rounded-2xl px-12 py-3",
        "border text-center shadow-lg backdrop-blur-md",
        tone === "action" && "border-amber-300/60 bg-amber-400/95 text-[#221b00]",
        tone === "waiting" && "border-white/15 bg-black/70 text-white",
        tone === "error" && "border-red-300/60 bg-red-600/95 text-white",
      )}
    >
      <button
        type="button"
        onClick={() => setDismissedKey(promptKey)}
        className={clsx(
          "absolute flex items-center justify-center rounded-full bg-black/10 font-black leading-none text-current opacity-65 active:scale-95",
          compact ? "right-1.5 top-1.5 h-7 w-7 text-lg" : "right-2 top-2 h-8 w-8 text-xl",
        )}
        aria-label="关闭提醒"
      >
        ×
      </button>
      <p className={clsx("truncate font-black", compact ? "text-base leading-5" : "text-lg leading-6")}>{title}</p>
      {!compact && detail && <p className="mt-1 text-base font-bold leading-5 opacity-90">{detail}</p>}
    </div>
  );
}
