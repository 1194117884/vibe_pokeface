"use client";

import { useEffect, useState } from "react";
import { Card } from "@/components/ui/Card";
import { adminFetch } from "@/lib/admin-fetch";

interface DashboardData {
  online_players: number;
  active_rooms: number;
  total_users: number;
}

export default function DashboardPage() {
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    adminFetch("/api/admin/dashboard")
      .then((r) => r.json())
      .then((d: DashboardData) => { setData(d); setLoading(false); })
      .catch(() => setLoading(false));
  }, []);

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-starbucks tracking-tight">Dashboard</h1>
        <p className="text-sm text-text-black-soft mt-0.5">Platform overview at a glance</p>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card padding="lg">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-8 h-8 rounded-full bg-green-light flex items-center justify-center text-sm">👥</div>
            <p className="text-sm font-semibold text-text-black-soft tracking-tight">Online Players</p>
          </div>
          <p className="text-4xl font-bold text-starbucks tracking-tight">{loading ? "--" : data?.online_players ?? 0}</p>
        </Card>
        <Card padding="lg">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-8 h-8 rounded-full bg-green-light flex items-center justify-center text-sm">🃏</div>
            <p className="text-sm font-semibold text-text-black-soft tracking-tight">Active Rooms</p>
          </div>
          <p className="text-4xl font-bold text-starbucks tracking-tight">{loading ? "--" : data?.active_rooms ?? 0}</p>
        </Card>
        <Card padding="lg">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-8 h-8 rounded-full bg-green-light flex items-center justify-center text-sm">📋</div>
            <p className="text-sm font-semibold text-text-black-soft tracking-tight">Total Users</p>
          </div>
          <p className="text-4xl font-bold text-starbucks tracking-tight">{loading ? "--" : data?.total_users ?? 0}</p>
        </Card>
      </div>
    </div>
  );
}
