"use client";

import { useState } from "react";
import clsx from "clsx";

interface VoiceButtonProps {
  onToggle?: (enabled: boolean) => void;
  disabled?: boolean;
}

export function VoiceButton({ onToggle, disabled }: VoiceButtonProps) {
  const [enabled, setEnabled] = useState(false);

  const handleToggle = () => {
    if (disabled) return;
    const next = !enabled;
    setEnabled(next);
    onToggle?.(next);
  };

  return (
    <button
      onClick={handleToggle}
      disabled={disabled}
      className={clsx(
        "w-10 h-10 rounded-full flex items-center justify-center transition-all duration-200 text-lg border-2 active:scale-[0.95]",
        enabled
          ? "bg-green-accent text-white border-green-accent shadow-md"
          : "bg-white text-text-black-soft border-ceramic hover:border-green-accent/50",
        disabled && "opacity-40 cursor-not-allowed"
      )}
      title={enabled ? "Mute" : "Unmute"}
    >
      {enabled ? "\u{1F3A4}" : "\u{1F507}"}
    </button>
  );
}
