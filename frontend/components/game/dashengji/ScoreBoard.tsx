"use client";

interface ScoreBoardProps {
  roundPoints: number;
  dealerLevel: number;
}

const levelNames = ["", "", "", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"];

export function ScoreBoard({ roundPoints, dealerLevel }: ScoreBoardProps) {
  return (
    <div className="fixed top-4 right-4 bg-black/60 backdrop-blur rounded-xl p-4 text-white text-sm z-40 min-w-[180px]">
      <div className="text-white/60 text-xs mb-2">打升级</div>
      <div className="flex justify-between mb-1">
        <span>当前打</span>
        <span className="text-amber-400 font-bold text-lg">{levelNames[dealerLevel] || dealerLevel}</span>
      </div>
      <div className="border-t border-white/10 mt-2 pt-2 flex justify-between">
        <span>本轮闲家得分</span>
        <span className="text-red-400 font-bold">{roundPoints}</span>
      </div>
    </div>
  );
}
