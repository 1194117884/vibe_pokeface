"use client";

import { Button } from "@/components/ui/Button";

interface ReadyBarProps {
  amIOwner: boolean;
  isReady: boolean;
  allReady: boolean;
  playerCount: number;
  maxPlayers: number;
  canStart: boolean;
  onReady: () => void;
  onStartGame: () => void;
  onAddBot: () => void;
}

export function ReadyBar({
  amIOwner,
  isReady,
  allReady,
  playerCount,
  maxPlayers,
  canStart,
  onReady,
  onStartGame,
  onAddBot,
}: ReadyBarProps) {
  const roomFull = playerCount >= maxPlayers;

  return (
    <div className="flex items-center justify-center gap-3 py-4">
      {amIOwner ? (
        <>
          {!roomFull && (
            <Button variant="outlined" onClick={onAddBot}>
              + 添加AI ({playerCount}/{maxPlayers})
            </Button>
          )}
          <Button
            variant="primary"
            onClick={onStartGame}
            disabled={!canStart}
          >
            {!roomFull
              ? "等待更多玩家..."
              : !allReady
                ? "等待准备..."
                : "开始游戏"}
          </Button>
        </>
      ) : (
        <Button
          variant={isReady ? "outlined" : "primary"}
          onClick={onReady}
        >
          {isReady ? "取消准备" : "准备"}
        </Button>
      )}
    </div>
  );
}
