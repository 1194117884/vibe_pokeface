"use client";

import { useState, useRef, useEffect } from "react";
import { EmojiPicker } from "./EmojiPicker";

interface ChatMessage {
  userId: string;
  nickname: string;
  content: string;
  type: "text" | "emoji";
  timestamp: number;
}

interface ChatPanelProps {
  messages: ChatMessage[];
  onSendMessage: (content: string, type: "text" | "emoji") => void;
  disabled?: boolean;
}

export function ChatPanel({ messages, onSendMessage, disabled }: ChatPanelProps) {
  const [input, setInput] = useState("");
  const [showEmoji, setShowEmoji] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSend = () => {
    if (!input.trim() || disabled) return;
    onSendMessage(input.trim(), "text");
    setInput("");
  };

  const handleEmojiSelect = (emoji: string) => {
    onSendMessage(emoji, "emoji");
    setShowEmoji(false);
  };

  return (
    <div className="relative flex flex-col h-64 bg-surface-container-high rounded-[12px] shadow-card border border-outline-variant">
      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-3 space-y-2">
        {messages.length === 0 && (
          <p className="text-center text-on-surface-variant text-sm">No messages yet</p>
        )}
        {messages.map((msg, i) => (
          <div key={i} className="text-sm">
            <span className="font-semibold text-primary">
              {msg.nickname}:{" "}
            </span>
            <span className="text-on-surface">
              {msg.type === "emoji" ? (
                <span className="text-2xl">{msg.content}</span>
              ) : (
                msg.content
              )}
            </span>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="border-t border-outline-variant p-2 flex gap-2">
        <button
          onClick={() => setShowEmoji(!showEmoji)}
          className="px-2 py-1 text-lg hover:bg-surface-container rounded-[4px] transition-colors"
          title="Open emoji picker"
        >
          😊
        </button>
        <input
          maxLength={500}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSend()}
          placeholder="Type a message..."
          disabled={disabled}
          className="flex-1 px-3 py-1.5 text-sm bg-surface border border-outline-variant rounded-pill outline-none transition-all duration-200 text-on-surface placeholder:text-on-surface-variant focus:border-primary disabled:opacity-50"
        />
        <button
          onClick={handleSend}
          disabled={disabled || !input.trim()}
          className="px-4 py-1.5 bg-primary text-on-primary text-sm rounded-pill font-semibold disabled:opacity-40 hover:brightness-110 transition-all duration-200 active:scale-[0.95]"
        >
          Send
        </button>
      </div>

      {/* Emoji picker */}
      {showEmoji && (
        <EmojiPicker onSelect={handleEmojiSelect} onClose={() => setShowEmoji(false)} />
      )}
    </div>
  );
}
