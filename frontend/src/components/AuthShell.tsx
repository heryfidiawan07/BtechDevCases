"use client";

import Link from "next/link";

export function AuthShell({
  title,
  subtitle,
  children,
}: {
  title: string;
  subtitle: string;
  children: React.ReactNode;
}) {
  return (
    <main className="mx-auto grid min-h-screen max-w-md place-items-center px-4 py-10">
      <section className="w-full rounded-xl border border-ink/10 bg-white/75 p-6 shadow-sm">
        <p className="text-sm uppercase tracking-[0.18em] text-moss">Outpost</p>
        <h1 className="mt-2 font-display text-4xl">{title}</h1>
        <p className="mt-2 text-sm text-ink/65">{subtitle}</p>
        <div className="mt-6">{children}</div>
      </section>
    </main>
  );
}

export function AuthField(props: {
  label: string;
  type: string;
  value: string;
  onChange: (value: string) => void;
  required?: boolean;
}) {
  return (
    <label className="block space-y-1.5">
      <span className="text-sm font-medium text-ink/80">{props.label}</span>
      <input
        type={props.type}
        value={props.value}
        required={props.required}
        onChange={(event) => props.onChange(event.target.value)}
        className="w-full rounded-md border border-ink/15 bg-white/80 px-3 py-2 outline-none ring-moss/30 focus:ring-2"
      />
    </label>
  );
}

export function AuthFooter({ href, label, prompt }: { href: string; label: string; prompt: string }) {
  return (
    <p className="mt-5 text-sm text-ink/65">
      {prompt}{" "}
      <Link href={href} className="font-medium text-moss-dark underline">
        {label}
      </Link>
    </p>
  );
}
