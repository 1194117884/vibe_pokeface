"use client";

import { useState, useRef, useEffect } from "react";
import { EmojiPicker } from "./EmojiPicker";

interface ChatMessage {
  userId: string;
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

  const formatUserId = (userId: string): string => {
    if (userId.startsWith("ai:")) return "Bot";
    return userId.length > 6 ? userId.slice(0, 6) + "..." : userId;
  };

  return (
    <div className="relative flex flex-col h-64 bg-white rounded-[12px] shadow-card border border-ceramic">
      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-3 space-y-2">
        {messages.length === 0 && (
          <p className="text-center text-text-black-soft text-sm">No messages yet</p>
        )}
        {messages.map((msg, i) => (
          <div key={i} className="text-sm">
            <span className="font-semibold text-green-accent">
              {formatUserId(msg.userId)}:{" "}
            </span>
            <span className="text-text-black">
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
      <div className="border-t border-ceramic p-2 flex gap-2">
        <button
          onClick={() => setShowEmoji(!showEmoji)}
          className="px-2 py-1 text-lg hover:bg-cream rounded-[4px] transition-colors"
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
          className="flex-1 px-3 py-1.5 text-sm border border-gray-300 rounded-pill outline-none transition-all duration-200 focus:border-green-accent disabled:opacity-50"
        />
        <button
          onClick={handleSend}
          disabled={disabled || !input.trim()}
          className="px-4 py-1.5 bg-green-accent text-white text-sm rounded-pill font-semibold disabled:opacity-40 hover:bg-green-uplift transition-all duration-200 active:scale-[0.95]"
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
