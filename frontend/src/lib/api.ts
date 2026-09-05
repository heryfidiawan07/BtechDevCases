import { getAccessToken, setAccessToken } from "@/lib/storage";
import type { ApiErrorBody } from "@/lib/types";

export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
    public details?: Record<string, string>,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export class NetworkError extends Error {
  constructor(message = "Network request failed") {
    super(message);
    this.name = "NetworkError";
  }
}

type RequestOptions = {
  method?: "GET" | "POST";
  body?: unknown;
  headers?: Record<string, string>;
  timeoutMs?: number;
  retry?: number;
  auth?: boolean;
};

export async function api<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const {
    method = "GET",
    body,
    headers = {},
    timeoutMs = 12000,
    retry = method === "GET" ? 2 : 0,
    auth = true,
  } = options;

  let lastError: unknown;

  for (let attempt = 0; attempt <= retry; attempt += 1) {
    try {
      return await once<T>(path, {
        method,
        body,
        headers,
        timeoutMs,
        auth,
      });
    } catch (error) {
      lastError = error;
      if (!(error instanceof NetworkError) || attempt === retry) {
        throw error;
      }

      await sleep(300 * 2 ** attempt);
    }
  }

  throw lastError;
}

async function once<T>(
  path: string,
  options: Required<Pick<RequestOptions, "method" | "timeoutMs" | "auth">> &
    Pick<RequestOptions, "body" | "headers">,
): Promise<T> {
  const controller = new AbortController();
  const timer = window.setTimeout(() => controller.abort(), options.timeoutMs);

  try {
    const response = await fetch(`${API_URL}${path}`, {
      method: options.method,
      headers: buildHeaders(options.headers, options.auth, options.body !== undefined),
      body: options.body === undefined ? undefined : JSON.stringify(options.body),
      signal: controller.signal,
    });

    const nextToken = response.headers.get("X-Access-Token");
    if (nextToken) {
      setAccessToken(nextToken);
    }

    if (response.status === 204) {
      return undefined as T;
    }

    const payload = (await response.json().catch(() => null)) as
      | { data: T }
      | ApiErrorBody
      | null;

    if (!response.ok) {
      const errorBody = payload as ApiErrorBody | null;
      throw new ApiError(
        response.status,
        errorBody?.error.code ?? "HTTP_ERROR",
        errorBody?.error.message ?? "Request failed",
        errorBody?.error.details,
      );
    }

    return (payload as { data: T }).data;
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    throw new NetworkError();
  } finally {
    window.clearTimeout(timer);
  }
}

function buildHeaders(
  extra: Record<string, string> | undefined,
  auth: boolean,
  hasBody: boolean,
): HeadersInit {
  const headers = new Headers(extra);

  if (hasBody && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  if (auth) {
    const token = getAccessToken();
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }
  }

  return headers;
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => {
    window.setTimeout(resolve, ms);
  });
}
