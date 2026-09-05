"use client";

import { useRouter } from "next/navigation";
import { FormEvent, useEffect, useState } from "react";

import { AuthField, AuthFooter, AuthShell } from "@/components/AuthShell";
import { ApiError } from "@/lib/api";
import { useAuth } from "@/features/auth/AuthProvider";

export default function RegisterPage() {
  const { user, ready, register } = useAuth();
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
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

    if (password !== confirmPassword) {
      setError("Passwords do not match");
      setSubmitting(false);
      return;
    }

    try {
      await register(email.trim(), password, confirmPassword);
      router.replace("/");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not reach the server. Try again.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <AuthShell title="Create account" subtitle="Every new ledger starts with Rp 100.000,00 so you can demo transfers immediately.">
      <form onSubmit={onSubmit} className="space-y-4">
        <AuthField label="Email" type="email" value={email} onChange={setEmail} required />
        <AuthField label="Password" type="password" value={password} onChange={setPassword} required />
        <AuthField
          label="Confirm password"
          type="password"
          value={confirmPassword}
          onChange={setConfirmPassword}
          required
        />
        {error ? <p className="text-sm text-terracotta">{error}</p> : null}
        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-md bg-moss px-4 py-2.5 font-medium text-paper hover:bg-moss-dark disabled:opacity-60"
        >
          {submitting ? "Creating…" : "Create account"}
        </button>
      </form>
      <AuthFooter href="/login" prompt="Already have an account?" label="Sign in" />
    </AuthShell>
  );
}
