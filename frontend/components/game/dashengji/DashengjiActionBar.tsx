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
        <div className="flex justify-center gap-4 pt-2">
          <button className="gold-button px-8 py-3 text-lg" disabled={disabled} onClick={() => onAction("set_trump", selectedCards)}>
            定主
          </button>
          <button className="px-8 py-3 text-lg rounded-xl bg-white/10 text-white/70 hover:bg-white/20" disabled={disabled} onClick={() => onAction("pass_trump")}>
            不定
          </button>
        </div>
      );
    case "counter_trump":
      // Only non-dealer team (闲家) can counter trump
      if (isDealerTeam) return null;
      return (
        <div className="flex justify-center gap-4 pt-2">
          <button className="gold-button px-8 py-3 text-lg" disabled={disabled} onClick={() => onAction("counter_trump", selectedCards)}>
            反主
          </button>
          <button className="px-8 py-3 text-lg rounded-xl bg-white/10 text-white/70 hover:bg-white/20" disabled={disabled} onClick={() => onAction("pass_counter")}>
            不反
          </button>
        </div>
      );
    case "take_bottom":
      // Only dealer team can take bottom cards
      if (!isDealerTeam) return null;
      return (
        <div className="flex justify-center gap-4 pt-2">
          <button className="gold-button px-8 py-3 text-lg" disabled={disabled} onClick={() => onAction("take_bottom")}>
            起底
          </button>
          <button className="px-8 py-3 text-lg rounded-xl bg-white/10 text-white/70 hover:bg-white/20" disabled={disabled} onClick={() => onAction("pass_take_bottom")}>
            队友起底
          </button>
        </div>
      );
    case "discard_bottom":
      // Only the player who took bottom cards can discard
      if (mySeat == null || mySeat !== takeBottomSeat) return null;
      return (
        <div className="flex justify-center gap-4 pt-2">
          <button className="gold-button px-8 py-3 text-lg" disabled={disabled || !selectedCards?.length} onClick={() => onAction("discard_bottom", selectedCards)}>
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
