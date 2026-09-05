"use client";

import { FormEvent, useState } from "react";

import { enqueueTransfer, flushTransferQueue } from "@/features/wallet/queue";
import { formatRupiah, formatRupiahLive, hasInsufficientFunds, sanitizeAmountInput, toApiAmount } from "@/lib/money";
import type { UserPublic } from "@/lib/types";

const DEMO_ACCOUNTS = ["fidiawan07@gmail.com", "heryfidiawan07@gmail.com"] as const;

type Props = {
  user: UserPublic;
  balance: string | null;
  onQueueChange: () => Promise<void>;
};

export function TransferForm({ user, balance, onQueueChange }: Props) {
  const [recipient, setRecipient] = useState("");
  const [amount, setAmount] = useState("");
  const [notes, setNotes] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [info, setInfo] = useState("");
  const [error, setError] = useState("");

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setSubmitting(true);
    setInfo("");
    setError("");

    const nextRecipient = recipient.trim().toLowerCase();
    const nextAmount = toApiAmount(amount);

    if (!isEmail(nextRecipient)) {
      setError("Recipient must be a valid email.");
      setSubmitting(false);
      return;
    }

    if (nextRecipient === user.email.toLowerCase()) {
      setError("You cannot transfer to yourself.");
      setSubmitting(false);
      return;
    }

    if (nextAmount === "" || nextAmount === "0") {
      setError("Enter a valid amount. Use digits and an optional comma for decimals.");
      setSubmitting(false);
      return;
    }

    if (balance !== null && hasInsufficientFunds(balance, amount)) {
      setError(
        `Insufficient funds. Available ${formatRupiah(balance)}, requested ${formatRupiahLive(amount)}.`,
      );
      setSubmitting(false);
      return;
    }

    const transferId = crypto.randomUUID();

    try {
      await enqueueTransfer({
        id: transferId,
        userId: user.id,
        recipient: nextRecipient,
        amount: nextAmount,
        notes: notes.trim(),
      });
      const queued = await flushTransferQueue(user.id);
      await onQueueChange();

      const item = queued.find((entry) => entry.id === transferId);
      if (item?.status === "failed") {
        setError(item.error || "Transfer failed.");
        return;
      }
      if (item?.status === "pending") {
        setRecipient("");
        setAmount("");
        setNotes("");
        setInfo(item.error || "Transfer queued. It will settle as soon as the network allows.");
        return;
      }

      setRecipient("");
      setAmount("");
      setNotes("");
      setInfo("Transfer sent.");
    } catch {
      setError("Could not queue this transfer. Try again.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={onSubmit} className="space-y-4" autoComplete="off">
      <Field
        label="Recipient"
        type="text"
        inputMode="email"
        autoComplete="off"
        name="transfer-recipient"
        value={recipient}
        onChange={setRecipient}
        placeholder={otherDemoEmail(user.email) || "heryfidiawan07@gmail.com"}
        required
      />
      <label className="block space-y-1.5">
        <span className="flex items-baseline justify-between gap-3 text-sm font-medium text-ink/80">
          Amount
          <span className="font-normal tabular-nums text-ink/55">{formatRupiahLive(amount)}</span>
        </span>
        <div className="flex items-center rounded-md border border-ink/15 bg-white/80 ring-moss/30 focus-within:ring-2">
          <span className="pl-3 text-sm text-ink/55">Rp</span>
          <input
            type="text"
            inputMode="decimal"
            autoComplete="off"
            name="transfer-amount"
            value={amount}
            required
            placeholder="50,00"
            onChange={(event) => setAmount(sanitizeAmountInput(event.target.value))}
            className="w-full rounded-md bg-transparent px-3 py-2 text-ink outline-none placeholder:text-ink/35"
          />
        </div>
      </label>
      <label className="block space-y-1.5">
        <span className="text-sm font-medium text-ink/80">Notes</span>
        <textarea
          value={notes}
          onChange={(event) => setNotes(event.target.value)}
          maxLength={280}
          rows={3}
          autoComplete="off"
          placeholder="Optional field note"
          className="w-full rounded-md border border-ink/15 bg-white/80 px-3 py-2 text-ink outline-none ring-moss/30 placeholder:text-ink/35 focus:ring-2"
        />
      </label>
      <button
        type="submit"
        disabled={submitting}
        className="w-full rounded-md bg-moss px-4 py-2.5 font-medium text-paper transition hover:bg-moss-dark disabled:cursor-not-allowed disabled:opacity-60"
      >
        {submitting ? "Sending…" : "Transfer funds"}
      </button>
      {error ? <p className="text-sm text-terracotta">{error}</p> : null}
      {info ? <p className="text-sm text-moss-dark">{info}</p> : null}
    </form>
  );
}

function Field(props: {
  label: string;
  type: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  required?: boolean;
  inputMode?: "decimal" | "email";
  autoComplete?: string;
  name?: string;
}) {
  return (
    <label className="block space-y-1.5">
      <span className="text-sm font-medium text-ink/80">{props.label}</span>
      <input
        type={props.type}
        inputMode={props.inputMode}
        autoComplete={props.autoComplete}
        name={props.name ?? props.label.toLowerCase()}
        value={props.value}
        required={props.required}
        placeholder={props.placeholder}
        onChange={(event) => props.onChange(event.target.value)}
        className="w-full rounded-md border border-ink/15 bg-white/80 px-3 py-2 text-ink outline-none ring-moss/30 placeholder:text-ink/35 focus:ring-2"
      />
    </label>
  );
}

function otherDemoEmail(email: string): string {
  const normalized = email.toLowerCase();
  return DEMO_ACCOUNTS.find((item) => item !== normalized) ?? "";
}

function isEmail(value: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
}
