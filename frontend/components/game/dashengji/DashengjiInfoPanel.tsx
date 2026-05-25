"use client";

import { useState } from "react";
import clsx from "clsx";

const LEVEL_NAMES = ["", "", "", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"];
const SUIT_SYMBOLS = ["♠", "♥", "♣", "♦"];

interface DashengjiInfoPanelProps {
  trumpSuit: number;
  roundPoints: number;
  dealerLevel: number;
}

export function DashengjiInfoPanel({ trumpSuit, roundPoints, dealerLevel }: DashengjiInfoPanelProps) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className="fixed right-0 top-1/2 -translate-y-1/2 z-40 pointer-events-none">
      <div
        className={clsx(
          "flex items-stretch transition-transform duration-300 ease-out pointer-events-auto",
          expanded ? "translate-x-0" : "translate-x-[calc(100%-2rem)]",
        )}
      >
        {/* Toggle button */}
        <button
          onClick={() => setExpanded(!expanded)}
          className="flex items-center justify-center w-8 bg-black/60 backdrop-blur rounded-l-lg text-white/70 hover:text-white hover:bg-black/80 transition-colors text-sm flex-shrink-0 pointer-events-auto"
          title={expanded ? "收起" : "展开"}
        >
          {expanded ? ">" : "<"}
        </button>

        {/* Panel content */}
        <div className="bg-black/60 backdrop-blur rounded-r-xl p-4 text-white text-sm min-w-[170px] pointer-events-none">
          {trumpSuit >= 0 && (
            <div className="pb-3 mb-3 border-b border-white/10">
              <div className="text-lg">{SUIT_SYMBOLS[trumpSuit]} 主花色</div>
            </div>
          )}
          <div className="text-white/60 text-xs mb-2">打升级</div>
          <div className="flex justify-between mb-1">
            <span>当前打</span>
            <span className="text-amber-400 font-bold text-lg">{LEVEL_NAMES[dealerLevel] || dealerLevel}</span>
          </div>
          <div className="border-t border-white/10 mt-2 pt-2 flex justify-between">
            <span>本轮闲家得分</span>
            <span className="text-red-400 font-bold">{roundPoints}</span>
          </div>
        </div>
      </div>
    </div>
  );
}
