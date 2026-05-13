"use client";

import { useEffect, useState, useCallback } from "react";
import { Card } from "@/components/ui/Card";
import { adminFetch } from "@/lib/admin-fetch";

interface User {
  id: number;
  nickname: string;
  role: string;
  status: number;
  created_at: string;
}

export default function AdminUsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(true);

  const fetchUsers = useCallback(async () => {
    setLoading(true);
    const params = new URLSearchParams({ page: String(page), size: "20" });
    if (query) params.set("q", query);
    try {
      const res = await adminFetch(`/api/admin/users?${params}`);
      const data = await res.json();
      setUsers(data.users || []);
      setTotal(data.total || 0);
    } catch { /* ignore */ }
    setLoading(false);
  }, [page, query]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    fetchUsers();
  }, [fetchUsers]);

  const toggleBan = async (userId: number, currentStatus: number) => {
    const newStatus = currentStatus === 1 ? 0 : 1;
    try {
      await adminFetch(`/api/admin/users/${userId}/status`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ status: newStatus }),
      });
      fetchUsers();
    } catch { /* ignore */ }
  };

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-starbucks tracking-tight">User Management</h1>
        <p className="text-sm text-text-black-soft mt-0.5">Manage player accounts</p>
      </div>
      <Card padding="md" className="overflow-hidden">
        <div className="mb-4">
          <input
            type="text"
            placeholder="Search by nickname..."
            className="w-full px-3 py-2 border border-gray-300 rounded-[4px] text-sm text-text-black outline-none transition-all duration-200 focus:border-green-accent"
            value={query}
            onChange={(e) => { setQuery(e.target.value); setPage(1); }}
          />
        </div>
        {loading ? (
          <p className="text-text-black-soft text-center py-4">Loading...</p>
        ) : users.length === 0 ? (
          <p className="text-text-black-soft text-center py-4">No users found.</p>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full min-w-[500px]">
              <thead>
                <tr className="border-b border-cream">
                  <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">ID</th>
                  <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">Nickname</th>
                  <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">Role</th>
                  <th className="text-left p-3 text-sm font-semibold text-text-black tracking-tight">Status</th>
                  <th className="text-right p-3 text-sm font-semibold text-text-black tracking-tight">Actions</th>
                </tr>
              </thead>
              <tbody>
                {users.map((u) => (
                  <tr key={u.id} className="border-b border-cream last:border-b-0 hover:bg-cream/50 transition-colors">
                    <td className="p-3 text-sm text-text-black">{u.id}</td>
                    <td className="p-3 text-sm text-text-black">{u.nickname}</td>
                    <td className="p-3 text-sm text-text-black capitalize">{u.role}</td>
                    <td className="p-3 text-sm">
                      <span className={`inline-block px-3 py-1 rounded-pill text-xs font-semibold tracking-tight ${u.status === 1 ? "bg-green-light text-starbucks" : "bg-red-error/10 text-red-error"}`}>
                        {u.status === 1 ? "Active" : "Banned"}
                      </span>
                    </td>
                    <td className="p-3 text-right">
                      <button
                        className={`text-sm font-semibold hover:underline transition-colors ${u.status === 1 ? "text-red-error" : "text-green-accent"}`}
                        onClick={() => toggleBan(u.id, u.status)}
                      >
                        {u.status === 1 ? "Ban" : "Unban"}
                      </button>
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
