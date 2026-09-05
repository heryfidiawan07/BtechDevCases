const TOKEN_KEY = "outpost.accessToken";
const DRAFT_KEY = "outpost.transferDraft";

let memoryToken: string | null = null;

export function getAccessToken(): string | null {
  if (memoryToken) {
    return memoryToken;
  }
  if (typeof window === "undefined") {
    return null;
  }

  memoryToken = sessionStorage.getItem(TOKEN_KEY);
  return memoryToken;
}

export function setAccessToken(token: string | null): void {
  memoryToken = token;
  if (typeof window === "undefined") {
    return;
  }

  if (token) {
    sessionStorage.setItem(TOKEN_KEY, token);
    return;
  }

  sessionStorage.removeItem(TOKEN_KEY);
}

export function clearAllTransferDrafts(): void {
  if (typeof window === "undefined") {
    return;
  }

  const keys: string[] = [];
  for (let index = 0; index < sessionStorage.length; index += 1) {
    const key = sessionStorage.key(index);
    if (key && (key === DRAFT_KEY || key.startsWith(`${DRAFT_KEY}.`))) {
      keys.push(key);
    }
  }

  keys.forEach((key) => sessionStorage.removeItem(key));
}
