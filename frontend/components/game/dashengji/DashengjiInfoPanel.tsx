import clsx from "clsx";

const LEVEL_NAMES = ["", "", "", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"];
const SUIT_SYMBOLS = ["♠", "♥", "♣", "♦"];

interface DashengjiInfoPanelProps {
  trumpSuit: number;
  roundPoints: number;
  teamLevels: [number, number];
  dealerSeats: [number, number];
  teamNames: [string, string];
}

export function DashengjiInfoPanel({ trumpSuit, roundPoints, teamLevels, dealerSeats, teamNames }: DashengjiInfoPanelProps) {
  const dealerTeam = dealerSeats[0] >= 0 ? dealerSeats[0] % 2 : -1;
  return (
    <>
      <div className="fixed top-3 left-3 z-40 flex items-center gap-2 rounded-lg border border-white/10 bg-black/55 px-2.5 py-1.5 text-white shadow-lg backdrop-blur-md">
        <span className="text-[10px] leading-none text-white/55">抢分</span>
        <span className="text-lg font-bold leading-none text-red-400">{roundPoints}</span>
      </div>

      <div className="fixed right-3 top-3 z-40 flex items-center rounded-lg border border-white/10 bg-black/55 text-xs text-white shadow-lg backdrop-blur-md">
        <div className="flex items-center gap-1.5 px-2.5 py-1.5 lg:hidden">
          <span className="text-[10px] leading-none text-white/55">等级</span>
          <span className={clsx("text-sm font-bold leading-none", dealerTeam === 0 && "text-amber-400")}>
            {LEVEL_NAMES[teamLevels[0]] || teamLevels[0]}
          </span>
          <span className="text-white/35">/</span>
          <span className={clsx("text-sm font-bold leading-none", dealerTeam === 1 && "text-amber-400")}>
            {LEVEL_NAMES[teamLevels[1]] || teamLevels[1]}
          </span>
          {trumpSuit >= 0 && <span className="text-[10px] leading-none text-white/65">{SUIT_SYMBOLS[trumpSuit]}</span>}
        </div>

        <div className="hidden w-[28rem] items-center gap-2 p-1.5 lg:flex">
        <div className="flex shrink-0 flex-col leading-none text-white/55">
          <span className="text-[10px]">等级</span>
          {trumpSuit >= 0 && <span className="mt-1 text-[10px]">{SUIT_SYMBOLS[trumpSuit]} 主</span>}
        </div>
        <div className="grid min-w-0 flex-1 grid-cols-2 gap-1">
          {[0, 1].map((team) => (
            <div
              key={team}
              className={clsx(
                "grid min-w-0 grid-cols-[1fr_auto] items-center gap-1 rounded-md px-2 py-1",
                dealerTeam === team && "bg-amber-400/15 ring-1 ring-amber-400/45",
              )}
            >
              <span className="truncate text-[11px] leading-4 text-white/85" title={teamNames[team]}>
                {teamNames[team]}
              </span>
              <span className={clsx("min-w-4 text-right text-sm font-bold leading-4", dealerTeam === team && "text-amber-400")}>
                {LEVEL_NAMES[teamLevels[team]] || teamLevels[team]}
              </span>
            </div>
          ))}
        </div>
        </div>
      </div>
    </>
  );
}
