"use client";

import { useEffect, useState } from "react";
import { listAICharacters, type AICharacterInfo } from "@/lib/api-rooms";
import { Button } from "@/components/ui/Button";

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
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-white rounded-2xl shadow-frap p-6 max-w-md w-full mx-4 max-h-[80vh] flex flex-col">
        <h2 className="text-xl font-bold text-text-black-strong mb-4 text-center">
          选择AI角色
        </h2>

        {loading ? (
          <p className="text-text-black-soft text-center py-8">加载中...</p>
        ) : characters.length === 0 ? (
          <p className="text-text-black-soft text-center py-8">
            暂无可用AI角色，请先在管理后台创建
          </p>
        ) : (
          <div className="space-y-2 overflow-y-auto flex-1">
            {characters.map((char) => (
              <button
                key={char.id}
                onClick={() => onSelect(char.id)}
                className="w-full text-left p-4 rounded-xl border border-ceramic hover:border-green-accent/50 hover:bg-cream transition-colors"
              >
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-green-accent/10 text-green-accent flex items-center justify-center font-bold text-sm shrink-0">
                    {char.name.charAt(0)}
                  </div>
                  <div className="min-w-0">
                    <div className="font-semibold text-text-black truncate">
                      {char.name}
                    </div>
                    <div className="flex items-center gap-2 mt-0.5">
                      <span className="text-xs px-2 py-0.5 rounded-full bg-cream text-text-black-soft">
                        {playStyleLabels[char.play_style] || char.play_style}
                      </span>
                      {char.personality && (
                        <span className="text-xs text-text-black-soft truncate">
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

        <div className="mt-4 pt-3 border-t border-ceramic">
          <Button variant="outlined" fullWidth onClick={onClose}>
            取消
          </Button>
        </div>
      </div>
    </div>
  );
}
