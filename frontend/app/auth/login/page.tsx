"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";
import { apiClient } from "@/lib/api-client";

export default function LoginPage() {
  const router = useRouter();
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    const result = await apiClient.login(password);
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
      <h1 className="text-2xl font-semibold text-starbucks mb-6">Sign In</h1>
      <form onSubmit={handleSubmit} className="space-y-4">
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
      <p className="mt-4 text-center text-sm text-text-black-soft">
        Don&apos;t have an account?{" "}
        <Link href="/auth/register" className="text-green-accent underline">
          Register
        </Link>
      </p>
    </Card>
  );
}
