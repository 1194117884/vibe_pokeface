"use client";

import { useEffect, useMemo, useState } from "react";
import { Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { adminFetch } from "@/lib/admin-fetch";

interface AIRun {
  id: number;
  provider: string;
  model: string;
  prompt_tokens: number;
  completion_tokens: number;
  duration_ms: number;
  success: boolean;
  error_message?: string;
  call_type: string;
  room_id?: string;
  user_id?: string;
  seat: number;
  phase?: string;
  turn_number: number;
  tool_count: number;
  action_count: number;
  created_at: string;
}

interface ToolExecution {
  id: number;
  tool_name: string;
  tool_type: string;
  args_json?: string;
  result_json?: string;
  created_at: string;
}

interface GameAction {
  id: number;
  action_type: string;
  cards?: string;
  full_state?: string;
  tool_execution_id?: number;
  created_at: string;
}

interface AIRunDetail extends AIRun {
  request_json?: string;
  response_json?: string;
  tools?: ToolExecution[];
  game_actions?: GameAction[];
}

function pretty(raw?: string) {
  if (!raw) return "";
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

function CodeBlock({ value }: { value?: string }) {
  return (
    <pre className="max-h-[420px] overflow-auto rounded-[4px] bg-black/90 p-4 text-xs leading-relaxed text-white">
      {pretty(value) || "Empty"}
    </pre>
  );
}

export default function AIRunsPage() {
  const [runs, setRuns] = useState<AIRun[]>([]);
  const [selected, setSelected] = useState<AIRunDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [filters, setFilters] = useState({ room_id: "", user_id: "", phase: "", success: "" });

  const query = useMemo(() => {
    const params = new URLSearchParams({ limit: "80" });
    Object.entries(filters).forEach(([key, value]) => {
      if (value) params.set(key, value);
    });
    return params.toString();
  }, [filters]);

  const loadRuns = async () => {
    setLoading(true);
    try {
      const res = await adminFetch(`/api/admin/ai-runs?${query}`);
      const data = await res.json();
      setRuns(Array.isArray(data) ? data : []);
    } finally {
      setLoading(false);
    }
  };

  const loadDetail = async (id: number) => {
    const res = await adminFetch(`/api/admin/ai-runs/${id}`);
    setSelected(await res.json());
  };

  useEffect(() => {
    loadRuns();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query]);

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-starbucks tracking-tight">AI Runs</h1>
        <p className="text-sm text-text-black-soft mt-0.5">Inspect LLM context, tool calls, and recorded actions.</p>
      </div>

      <Card padding="md" className="mb-5">
        <div className="grid grid-cols-1 gap-3 md:grid-cols-5">
          <input className="rounded-[4px] border border-gray-300 px-3 py-2 text-sm" placeholder="Room ID" value={filters.room_id} onChange={(e) => setFilters({ ...filters, room_id: e.target.value })} />
          <input className="rounded-[4px] border border-gray-300 px-3 py-2 text-sm" placeholder="User ID" value={filters.user_id} onChange={(e) => setFilters({ ...filters, user_id: e.target.value })} />
          <input className="rounded-[4px] border border-gray-300 px-3 py-2 text-sm" placeholder="Phase" value={filters.phase} onChange={(e) => setFilters({ ...filters, phase: e.target.value })} />
          <select className="rounded-[4px] border border-gray-300 px-3 py-2 text-sm" value={filters.success} onChange={(e) => setFilters({ ...filters, success: e.target.value })}>
            <option value="">Any result</option>
            <option value="true">Success</option>
            <option value="false">Failed</option>
          </select>
          <Button className="min-h-10 py-2 text-sm" variant="dark-outlined" onClick={loadRuns}>Refresh</Button>
        </div>
      </Card>

      <div className="grid grid-cols-1 gap-5 xl:grid-cols-[minmax(420px,0.9fr)_1.4fr]">
        <Card padding="md" className="overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full min-w-[720px] text-sm">
              <thead>
                <tr className="border-b border-cream text-left">
                  <th className="p-3">Time</th>
                  <th className="p-3">Bot</th>
                  <th className="p-3">Phase</th>
                  <th className="p-3">Model</th>
                  <th className="p-3">Tools</th>
                  <th className="p-3">Status</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr><td className="p-4 text-text-black-soft" colSpan={6}>Loading...</td></tr>
                ) : runs.length === 0 ? (
                  <tr><td className="p-4 text-text-black-soft" colSpan={6}>No AI runs found.</td></tr>
                ) : runs.map((run) => (
                  <tr key={run.id} onClick={() => loadDetail(run.id)} className="cursor-pointer border-b border-cream last:border-b-0 hover:bg-cream/50">
                    <td className="p-3 whitespace-nowrap">{new Date(run.created_at).toLocaleString()}</td>
                    <td className="p-3 font-mono text-xs">{run.user_id || "---"} / seat {run.seat}</td>
                    <td className="p-3">{run.phase || "---"}</td>
                    <td className="p-3">{run.provider}:{run.model}</td>
                    <td className="p-3">{run.tool_count}</td>
                    <td className="p-3">
                      <span className={`rounded-[4px] px-2 py-1 text-xs font-semibold ${run.success ? "bg-green-light text-starbucks" : "bg-red-error/10 text-red-error"}`}>
                        {run.success ? "OK" : "Failed"}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>

        <div className="space-y-5">
          {!selected ? (
            <Card padding="lg"><p className="text-text-black-soft">Select a run to inspect full context.</p></Card>
          ) : (
            <>
              <Card padding="md">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <h2 className="text-lg font-bold text-text-black">Run #{selected.id}</h2>
                    <p className="text-sm text-text-black-soft">{selected.provider}:{selected.model} · {selected.duration_ms}ms · {(selected.prompt_tokens + selected.completion_tokens).toLocaleString()} tokens</p>
                  </div>
                  <span className={`rounded-[4px] px-3 py-1 text-sm font-semibold ${selected.success ? "bg-green-light text-starbucks" : "bg-red-error/10 text-red-error"}`}>
                    {selected.success ? "Success" : "Failed"}
                  </span>
                </div>
                {selected.error_message && <p className="mt-3 text-sm text-red-error">{selected.error_message}</p>}
              </Card>

              <Card padding="md">
                <h3 className="mb-3 font-semibold text-text-black">Request Context</h3>
                <CodeBlock value={selected.request_json} />
              </Card>

              <Card padding="md">
                <h3 className="mb-3 font-semibold text-text-black">Model Response</h3>
                <CodeBlock value={selected.response_json} />
              </Card>

              <Card padding="md">
                <h3 className="mb-3 font-semibold text-text-black">Tool Executions</h3>
                <div className="space-y-4">
                  {(selected.tools || []).map((tool) => (
                    <div key={tool.id} className="rounded-[4px] border border-cream p-3">
                      <div className="mb-2 flex items-center justify-between gap-2">
                        <p className="font-semibold text-text-black">{tool.tool_name}</p>
                        <p className="text-xs text-text-black-soft">{tool.tool_type}</p>
                      </div>
                      <p className="mb-1 text-xs font-semibold uppercase text-text-black-soft">Args</p>
                      <CodeBlock value={tool.args_json} />
                      <p className="mb-1 mt-3 text-xs font-semibold uppercase text-text-black-soft">Result</p>
                      <CodeBlock value={tool.result_json} />
                    </div>
                  ))}
                  {(selected.tools || []).length === 0 && <p className="text-sm text-text-black-soft">No tool executions.</p>}
                </div>
              </Card>

              <Card padding="md">
                <h3 className="mb-3 font-semibold text-text-black">Game Actions</h3>
                <CodeBlock value={JSON.stringify(selected.game_actions || [])} />
              </Card>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
