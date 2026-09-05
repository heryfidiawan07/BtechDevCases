"use client";

import { useRouter } from "next/navigation";
import { FormEvent, useEffect, useState } from "react";

import { AuthField, AuthFooter, AuthShell } from "@/components/AuthShell";
import { ApiError } from "@/lib/api";
import { useAuth } from "@/features/auth/AuthProvider";

export default function LoginPage() {
  const { user, ready, login } = useAuth();
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (ready && user) {
      router.replace("/");
    }
  }, [ready, user, router]);

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setSubmitting(true);
    setError("");

    try {
      await login(email.trim(), password);
      router.replace("/");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not reach the server. Try again.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <AuthShell
      title="Sign in"
      subtitle="Open your field ledger. Demo accounts are ready if you want a faster start."
    >
      <form onSubmit={onSubmit} className="space-y-4">
        <AuthField label="Email" type="email" value={email} onChange={setEmail} required />
        <AuthField label="Password" type="password" value={password} onChange={setPassword} required />
        {error ? <p className="text-sm text-terracotta">{error}</p> : null}
        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-md bg-moss px-4 py-2.5 font-medium text-paper hover:bg-moss-dark disabled:opacity-60"
        >
          {submitting ? "Signing in…" : "Sign in"}
        </button>
      </form>
      <AuthFooter href="/register" prompt="New here?" label="Create an account" />
      <p className="mt-3 text-xs text-ink/50">
        Demo: fidiawan07@gmail.com or heryfidiawan07@gmail.com / 12345678
      </p>
    </AuthShell>
  );
}
