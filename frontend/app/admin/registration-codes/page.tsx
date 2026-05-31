"use client";

import { useEffect, useState, useCallback } from "react";
import { Card } from "@/components/ui/Card";
import {
  fetchRegistrationCodes,
  generateRegistrationCodes,
  disableRegistrationCode,
  type RegistrationCode,
} from "@/lib/admin-fetch";

export default function AdminRegistrationCodesPage() {
  const [codes, setCodes] = useState<RegistrationCode[]>([]);
  const [total, setTotal] = useState(0);
  const [totalUnused, setTotalUnused] = useState(0);
  const [totalUsed, setTotalUsed] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [showGenerate, setShowGenerate] = useState(false);
  const [genCount, setGenCount] = useState(1);
  const [genNote, setGenNote] = useState("");
  const [genResult, setGenResult] = useState<RegistrationCode[] | null>(null);
  const [genLoading, setGenLoading] = useState(false);

  const fetchCodes = useCallback(async () => {
    setLoading(true);
    try {
      const data = await fetchRegistrationCodes(page, 20);
      setCodes(data.codes || []);
      setTotal(data.total || 0);
    } catch { /* ignore */ }
    setLoading(false);
  }, [page]);

  const fetchStats = useCallback(async () => {
    try {
      const [unused, used] = await Promise.all([
        fetchRegistrationCodes(1, 1, false),
        fetchRegistrationCodes(1, 1, true),
      ]);
      setTotalUnused(unused.total || 0);
      setTotalUsed(used.total || 0);
    } catch { /* ignore */ }
  }, []);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    fetchCodes();
    fetchStats();
  }, [fetchCodes, fetchStats]);

  const handleGenerate = async () => {
    setGenLoading(true);
    try {
      const result = await generateRegistrationCodes(genCount, genNote || undefined);
      setGenResult(result.codes);
      fetchCodes();
      fetchStats();
    } catch { /* ignore */ }
    setGenLoading(false);
  };

  const handleDisable = async (id: number) => {
    if (!confirm("确定要禁用这个邀请码吗？")) return;
    try {
      await disableRegistrationCode(id);
      fetchCodes();
      fetchStats();
    } catch { /* ignore */ }
  };

  const handleCopy = (code: string) => {
    navigator.clipboard.writeText(code);
  };

  const statusBadge = (c: RegistrationCode) => {
    if (c.is_disabled) {
      return <span className="inline-block px-3 py-1 rounded-pill text-xs font-semibold tracking-tight bg-red-error/10 text-red-error">已禁用</span>;
    }
    if (c.is_used) {
      return <span className="inline-block px-3 py-1 rounded-pill text-xs font-semibold tracking-tight bg-blue-100 text-blue-700">已使用</span>;
    }
    return <span className="inline-block px-3 py-1 rounded-pill text-xs font-semibold tracking-tight bg-green-light text-starbucks">未使用</span>;
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-starbucks tracking-tight">邀请码管理</h1>
          <p className="text-sm text-text-black-soft mt-0.5">生成和管理注册邀请码</p>
        </div>
        <button
          className="px-4 py-2 rounded-pill bg-green-accent text-white text-sm font-bold hover:bg-starbucks transition-colors"
          onClick={() => setShowGenerate(true)}
        >
          生成邀请码
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4 mb-6">
        <Card padding="md" className="text-center">
          <div className="text-2xl font-black text-starbucks">{total}</div>
          <div className="text-sm text-text-black-soft mt-1">总邀请码</div>
        </Card>
        <Card padding="md" className="text-center">
          <div className="text-2xl font-black text-starbucks">{totalUsed}</div>
          <div className="text-sm text-text-black-soft mt-1">已使用</div>
        </Card>
        <Card padding="md" className="text-center">
          <div className="text-2xl font-black text-starbucks">{totalUnused}</div>
          <div className="text-sm text-text-black-soft mt-1">未使用</div>
        </Card>
      </div>

      {/* Generate Modal */}
      {showGenerate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={() => { setShowGenerate(false); setGenResult(null); }}>
          <div className="bg-white rounded-xl p-6 w-full max-w-md mx-4 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <h2 className="text-lg font-bold text-starbucks mb-4">生成邀请码</h2>

            {genResult ? (
              <div>
                <p className="text-sm text-text-black-soft mb-3">已生成 {genResult.length} 个邀请码：</p>
                <div className="space-y-2 mb-4 max-h-60 overflow-y-auto">
                  {genResult.map((c) => (
                    <div key={c.id} className="flex items-center justify-between bg-cream/50 rounded-lg px-3 py-2">
                      <code className="text-lg font-bold text-starbucks tracking-widest">{c.code}</code>
                      <button
                        className="text-xs font-bold text-green-accent hover:underline"
                        onClick={() => handleCopy(c.code)}
                      >
                        复制
                      </button>
                    </div>
                  ))}
                </div>
                <button
                  className="w-full py-2 rounded-pill bg-green-accent text-white text-sm font-bold"
                  onClick={() => { setShowGenerate(false); setGenResult(null); }}
                >
                  完成
                </button>
              </div>
            ) : (
              <div>
                <label className="block text-sm font-bold text-text-black mb-2">数量</label>
                <input
                  type="number"
                  min={1}
                  max={100}
                  value={genCount}
                  onChange={(e) => setGenCount(Math.max(1, Math.min(100, Number(e.target.value))))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm mb-4"
                />
                <label className="block text-sm font-bold text-text-black mb-2">备注（可选）</label>
                <input
                  type="text"
                  value={genNote}
                  onChange={(e) => setGenNote(e.target.value)}
                  placeholder="例如：给朋友"
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm mb-4"
                />
                <div className="flex gap-2">
                  <button
                    className="flex-1 py-2 rounded-pill border border-cream text-text-black-soft text-sm font-bold"
                    onClick={() => setShowGenerate(false)}
                  >
                    取消
                  </button>
                  <button
                    className="flex-1 py-2 rounded-pill bg-green-accent text-white text-sm font-bold disabled:opacity-50"
                    disabled={genLoading}
                    onClick={handleGenerate}
                  >
                    {genLoading ? "生成中..." : "生成"}
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Table */}
      <Card padding="md" className="overflow-hidden">
        {loading ? (
          <p className="text-text-black-soft text-center py-4">Loading...</p>
        ) : codes.length === 0 ? (
          <p className="text-text-black-soft text-center py-4">暂无邀请码，点击上方按钮生成。</p>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full min-w-[600px]">
                <thead>
                  <tr className="border-b border-cream">
                    <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">邀请码</th>
                    <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">状态</th>
                    <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">备注</th>
                    <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">使用者</th>
                    <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">创建时间</th>
                    <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">使用时间</th>
                    <th className="text-right p-3 text-sm font-semibold text-text-black tracking-tight">操作</th>
                  </tr>
                </thead>
                <tbody>
                  {codes.map((c) => (
                    <tr key={c.id} className="border-b border-cream last:border-b-0 hover:bg-cream/50 transition-colors">
                      <td className="p-3 text-sm">
                        <code className="text-base font-bold text-starbucks tracking-widest">{c.code}</code>
                      </td>
                      <td className="p-3 text-sm">{statusBadge(c)}</td>
                      <td className="p-3 text-sm text-text-black-soft">{c.note || "-"}</td>
                      <td className="p-3 text-sm text-text-black">{c.used_by_nickname || "-"}</td>
                      <td className="p-3 text-sm text-text-black-soft">{c.created_at ? new Date(c.created_at).toLocaleDateString("zh-CN") : "-"}</td>
                      <td className="p-3 text-sm text-text-black-soft">{c.used_at ? new Date(c.used_at).toLocaleDateString("zh-CN") : "-"}</td>
                      <td className="p-3 text-right">
                        <button
                          className="text-sm font-semibold text-green-accent hover:underline mr-2"
                          onClick={() => handleCopy(c.code)}
                        >
                          复制
                        </button>
                        {!c.is_used && !c.is_disabled && (
                          <button
                            className="text-sm font-semibold text-red-error hover:underline"
                            onClick={() => handleDisable(c.id)}
                          >
                            禁用
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {total > 20 && (
              <div className="flex items-center justify-center gap-2 mt-4 pb-2">
                <button
                  className="px-3 py-1.5 text-sm font-semibold rounded-pill border border-green-accent text-green-accent transition-all duration-200 active:scale-[0.95] disabled:opacity-40 disabled:cursor-not-allowed"
                  disabled={page <= 1}
                  onClick={() => setPage((p) => p - 1)}
                >
                  Previous
                </button>
                <span className="px-3 py-1 text-sm text-text-black-soft">
                  Page {page} of {Math.ceil(total / 20)}
                </span>
                <button
                  className="px-3 py-1.5 text-sm font-semibold rounded-pill border border-green-accent text-green-accent transition-all duration-200 active:scale-[0.95] disabled:opacity-40 disabled:cursor-not-allowed"
                  disabled={page >= Math.ceil(total / 20)}
                  onClick={() => setPage((p) => p + 1)}
                >
                  Next
                </button>
              </div>
            )}
          </>
        )}
      </Card>
    </div>
  );
}
