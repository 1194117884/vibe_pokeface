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
      <h1 className="text-2xl font-semibold text-starbucks mb-6">Register</h1>
      <form onSubmit={handleSubmit} className="space-y-4">
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
      <p className="mt-4 text-center text-sm text-text-black-soft">
        Already have an account?{" "}
        <Link href="/auth/login" className="text-green-accent underline">
          Sign In
        </Link>
      </p>
    </Card>
  );
}
