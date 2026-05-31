"use client";

interface DashengjiActionBarProps {
  phase: string;
  isMyTurn: boolean;
  isDealerTeam: boolean;
  selectedCards?: number[];
  mySeat?: number | null;
  takeBottomSeat?: number;
  onAction: (action: string, cards?: number[]) => void;
}

export function DashengjiActionBar({ phase, isMyTurn, isDealerTeam, selectedCards, mySeat, takeBottomSeat, onAction }: DashengjiActionBarProps) {
  const disabled = !isMyTurn;

  switch (phase) {
    case "set_trump":
      // Only dealer team can set trump
      if (!isDealerTeam) return null;
      return (
        <div className="grid grid-cols-2 gap-2 px-3 pt-2">
          <button className="min-h-14 rounded-full gold-button px-4 py-3 text-lg font-black" onClick={() => onAction("set_trump", selectedCards)}>
            定主
          </button>
          <button className="min-h-14 rounded-full bg-white/10 px-4 py-3 text-lg font-black text-white/80 hover:bg-white/20" onClick={() => onAction("pass_trump")}>
            不定
          </button>
        </div>
      );
    case "counter_trump":
      // Only non-dealer team (闲家) can counter trump
      if (isDealerTeam) return null;
      return (
        <div className="grid grid-cols-2 gap-2 px-3 pt-2">
          <button className="min-h-14 rounded-full gold-button px-4 py-3 text-lg font-black" onClick={() => onAction("counter_trump", selectedCards)}>
            反主
          </button>
          <button className="min-h-14 rounded-full bg-white/10 px-4 py-3 text-lg font-black text-white/80 hover:bg-white/20" onClick={() => onAction("pass_counter")}>
            不反
          </button>
        </div>
      );
    case "take_bottom":
      // Only dealer team can take bottom cards
      if (!isDealerTeam) return null;
      return (
        <div className="grid grid-cols-2 gap-2 px-3 pt-2">
          <button className="min-h-14 rounded-full gold-button px-4 py-3 text-lg font-black disabled:opacity-40" disabled={disabled} onClick={() => onAction("take_bottom")}>
            起底
          </button>
          <button className="min-h-14 rounded-full bg-white/10 px-4 py-3 text-lg font-black text-white/80 hover:bg-white/20 disabled:opacity-40" disabled={disabled} onClick={() => onAction("pass_take_bottom")}>
            队友起底
          </button>
        </div>
      );
    case "discard_bottom":
      // Only the player who took bottom cards can discard
      if (mySeat == null || mySeat !== takeBottomSeat) return null;
      return (
        <div className="px-3 pt-2">
          <button className="min-h-14 w-full rounded-full gold-button px-4 py-3 text-lg font-black disabled:opacity-40" disabled={disabled || !selectedCards?.length} onClick={() => onAction("discard_bottom", selectedCards)}>
            扣底
          </button>
        </div>
      );
    case "playing":
      // HandCards component renders 出牌/不出 during playing phase
      return null;
    default:
      return null;
  }
}
