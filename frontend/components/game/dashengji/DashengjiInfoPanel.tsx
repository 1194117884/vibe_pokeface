import clsx from "clsx";

const LEVEL_NAMES = ["", "", "", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"];
const SUIT_SYMBOLS = ["♠", "♥", "♣", "♦"];
const RED_SUITS = new Set([1, 3]);

interface DashengjiInfoPanelProps {
  trumpSuit: number;
  roundPoints: number;
  teamLevels: [number, number];
  dealerSeats: [number, number];
  teamNames: [string, string];
}

export function DashengjiInfoPanel({ trumpSuit, roundPoints, teamLevels, dealerSeats }: DashengjiInfoPanelProps) {
  const dealerTeam = dealerSeats[0] >= 0 ? dealerSeats[0] % 2 : -1;
  return (
    <>
      <div className="fixed left-3 top-[calc(var(--safe-area-top)+0.75rem)] z-40 flex items-center gap-2 rounded-xl border border-white/10 bg-black/70 px-3 py-2 text-white shadow-lg backdrop-blur-md">
        <span className="text-sm font-bold leading-none text-white/70">抢分</span>
        <span className="text-xl font-black leading-none text-red-300">{roundPoints}</span>
      </div>

      <div className="fixed right-3 top-[calc(var(--safe-area-top)+0.75rem)] z-40 flex items-center rounded-xl border border-white/10 bg-black/70 text-white shadow-lg backdrop-blur-md">
        <div className="flex items-center gap-2 px-3 py-2">
          <span className="text-sm font-bold leading-none text-white/70">等级</span>
          <span className={clsx("text-lg font-black leading-none", dealerTeam === 0 && "text-amber-300")}>
            {LEVEL_NAMES[teamLevels[0]] || teamLevels[0]}
          </span>
          <span className="text-white/35">/</span>
          <span className={clsx("text-lg font-black leading-none", dealerTeam === 1 && "text-amber-300")}>
            {LEVEL_NAMES[teamLevels[1]] || teamLevels[1]}
          </span>
          {trumpSuit >= 0 && (
            <span className={clsx("rounded bg-white px-1.5 text-base font-black leading-6", RED_SUITS.has(trumpSuit) ? "text-red-600" : "text-black")}>
              {SUIT_SYMBOLS[trumpSuit]}
            </span>
          )}
        </div>
      </div>
    </>
  );
}
