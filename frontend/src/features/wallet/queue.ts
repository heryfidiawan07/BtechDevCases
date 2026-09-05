import { api, ApiError, NetworkError } from "@/lib/api";
import type { QueuedTransfer, TransferRecord } from "@/lib/types";

const DB_NAME = "outpost";
const STORE = "transfers";
const VERSION = 3;

function openDb(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, VERSION);
    request.onupgradeneeded = () => {
      const db = request.result;
      if (!db.objectStoreNames.contains(STORE)) {
        db.createObjectStore(STORE, { keyPath: "id" });
        return;
      }

      request.transaction?.objectStore(STORE).clear();
    };
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

async function withStore<T>(
  mode: IDBTransactionMode,
  fn: (store: IDBObjectStore) => IDBRequest<T> | void,
): Promise<T> {
  const db = await openDb();

  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE, mode);
    const store = tx.objectStore(STORE);
    const request = fn(store);

    tx.oncomplete = () => {
      db.close();
      resolve((request ? request.result : undefined) as T);
    };
    tx.onerror = () => {
      db.close();
      reject(tx.error);
    };
  });
}

function ownedBy(item: QueuedTransfer, userId: number): boolean {
  return item.userId === userId;
}

export async function enqueueTransfer(
  item: Omit<QueuedTransfer, "status" | "createdAt" | "attempts">,
): Promise<QueuedTransfer> {
  const record: QueuedTransfer = {
    ...item,
    status: "pending",
    createdAt: Date.now(),
    attempts: 0,
  };

  await withStore("readwrite", (store) => store.put(record));
  return record;
}

export async function listQueuedTransfers(userId: number): Promise<QueuedTransfer[]> {
  const items = await withStore<QueuedTransfer[]>("readonly", (store) => store.getAll());
  return (items ?? [])
    .filter((item) => ownedBy(item, userId) && item.status === "pending")
    .sort((a, b) => b.createdAt - a.createdAt);
}

export async function saveQueuedTransfer(item: QueuedTransfer): Promise<void> {
  await withStore("readwrite", (store) => store.put(item));
}

export async function removeQueuedTransfer(id: string): Promise<void> {
  await withStore("readwrite", (store) => store.delete(id));
}

export async function flushTransferQueue(userId: number): Promise<QueuedTransfer[]> {
  const items = await listQueuedTransfers(userId);
  const pending = items.filter((item) => item.status === "pending");
  const failed: QueuedTransfer[] = [];

  for (const item of pending) {
    item.attempts += 1;
    try {
      await api<TransferRecord>("/api/v1/wallet/transfers", {
        method: "POST",
        body: {
          recipient: item.recipient,
          amount: item.amount,
          notes: item.notes,
        },
        headers: { "Idempotency-Key": item.id },
        retry: 0,
      });
      item.status = "synced";
      item.error = undefined;
      await removeQueuedTransfer(item.id);
    } catch (error) {
      if (error instanceof NetworkError) {
        item.status = "pending";
        item.error = "Waiting for a stable connection";
        await saveQueuedTransfer(item);
        continue;
      }

      item.status = "failed";
      item.error = error instanceof ApiError ? error.message : "Transfer failed";
      await removeQueuedTransfer(item.id);
      failed.push(item);
    }
  }

  return [...failed, ...(await listQueuedTransfers(userId))];
}

export async function clearAllQueuedTransfers(): Promise<void> {
  await withStore("readwrite", (store) => store.clear());
}
