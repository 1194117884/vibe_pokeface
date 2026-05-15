"use client";

import { Button } from "@/components/ui/Button";

interface ActionBarProps {
  phase: "calling" | "snatching" | "revealing" | "doubling" | "playing" | "ended";
  isMyTurn: boolean;
  onBidCall?: () => void;
  onBidPass?: () => void;
  onReveal?: () => void;
  onRevealPass?: () => void;
  onDouble?: () => void;
  onNoDouble?: () => void;
  onPlay?: () => void;
  onPass?: () => void;
  timer?: number;
}

export function ActionBar({
  phase,
  isMyTurn,
  onBidCall,
  onBidPass,
  onReveal,
  onRevealPass,
  onDouble,
  onNoDouble,
  onPlay,
  onPass,
  timer,
}: ActionBarProps) {
  const isBidding = phase === "calling" || phase === "snatching";
  const callLabel = phase === "calling" ? "叫地主" : "抢地主";
  const passLabel = phase === "calling" ? "不叫" : "不抢";

  return (
    <div className="flex justify-center gap-3 py-3">
      {isBidding && (
        <>
          <Button variant="primary" onClick={onBidCall} disabled={!isMyTurn}>
            {isMyTurn ? callLabel : `${callLabel} ⌛`}
          </Button>
          <Button variant="outlined" onClick={onBidPass} disabled={!isMyTurn}>
            {isMyTurn ? passLabel : `${passLabel} ⌛`}
          </Button>
        </>
      )}
      {phase === "revealing" && (
        <>
          <Button variant="primary" onClick={onReveal} disabled={!isMyTurn}>
            {isMyTurn ? "明牌" : "明牌 ⌛"}
          </Button>
          <Button variant="outlined" onClick={onRevealPass} disabled={!isMyTurn}>
            {isMyTurn ? "不亮" : "不亮 ⌛"}
          </Button>
        </>
      )}
      {phase === "doubling" && (
        <>
          <Button variant="primary" onClick={onDouble} disabled={!isMyTurn}>
            {isMyTurn ? "加倍" : "加倍 ⌛"}
          </Button>
          <Button variant="outlined" onClick={onNoDouble} disabled={!isMyTurn}>
            {isMyTurn ? "不加倍" : "不加倍 ⌛"}
          </Button>
        </>
      )}
      {phase === "playing" && (
        <>
          <Button variant="primary" onClick={onPlay} disabled={!isMyTurn}>
            {isMyTurn ? "出牌" : "出牌 ⌛"}
          </Button>
          <Button variant="outlined" onClick={onPass} disabled={!isMyTurn}>
            {isMyTurn ? "不出" : "不出 ⌛"}
          </Button>
        </>
      )}
      {timer !== undefined && (
        <span className="text-lg font-bold text-green-accent ml-2 self-center">{timer}s</span>
      )}
    </div>
  );
}
