"use client";

import { useState, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";
import { apiClient } from "@/lib/api-client";

function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const redirectTo = searchParams.get("redirect") || "/lobby";
  const [nickname, setNickname] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    if (!nickname.trim()) {
      setError("Nickname is required");
      setLoading(false);
      return;
    }

    const result = await apiClient.login(password, nickname.trim());
    if (result.error) {
      setError(result.error);
      setLoading(false);
      return;
    }

    if (result.data) {
      apiClient.setToken(result.data.token);
      router.push(redirectTo);
    }
    setLoading(false);
  };

  return (
    <Card padding="lg">
      <div className="text-center mb-8">
        <div className="text-4xl mb-3">🃏</div>
        <h1 className="text-2xl font-bold text-starbucks">Sign In</h1>
        <p className="text-sm text-text-black-soft mt-1">
          Sign in with your nickname and password
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
          {loading ? "Signing in..." : "Sign In"}
        </Button>
      </form>

      <div className="mt-8 pt-6 border-t border-cream text-center">
        <p className="text-sm text-text-black-soft">
          Don&apos;t have an account?{" "}
          <Link
            href="/auth/register"
            className="text-green-accent font-semibold hover:underline"
          >
            Register
          </Link>
        </p>
      </div>
    </Card>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={null}>
      <LoginForm />
    </Suspense>
  );
}
