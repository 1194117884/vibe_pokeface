"use client";

import { useCallback, useRef } from "react";

export function useGameAudio(muted: boolean) {
  const speakingRef = useRef(false);

  const speak = useCallback(
    (text: string) => {
      if (muted || typeof window === "undefined" || !window.speechSynthesis) {
        return;
      }
      window.speechSynthesis.cancel();
      const utterance = new SpeechSynthesisUtterance(text);
      utterance.lang = "zh-CN";
      utterance.rate = 1.0;
      utterance.pitch = 1.0;
      utterance.volume = 1.0;
      utterance.onstart = () => {
        speakingRef.current = true;
      };
      utterance.onend = () => {
        speakingRef.current = false;
      };
      window.speechSynthesis.speak(utterance);
    },
    [muted],
  );

  return { speak };
}
