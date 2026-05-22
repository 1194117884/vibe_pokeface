"use client";

interface DashengjiActionBarProps {
  phase: string;
  isMyTurn: boolean;
  onAction: (action: string, cards?: number[]) => void;
}

export function DashengjiActionBar({ phase, isMyTurn, onAction }: DashengjiActionBarProps) {
  const disabled = !isMyTurn;

  switch (phase) {
    case "set_trump":
      return (
        <div className="fixed bottom-24 left-1/2 -translate-x-1/2 flex gap-4 z-30">
          <button className="gold-button px-8 py-3 text-lg" disabled={disabled} onClick={() => onAction("set_trump")}>
            定主
          </button>
          <button className="px-8 py-3 text-lg rounded-xl bg-white/10 text-white/70 hover:bg-white/20" disabled={disabled} onClick={() => onAction("pass_trump")}>
            不定
          </button>
        </div>
      );
    case "counter_trump":
      return (
        <div className="fixed bottom-24 left-1/2 -translate-x-1/2 flex gap-4 z-30">
          <button className="gold-button px-8 py-3 text-lg" disabled={disabled} onClick={() => onAction("counter_trump")}>
            反主
          </button>
          <button className="px-8 py-3 text-lg rounded-xl bg-white/10 text-white/70 hover:bg-white/20" disabled={disabled} onClick={() => onAction("pass_counter")}>
            不反
          </button>
        </div>
      );
    case "take_bottom":
      return (
        <div className="fixed bottom-24 left-1/2 -translate-x-1/2 flex gap-4 z-30">
          <button className="gold-button px-8 py-3 text-lg" disabled={disabled} onClick={() => onAction("take_bottom")}>
            起底
          </button>
        </div>
      );
    case "discard_bottom":
      return (
        <div className="fixed bottom-24 left-1/2 -translate-x-1/2 flex gap-4 z-30">
          <button className="gold-button px-8 py-3 text-lg" disabled={disabled} onClick={() => onAction("discard_bottom")}>
            扣底
          </button>
        </div>
      );
    case "playing":
      return (
        <div className="fixed bottom-24 left-1/2 -translate-x-1/2 flex gap-4 z-30">
          <button className="emerald-button px-8 py-3 text-lg" disabled={disabled} onClick={() => onAction("play")}>
            出牌
          </button>
          <button className="px-8 py-3 text-lg rounded-xl bg-white/10 text-white/70 hover:bg-white/20" disabled={disabled} onClick={() => onAction("pass")}>
            不出
          </button>
        </div>
      );
    default:
      return null;
  }
}
