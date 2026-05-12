"use client";

const EMOJIS = [
  "😊", "😂", "🤣", "❤️", "🎉", "👍", "🙌", "👏",
  "🔥", "💯", "😅", "🤔", "😤", "😈", "💪", "🃏",
  "🎯", "🎲", "💰", "🌟", "💥", "🍀", "👑", "🤡",
];

interface EmojiPickerProps {
  onSelect: (emoji: string) => void;
  onClose: () => void;
}

export function EmojiPicker({ onSelect, onClose }: EmojiPickerProps) {
  return (
    <div className="absolute bottom-16 left-4 bg-white border border-ceramic rounded-[12px] shadow-card p-2 z-10">
      <div className="grid grid-cols-6 gap-1">
        {EMOJIS.map((emoji) => (
          <button
            key={emoji}
            onClick={() => onSelect(emoji)}
            className="w-8 h-8 flex items-center justify-center hover:bg-cream rounded-[4px] transition-colors text-lg"
          >
            {emoji}
          </button>
        ))}
      </div>
      <button
        onClick={onClose}
        className="w-full mt-1 text-xs text-text-black-soft hover:text-text-black transition-colors"
      >
        Close
      </button>
    </div>
  );
}
