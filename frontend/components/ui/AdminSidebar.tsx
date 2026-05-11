"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import clsx from "clsx";

const navItems = [
  { href: "/admin/dashboard", label: "Dashboard", icon: "📊" },
  { href: "/admin/users", label: "Users", icon: "👥" },
  { href: "/admin/rooms", label: "Rooms", icon: "🃏" },
  { href: "/admin/ai-characters", label: "AI Characters", icon: "🧑" },
  { href: "/admin/llm-config", label: "LLM Config", icon: "🤖" },
];

export function AdminSidebar() {
  const pathname = usePathname();

  return (
    <aside className="w-64 bg-house-green min-h-screen text-white p-6">
      <div className="mb-8">
        <h2 className="text-xl font-semibold tracking-tight">Admin</h2>
        <p className="text-sm text-white/70 mt-1">Vibe Pokeface</p>
      </div>
      <nav className="space-y-1">
        {navItems.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className={clsx(
              "flex items-center gap-3 px-4 py-3 rounded-lg text-sm transition-colors",
              pathname === item.href
                ? "bg-white/10 text-white font-medium"
                : "text-white/70 hover:text-white hover:bg-white/5"
            )}
          >
            <span>{item.icon}</span>
            {item.label}
          </Link>
        ))}
      </nav>
    </aside>
  );
}
