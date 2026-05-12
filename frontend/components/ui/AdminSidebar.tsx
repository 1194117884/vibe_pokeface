"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import clsx from "clsx";
import { useState } from "react";

const navItems = [
  { href: "/admin/dashboard", label: "Dashboard", icon: "📊" },
  { href: "/admin/users", label: "Users", icon: "👥" },
  { href: "/admin/rooms", label: "Rooms", icon: "🃏" },
  { href: "/admin/ai-characters", label: "AI Characters", icon: "🧑" },
  { href: "/admin/llm-config", label: "LLM Config", icon: "🤖" },
  { href: "/admin/stats", label: "LLM Stats", icon: "📈" },
];

export function AdminSidebar() {
  const pathname = usePathname();
  const [open, setOpen] = useState(false);

  const close = () => setOpen(false);

  const sidebar = (
    <aside className="w-64 bg-house-green text-white p-6 flex flex-col min-h-full">
      <div className="mb-8">
        <div className="flex items-center gap-2 mb-1">
          <span className="text-xl">🎴</span>
          <h2 className="text-lg font-bold tracking-tight">PokeFace</h2>
        </div>
        <p className="text-xs text-white/60 ml-1">Admin Console</p>
      </div>
      <nav className="space-y-1 flex-1">
        {navItems.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            onClick={close}
            className={clsx(
              "flex items-center gap-3 px-4 py-2.5 rounded-pill text-sm tracking-tight transition-all duration-200",
              pathname === item.href
                ? "bg-green-accent text-white font-semibold shadow-md"
                : "text-white/70 hover:text-white hover:bg-white/10"
            )}
          >
            <span className="text-base">{item.icon}</span>
            {item.label}
          </Link>
        ))}
      </nav>
      <Link
        href="/lobby"
        onClick={close}
        className="flex items-center gap-2 px-4 py-2.5 rounded-pill text-sm text-white/50 hover:text-white hover:bg-white/5 transition-colors mt-auto"
      >
        ← Back to Lobby
      </Link>
    </aside>
  );

  return (
    <>
      {/* Mobile hamburger */}
      <button
        onClick={() => setOpen(!open)}
        className="fixed top-4 left-4 z-50 lg:hidden bg-house-green text-white w-10 h-10 rounded-lg flex items-center justify-center shadow-nav"
        aria-label="Toggle admin menu"
      >
        <span className="text-lg leading-none">{open ? "✕" : "☰"}</span>
      </button>

      {/* Desktop sidebar — always visible */}
      <div className="hidden lg:flex shrink-0">{sidebar}</div>

      {/* Mobile drawer overlay */}
      {open && (
        <div
          className="fixed inset-0 z-40 bg-black/40 lg:hidden"
          onClick={close}
        />
      )}

      {/* Mobile drawer */}
      <div
        className={clsx(
          "fixed top-0 left-0 z-40 h-full transition-transform duration-300 ease-in-out lg:hidden",
          open ? "translate-x-0" : "-translate-x-full"
        )}
      >
        {sidebar}
      </div>
    </>
  );
}
