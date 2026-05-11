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

  // Auto-scroll to bottom when new messages arrive
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

  // Convert user_id to display-friendly short name
  const formatUserId = (userId: string): string => {
    if (userId.startsWith("ai:")) return "Bot";
    return userId.length > 6 ? userId.slice(0, 6) + "..." : userId;
  };

  return (
    <div className="relative flex flex-col h-64 bg-white rounded-lg shadow-md border border-gray-200">
      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-3 space-y-2">
        {messages.length === 0 && (
          <p className="text-center text-gray-500 text-sm">No messages yet</p>
        )}
        {messages.map((msg, i) => (
          <div key={i} className="text-sm">
            <span className="font-semibold text-green-700">
              {formatUserId(msg.userId)}:{" "}
            </span>
            <span className="text-gray-800">
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
      <div className="border-t border-gray-200 p-2 flex gap-2">
        <button
          onClick={() => setShowEmoji(!showEmoji)}
          className="px-2 py-1 text-lg hover:bg-gray-100 rounded"
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
          className="flex-1 px-3 py-1.5 text-sm border border-gray-300 rounded-full outline-none focus:border-green-500 disabled:opacity-50"
        />
        <button
          onClick={handleSend}
          disabled={disabled || !input.trim()}
          className="px-4 py-1.5 bg-green-600 text-white text-sm rounded-full disabled:opacity-40 hover:bg-green-700 transition-colors"
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
