"use client";

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
    <div className="flex justify-center gap-6 py-3">
      {isBidding && (
        <>
          <button
            onClick={onBidCall}
            disabled={!isMyTurn}
            className="px-10 py-3 rounded-full gold-button text-button-text font-button-text hover:brightness-110 active:scale-95 transition-all disabled:opacity-40"
          >
            {isMyTurn ? callLabel : `${callLabel} ⌛`}
          </button>
          <button
            onClick={onBidPass}
            disabled={!isMyTurn}
            className="px-10 py-3 rounded-full emerald-button text-button-text font-button-text hover:brightness-110 active:scale-95 transition-all disabled:opacity-40"
          >
            {isMyTurn ? passLabel : `${passLabel} ⌛`}
          </button>
        </>
      )}
      {phase === "revealing" && (
        <>
          <button
            onClick={onReveal}
            disabled={!isMyTurn}
            className="px-10 py-3 rounded-full gold-button text-button-text font-button-text hover:brightness-110 active:scale-95 transition-all disabled:opacity-40"
          >
            {isMyTurn ? "明牌" : "明牌 ⌛"}
          </button>
          <button
            onClick={onRevealPass}
            disabled={!isMyTurn}
            className="px-10 py-3 rounded-full emerald-button text-button-text font-button-text hover:brightness-110 active:scale-95 transition-all disabled:opacity-40"
          >
            {isMyTurn ? "不亮" : "不亮 ⌛"}
          </button>
        </>
      )}
      {phase === "doubling" && (
        <>
          <button
            onClick={onDouble}
            disabled={!isMyTurn}
            className="px-10 py-3 rounded-full gold-button text-button-text font-button-text hover:brightness-110 active:scale-95 transition-all disabled:opacity-40"
          >
            {isMyTurn ? "加倍" : "加倍 ⌛"}
          </button>
          <button
            onClick={onNoDouble}
            disabled={!isMyTurn}
            className="px-10 py-3 rounded-full emerald-button text-button-text font-button-text hover:brightness-110 active:scale-95 transition-all disabled:opacity-40"
          >
            {isMyTurn ? "不加倍" : "不加倍 ⌛"}
          </button>
        </>
      )}
      {phase === "playing" && (
        <>
          <button
            onClick={onPlay}
            disabled={!isMyTurn}
            className="px-10 py-3 rounded-full gold-button text-button-text font-button-text hover:brightness-110 active:scale-95 transition-all disabled:opacity-40"
          >
            {isMyTurn ? "出牌" : "出牌 ⌛"}
          </button>
          <button
            onClick={onPass}
            disabled={!isMyTurn}
            className="px-10 py-3 rounded-full emerald-button text-button-text font-button-text hover:brightness-110 active:scale-95 transition-all disabled:opacity-40"
          >
            {isMyTurn ? "不出" : "不出 ⌛"}
          </button>
        </>
      )}
      {timer !== undefined && (
        <span className="text-lg font-bold text-primary ml-2 self-center">{timer}s</span>
      )}
    </div>
  );
}
