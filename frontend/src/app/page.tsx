"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { useAuth } from "@/features/auth/AuthProvider";
import { WalletPanel } from "@/features/wallet/WalletPanel";

export default function HomePage() {
  const { user, message, ready, logout } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (ready && !user) {
      router.replace("/login");
    }
  }, [ready, user, router]);

  if (!ready || !user) {
    return <CenteredNote text="Checking your session…" />;
  }

  return (
    <main className="mx-auto min-h-screen max-w-5xl px-4 py-8">
      <header className="mb-8 flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-sm uppercase tracking-[0.18em] text-moss">Outpost</p>
          <h1 className="mt-1 font-display text-4xl text-ink">{message}</h1>
          <p className="mt-2 max-w-xl text-ink/65">
            You will be signed out after 15 minutes of inactivity. Transfers survive weak
            connections and never debit twice.
          </p>
        </div>
        <button
          type="button"
          onClick={logout}
          className="rounded-md border border-ink/15 px-4 py-2 text-sm font-medium text-ink hover:bg-white/70"
        >
          Sign out
        </button>
      </header>
      <WalletPanel />
    </main>
  );
}

function CenteredNote({ text }: { text: string }) {
  return (
    <main className="grid min-h-screen place-items-center px-4">
      <p className="text-ink/60">{text}</p>
    </main>
  );
}
