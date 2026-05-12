"use client";

import { useEffect } from "react";
import { AdminSidebar } from "@/components/ui/AdminSidebar";
import { apiClient } from "@/lib/api-client";

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  useEffect(() => {
    if (!apiClient.getToken()) {
      const current = window.location.pathname + window.location.search;
      window.location.href = `/auth/login?redirect=${encodeURIComponent(current)}`;
    } else if (!apiClient.isAdmin()) {
      window.location.href = "/lobby";
    }
  }, []);

  return (
    <div className="flex min-h-screen">
      <AdminSidebar />
      <main className="flex-1 bg-cream p-4 pt-16 lg:pt-8 lg:p-8 border-l border-white/10 overflow-x-auto">
        {children}
      </main>
    </div>
  );
}
