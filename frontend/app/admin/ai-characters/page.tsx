"use client";

import { useState, useEffect } from "react";

interface AICharacter {
  id: number;
  name: string;
  play_style: string;
  enabled: boolean;
}

export default function AICharactersPage() {
  const [characters, setCharacters] = useState<AICharacter[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // TODO: fetch from API when backend endpoint is available
    setLoading(false);
  }, []);

  if (loading) {
    return <p className="text-gray-500">Loading...</p>;
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-semibold text-gray-800">AI Characters</h1>
        <button className="px-4 py-2 bg-green-600 text-white rounded-full text-sm font-semibold hover:bg-green-700">
          + Add Character
        </button>
      </div>

      {characters.length === 0 && (
        <div className="bg-white rounded-xl shadow p-8 text-center">
          <p className="text-gray-500">No AI characters yet.</p>
          <p className="text-sm text-gray-400 mt-1">
            Create characters to populate bot players.
          </p>
        </div>
      )}

      {characters.length > 0 && (
        <div className="bg-white rounded-xl shadow overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200">
                <th className="text-left p-4 text-sm font-semibold text-gray-700">Name</th>
                <th className="text-left p-4 text-sm font-semibold text-gray-700">Play Style</th>
                <th className="text-left p-4 text-sm font-semibold text-gray-700">Status</th>
                <th className="text-right p-4 text-sm font-semibold text-gray-700">Actions</th>
              </tr>
            </thead>
            <tbody>
              {characters.map((c) => (
                <tr key={c.id} className="border-b border-gray-100">
                  <td className="p-4 text-sm">{c.name}</td>
                  <td className="p-4 text-sm capitalize">{c.play_style}</td>
                  <td className="p-4 text-sm">
                    <span className={`px-2 py-1 rounded-full text-xs ${c.enabled ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-500"}`}>
                      {c.enabled ? "Active" : "Disabled"}
                    </span>
                  </td>
                  <td className="p-4 text-right">
                    <button className="text-green-600 text-sm hover:underline">Edit</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
