"use client";

import { useState, FormEvent } from "react";

interface PasswordPromptProps {
  open: boolean;
  error?: string;
  loading?: boolean;
  onSubmit: (password: string) => void;
  onCancel: () => void;
}

export function PasswordPrompt({ open, error, loading, onSubmit, onCancel }: PasswordPromptProps) {
  const [password, setPassword] = useState("");

  if (!open) return null;

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (!loading && password) {
      onSubmit(password);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div
        role="dialog"
        className="w-[90vw] max-w-[380px] rounded-2xl bg-white px-8 pb-8 pt-6 shadow-lg"
      >
        <h2 className="mb-6 text-center text-lg font-bold text-text-black-strong">
          此房间需要密码
        </h2>
        <form onSubmit={handleSubmit}>
          <div className="mb-1">
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="请输入密码"
              className="w-full rounded-[8px] border border-gray-300 px-4 py-3 text-base text-text-black outline-none transition-colors focus:border-green-accent"
            />
          </div>
          {error && (
            <p className="mb-4 text-base font-bold text-red-error">{error}</p>
          )}
          <div className="mt-6 flex flex-col gap-3">
            <button
              type="submit"
              disabled={!password || loading}
              className="min-h-[52px] w-full rounded-pill bg-green-accent px-5 py-3 text-base font-bold tracking-tight text-white transition-all duration-200 active:scale-[0.95] disabled:cursor-not-allowed disabled:opacity-50"
            >
              {loading ? "验证中..." : "确认"}
            </button>
            <button
              type="button"
              onClick={onCancel}
              className="min-h-[52px] w-full rounded-pill bg-transparent px-5 py-3 text-base font-bold tracking-tight text-green-accent transition-all duration-200 active:scale-[0.95]"
            >
              取消
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
