"use client";

import { useEffect, useState, useCallback } from "react";
import { Card } from "@/components/ui/Card";

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
    const token = localStorage.getItem("token");
    if (!token) return;
    setLoading(true);
    const params = new URLSearchParams({ page: String(page), size: "20" });
    if (query) params.set("q", query);
    try {
      const res = await fetch(`/api/admin/users?${params}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = await res.json();
      setUsers(data.users || []);
      setTotal(data.total || 0);
    } catch { /* ignore */ }
    setLoading(false);
  }, [page, query]);

  useEffect(() => { fetchUsers(); }, [fetchUsers]);

  const toggleBan = async (userId: number, currentStatus: number) => {
    const token = localStorage.getItem("token");
    if (!token) return;
    const newStatus = currentStatus === 1 ? 0 : 1;
    await fetch(`/api/admin/users/${userId}/status`, {
      method: "PUT",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
      body: JSON.stringify({ status: newStatus }),
    });
    fetchUsers();
  };

  return (
    <div>
      <h1 className="text-2xl font-semibold text-starbucks mb-6">User Management</h1>
      <Card>
        <div className="mb-4">
          <input
            type="text"
            placeholder="Search by nickname..."
            className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm"
            value={query}
            onChange={(e) => { setQuery(e.target.value); setPage(1); }}
          />
        </div>
        {loading ? (
          <p className="text-gray-500 py-4 text-center">Loading...</p>
        ) : users.length === 0 ? (
          <p className="text-gray-500 py-4 text-center">No users found.</p>
        ) : (
          <>
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-200">
                  <th className="text-left p-3 text-sm font-semibold">ID</th>
                  <th className="text-left p-3 text-sm font-semibold">Nickname</th>
                  <th className="text-left p-3 text-sm font-semibold">Role</th>
                  <th className="text-left p-3 text-sm font-semibold">Status</th>
                  <th className="text-right p-3 text-sm font-semibold">Actions</th>
                </tr>
              </thead>
              <tbody>
                {users.map((u) => (
                  <tr key={u.id} className="border-b border-gray-100 hover:bg-gray-50">
                    <td className="p-3 text-sm">{u.id}</td>
                    <td className="p-3 text-sm">{u.nickname}</td>
                    <td className="p-3 text-sm capitalize">{u.role}</td>
                    <td className="p-3 text-sm">
                      <span className={`px-2 py-1 rounded-full text-xs ${u.status === 1 ? "bg-green-100 text-green-700" : "bg-red-100 text-red-700"}`}>
                        {u.status === 1 ? "Active" : "Banned"}
                      </span>
                    </td>
                    <td className="p-3 text-right">
                      <button
                        className={`text-sm hover:underline ${u.status === 1 ? "text-red-600" : "text-green-600"}`}
                        onClick={() => toggleBan(u.id, u.status)}
                      >
                        {u.status === 1 ? "Ban" : "Unban"}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {total > 20 && (
              <div className="flex items-center justify-center gap-2 mt-4 pb-2">
                <button
                  className="px-3 py-1 text-sm border rounded disabled:opacity-50"
                  disabled={page <= 1}
                  onClick={() => setPage((p) => p - 1)}
                >
                  Previous
                </button>
                <span className="px-3 py-1 text-sm text-gray-600">
                  Page {page} of {Math.ceil(total / 20)}
                </span>
                <button
                  className="px-3 py-1 text-sm border rounded disabled:opacity-50"
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
