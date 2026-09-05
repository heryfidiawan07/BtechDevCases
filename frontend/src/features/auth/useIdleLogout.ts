"use client";

import { useEffect, useRef } from "react";

const IDLE_MS = 15 * 60 * 1000;
const REFRESH_THROTTLE_MS = 60 * 1000;

type Options = {
  enabled: boolean;
  onIdle: () => void;
  onActivity: () => void;
};

export function useIdleLogout({ enabled, onIdle, onActivity }: Options) {
  const idleTimer = useRef<number | null>(null);
  const lastRefresh = useRef(0);

  useEffect(() => {
    if (!enabled) {
      return;
    }

    const reset = () => {
      if (idleTimer.current) {
        window.clearTimeout(idleTimer.current);
      }
      idleTimer.current = window.setTimeout(onIdle, IDLE_MS);

      const now = Date.now();
      if (now - lastRefresh.current >= REFRESH_THROTTLE_MS) {
        lastRefresh.current = now;
        onActivity();
      }
    };

    reset();

    const windowEvents: Array<keyof WindowEventMap> = ["pointerdown", "keydown"];
    windowEvents.forEach((event) => window.addEventListener(event, reset));
    document.addEventListener("visibilitychange", reset);

    return () => {
      if (idleTimer.current) {
        window.clearTimeout(idleTimer.current);
      }
      windowEvents.forEach((event) => window.removeEventListener(event, reset));
      document.removeEventListener("visibilitychange", reset);
    };
  }, [enabled, onIdle, onActivity]);
}
