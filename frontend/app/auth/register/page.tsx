"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";
import { apiClient } from "@/lib/api-client";

export default function RegisterPage() {
  const router = useRouter();
  const [nickname, setNickname] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    const result = await apiClient.register(nickname, password);
    if (result.error) {
      setError(result.error);
      setLoading(false);
      return;
    }

    if (result.data) {
      apiClient.setToken(result.data.token);
      router.push("/lobby");
    }
    setLoading(false);
  };

  return (
    <Card padding="lg">
      <div className="text-center mb-8">
        <div className="text-4xl mb-3">✨</div>
        <h1 className="text-2xl font-bold text-starbucks">Create Account</h1>
        <p className="text-sm text-text-black-soft mt-1">
          Pick a nickname and set a password
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-5">
        <Input
          label="Nickname"
          value={nickname}
          onChange={(e) => setNickname(e.target.value)}
        />
        <Input
          label="Password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          error={error}
        />
        <Button type="submit" fullWidth disabled={loading}>
          {loading ? "Creating account..." : "Create Account"}
        </Button>
      </form>

      <div className="mt-8 pt-6 border-t border-cream text-center">
        <p className="text-sm text-text-black-soft">
          Already have an account?{" "}
          <Link
            href="/auth/login"
            className="text-green-accent font-semibold hover:underline"
          >
            Sign In
          </Link>
        </p>
      </div>
    </Card>
  );
}
