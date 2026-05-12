"use client";

import { useEffect, useState } from "react";
import { Card } from "@/components/ui/Card";
import { adminFetch } from "@/lib/admin-fetch";

interface LLMStats {
  total_calls: number;
  total_tokens: number;
  avg_latency_ms: number;
  success_rate: number;
}

export default function LLMStatsPage() {
  const [stats, setStats] = useState<LLMStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    adminFetch("/api/admin/llm-stats")
      .then((r) => r.json())
      .then((d: LLMStats) => { setStats(d); setLoading(false); })
      .catch(() => setLoading(false));
  }, []);

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-starbucks tracking-tight">LLM Call Statistics</h1>
        <p className="text-sm text-text-black-soft mt-0.5">AI performance metrics</p>
      </div>
      {loading ? (
        <Card padding="lg">
          <p className="text-text-black-soft text-center py-4">Loading...</p>
        </Card>
      ) : !stats ? (
        <Card padding="lg">
          <p className="text-text-black-soft text-center py-4">Unable to load stats.</p>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card padding="lg">
            <div className="flex items-center gap-3 mb-3">
              <div className="w-8 h-8 rounded-full bg-green-light flex items-center justify-center text-sm">📞</div>
              <p className="text-sm font-semibold text-text-black-soft tracking-tight">Total Calls</p>
            </div>
            <p className="text-4xl font-bold text-starbucks tracking-tight">{stats.total_calls}</p>
          </Card>
          <Card padding="lg">
            <div className="flex items-center gap-3 mb-3">
              <div className="w-8 h-8 rounded-full bg-green-light flex items-center justify-center text-sm">🔤</div>
              <p className="text-sm font-semibold text-text-black-soft tracking-tight">Total Tokens</p>
            </div>
            <p className="text-4xl font-bold text-starbucks tracking-tight">{stats.total_tokens.toLocaleString()}</p>
          </Card>
          <Card padding="lg">
            <div className="flex items-center gap-3 mb-3">
              <div className="w-8 h-8 rounded-full bg-green-light flex items-center justify-center text-sm">⚡</div>
              <p className="text-sm font-semibold text-text-black-soft tracking-tight">Avg Latency</p>
            </div>
            <p className="text-4xl font-bold text-starbucks tracking-tight">{stats.avg_latency_ms}ms</p>
          </Card>
          <Card padding="lg">
            <div className="flex items-center gap-3 mb-3">
              <div className="w-8 h-8 rounded-full bg-green-light flex items-center justify-center text-sm">✅</div>
              <p className="text-sm font-semibold text-text-black-soft tracking-tight">Success Rate</p>
            </div>
            <p className="text-4xl font-bold text-starbucks tracking-tight">{stats.success_rate}%</p>
          </Card>
        </div>
      )}
    </div>
  );
}
