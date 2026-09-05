"use client";

import { useEffect, useState } from "react";

export function OfflineBanner() {
  const [online, setOnline] = useState(true);

  useEffect(() => {
    const sync = () => setOnline(navigator.onLine);
    sync();
    window.addEventListener("online", sync);
    window.addEventListener("offline", sync);
    return () => {
      window.removeEventListener("online", sync);
      window.removeEventListener("offline", sync);
    };
  }, []);

  if (online) {
    return null;
  }

  return (
    <div className="border-b border-amber-800/20 bg-amber-100 px-4 py-2 text-center text-sm text-amber-950">
      Connection dropped. Queued transfers stay in the outbox and retry when the signal returns.
    </div>
  );
}
