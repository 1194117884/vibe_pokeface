"use client";

import { useState } from "react";

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
      className={`w-10 h-10 rounded-full flex items-center justify-center transition-all text-lg border-2 ${
        enabled
          ? "bg-green-600 text-white border-green-600"
          : "bg-white text-gray-500 border-gray-300"
      } ${disabled ? "opacity-40 cursor-not-allowed" : ""}`}
      title={enabled ? "Mute" : "Unmute"}
    >
      {enabled ? "\u{1F3A4}" : "\u{1F507}"}
    </button>
  );
}
