export function formatRupiah(value: string): string {
  const amount = parseMoney(value);
  if (amount === null) {
    return value;
  }

  const formatted = new Intl.NumberFormat("id-ID", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount);

  return `Rp ${formatted}`;
}

export function formatRupiahLive(input: string): string {
  const trimmed = input.trim();
  if (trimmed === "" || trimmed === ",") {
    return "Rp";
  }

  return formatRupiah(trimmed);
}

export function sanitizeAmountInput(raw: string): string {
  const cleaned = raw.replace(/[^\d,]/g, "");
  const comma = cleaned.indexOf(",");
  if (comma === -1) {
    return cleaned;
  }

  const intPart = cleaned.slice(0, comma).replace(/,/g, "") || "0";
  const decPart = cleaned.slice(comma + 1).replace(/,/g, "").slice(0, 2);
  if (decPart === "") {
    return `${intPart},`;
  }

  return `${intPart},${decPart}`;
}

export function toApiAmount(input: string): string {
  const trimmed = input.trim().replace(/,$/, "");
  if (trimmed === "" || trimmed === ",") {
    return "";
  }

  const [wholeRaw, decRaw] = trimmed.split(",");
  const whole = wholeRaw === "" ? "0" : wholeRaw;
  if (decRaw === undefined) {
    return whole;
  }

  return `${whole}.${decRaw}`;
}

export function parseMoney(value: string): number | null {
  const trimmed = value.trim();
  if (trimmed === "") {
    return null;
  }

  const normalized = trimmed.includes(",")
    ? trimmed.replace(/\./g, "").replace(",", ".")
    : trimmed;
  const amount = Number(normalized);
  if (Number.isNaN(amount)) {
    return null;
  }

  return amount;
}

export function hasInsufficientFunds(balance: string, amount: string): boolean {
  const available = parseMoney(balance);
  const requested = parseMoney(amount);
  if (available === null || requested === null) {
    return false;
  }

  return requested > available;
}
