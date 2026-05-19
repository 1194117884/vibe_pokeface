"use client";

import { useEffect, useState } from "react";
import clsx from "clsx";
import { listAICharacters, type AICharacterInfo } from "@/lib/api-rooms";

interface AICharacterPickerProps {
  open: boolean;
  onClose: () => void;
  onSelect: (characterId: number) => void;
}

const playStyleLabels: Record<string, string> = {
  aggressive: "激进",
  conservative: "保守",
  balanced: "稳健",
  unpredictable: "随机",
};

const playStyleColors: Record<string, string> = {
  aggressive: "bg-error-container text-error",
  conservative: "bg-primary-container text-primary",
  balanced: "bg-surface-container-high text-on-surface-variant",
  unpredictable: "bg-tertiary-container text-tertiary",
};

export function AICharacterPicker({ open, onClose, onSelect }: AICharacterPickerProps) {
  const [characters, setCharacters] = useState<AICharacterInfo[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    listAICharacters().then((chars) => {
      setCharacters(chars);
      setLoading(false);
    });
  }, []);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="bg-surface-container-high rounded-2xl shadow-frap p-6 max-w-md w-full mx-4 max-h-[80vh] flex flex-col border border-outline-variant">
        <h2 className="text-xl font-bold text-on-surface mb-4 text-center">
          选择AI角色
        </h2>

        {loading ? (
          <p className="text-on-surface-variant text-center py-8">加载中...</p>
        ) : characters.length === 0 ? (
          <p className="text-on-surface-variant text-center py-8">
            暂无可用AI角色，请先在管理后台创建
          </p>
        ) : (
          <div className="space-y-2 overflow-y-auto flex-1">
            {characters.map((char) => (
              <button
                key={char.id}
                onClick={() => onSelect(char.id)}
                className="w-full text-left p-4 rounded-xl border border-outline-variant hover:border-primary/50 hover:bg-surface-container transition-colors"
              >
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-primary-container text-primary flex items-center justify-center font-bold text-sm shrink-0">
                    {char.name.charAt(0)}
                  </div>
                  <div className="min-w-0">
                    <div className="font-semibold text-on-surface truncate">
                      {char.name}
                    </div>
                    <div className="flex items-center gap-2 mt-0.5">
                      <span className={clsx(
                        "text-xs px-2 py-0.5 rounded-full",
                        playStyleColors[char.play_style] || "bg-surface-container text-on-surface-variant"
                      )}>
                        {playStyleLabels[char.play_style] || char.play_style}
                      </span>
                      {char.personality && (
                        <span className="text-xs text-on-surface-variant truncate">
                          {char.personality}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              </button>
            ))}
          </div>
        )}

        <div className="mt-4 pt-3 border-t border-outline-variant">
          <button
            onClick={onClose}
            className="w-full px-6 py-2.5 rounded-full emerald-button text-button-text font-button-text hover:brightness-110 active:scale-95 transition-all"
          >
            取消
          </button>
        </div>
      </div>
    </div>
  );
}

