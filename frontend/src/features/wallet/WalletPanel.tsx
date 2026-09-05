"use client";

import { useCallback, useEffect, useState } from "react";

import { useAuth } from "@/features/auth/AuthProvider";
import { TransferForm } from "@/features/wallet/TransferForm";
import { flushTransferQueue, listQueuedTransfers, removeQueuedTransfer } from "@/features/wallet/queue";
import { api, ApiError } from "@/lib/api";
import { formatRupiah } from "@/lib/money";
import type { QueuedTransfer, WalletPayload } from "@/lib/types";

export function WalletPanel() {
  const { user, logout } = useAuth();
  const [wallet, setWallet] = useState<WalletPayload | null>(null);
  const [queue, setQueue] = useState<QueuedTransfer[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    if (!user) {
      return;
    }

    setError("");
    try {
      const [nextWallet, nextQueue] = await Promise.all([
        api<WalletPayload>("/api/v1/wallet"),
        listQueuedTransfers(user.id),
      ]);
      setWallet(nextWallet);
      setQueue(nextQueue.filter((item) => item.status !== "synced"));
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        logout();
        return;
      }
      setError(err instanceof ApiError ? err.message : "Could not reach the ledger. Showing last known state.");
      setQueue(await listQueuedTransfers(user.id));
    } finally {
      setLoading(false);
    }
  }, [logout, user]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    const sync = () => {
      void (async () => {
        if (!user) {
          return;
        }
        await flushTransferQueue(user.id);
        await refresh();
      })();
    };

    window.addEventListener("online", sync);
    return () => window.removeEventListener("online", sync);
  }, [refresh, user]);

  if (!user) {
    return null;
  }

  return (
    <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.1fr)]">
      <section className="rounded-xl border border-ink/10 bg-white/70 p-5 shadow-sm">
        <p className="text-sm uppercase tracking-[0.16em] text-ink/50">Available balance</p>
        <p className="mt-2 font-display text-5xl text-ink">
          {wallet ? formatRupiah(wallet.balance) : loading ? "…" : "—"}
        </p>
        <p className="mt-3 text-sm text-ink/60">
          Transfers are written with an idempotency key, so a retry never double-debits.
        </p>
        <div className="mt-6">
          <h2 className="mb-3 font-display text-2xl">Send funds</h2>
          <TransferForm
            user={user}
            balance={wallet?.balance ?? null}
            onQueueChange={async () => {
              await refresh();
            }}
          />
        </div>
      </section>

      <section className="space-y-6">
        {error ? (
          <p className="rounded-md border border-amber-800/20 bg-amber-50 px-3 py-2 text-sm text-amber-950">
            {error}
          </p>
        ) : null}

        {queue.length > 0 ? (
          <div className="rounded-xl border border-ink/10 bg-white/70 p-5 shadow-sm">
            <h2 className="font-display text-2xl">Outbox</h2>
            <ul className="mt-3 space-y-3">
              {queue.map((item) => (
                <li key={item.id} className="rounded-md border border-ink/8 bg-paper/70 px-3 py-2">
                  <div className="flex items-center justify-between gap-3">
                    <p className="font-medium">To {item.recipient}</p>
                    <StatusPill status={item.status} />
                  </div>
                  <p className="text-sm text-ink/70">{formatRupiah(item.amount)}</p>
                  {item.error ? <p className="mt-1 text-sm text-terracotta">{item.error}</p> : null}
                  {item.status === "failed" ? (
                    <button
                      type="button"
                      className="mt-2 text-xs font-medium text-ink/60 underline hover:text-ink"
                      onClick={async () => {
                        await removeQueuedTransfer(item.id);
                        await refresh();
                      }}
                    >
                      Dismiss
                    </button>
                  ) : null}
                </li>
              ))}
            </ul>
          </div>
        ) : null}

        <div className="rounded-xl border border-ink/10 bg-white/70 p-5 shadow-sm">
          <h2 className="font-display text-2xl">Recent transfers</h2>
          {wallet && wallet.transfers.length === 0 ? (
            <p className="mt-3 text-sm text-ink/60">No transfers yet.</p>
          ) : (
            <ul className="mt-3 divide-y divide-ink/8">
              {wallet?.transfers.map((item) => (
                <li key={item.id} className="flex items-start justify-between gap-4 py-3">
                  <div>
                    <p className="font-medium">
                      {item.direction === "out"
                        ? `To ${item.recipientEmail}`
                        : `From ${item.senderEmail}`}
                    </p>
                    <p className="text-sm text-ink/55">
                      {item.notes || "No note"} · {new Date(item.createdAt).toLocaleString()}
                    </p>
                  </div>
                  <p className={item.direction === "out" ? "text-terracotta" : "text-moss-dark"}>
                    {item.direction === "out" ? "−" : "+"}
                    {formatRupiah(item.amount)}
                  </p>
                </li>
              ))}
            </ul>
          )}
        </div>
      </section>
    </div>
  );
}

function StatusPill({ status }: { status: QueuedTransfer["status"] }) {
  const label = status === "pending" ? "Pending" : status === "failed" ? "Failed" : "Synced";
  const tone =
    status === "pending"
      ? "bg-amber-100 text-amber-950"
      : status === "failed"
        ? "bg-red-100 text-red-900"
        : "bg-emerald-100 text-emerald-900";

  return <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${tone}`}>{label}</span>;
}