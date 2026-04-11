import {
  dehydrate,
  hydrate,
  type DehydratedState,
  type Query,
  type QueryClient,
  type QueryKey,
} from "@tanstack/react-query";

const DB_NAME = "fitcommerce-offline-cache";
const STORE_NAME = "app-state";
const QUERY_CACHE_KEY = "react-query";
const QUERY_CACHE_VERSION = 1;
const MUTATION_QUEUE_KEY = "offline-mutations";
const MUTATION_QUEUE_VERSION = 1;
const ITEM_ID_MAP_KEY = "offline-item-id-map";

const OFFLINE_QUERY_ROOTS = new Set([
  "dashboard",
  "items",
  "campaigns",
  "orders",
  "inventory",
  "purchase-orders",
  "suppliers",
  "variances",
  "reports",
  "exports",
  "warehouse-bins",
  "locations",
  "members",
  "coaches",
  "landed-costs",
  // Admin read domains — persisted so admins can review data while offline
  "users",
  "audit",
  "backups",
  "biometrics",
  "encryption-keys",
  "retention-policies",
]);

interface PersistedQuerySnapshot {
  version: number;
  updatedAt: number;
  state: DehydratedState;
}

export type OfflineMutationType = "create-item" | "update-item" | "run-export";
export type OfflineMutationStatus = "pending" | "failed";

export interface OfflineMutationEntry {
  id: string;
  type: OfflineMutationType;
  payload: Record<string, unknown>;
  createdAt: number;
  status?: OfflineMutationStatus;
  updatedAt?: number;
  lastError?: string;
}

interface PersistedMutationQueue {
  version: number;
  updatedAt: number;
  mutations: OfflineMutationEntry[];
}

interface PersistedItemIDMap {
  version: number;
  updatedAt: number;
  mappings: Record<string, string>;
}

function openDatabase(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = window.indexedDB.open(DB_NAME, 1);
    request.onerror = () => reject(request.error);
    request.onupgradeneeded = () => {
      const database = request.result;
      if (!database.objectStoreNames.contains(STORE_NAME)) {
        database.createObjectStore(STORE_NAME);
      }
    };
    request.onsuccess = () => resolve(request.result);
  });
}

async function readStoreValue<T>(key: string): Promise<T | null> {
  const database = await openDatabase();
  return new Promise((resolve, reject) => {
    const transaction = database.transaction(STORE_NAME, "readonly");
    const store = transaction.objectStore(STORE_NAME);
    const request = store.get(key);
    request.onerror = () => reject(request.error);
    request.onsuccess = () =>
      resolve((request.result as T | undefined) ?? null);
    transaction.oncomplete = () => database.close();
    transaction.onerror = () => reject(transaction.error);
  });
}

async function writeStoreValue<T>(key: string, value: T): Promise<void> {
  const database = await openDatabase();
  return new Promise((resolve, reject) => {
    const transaction = database.transaction(STORE_NAME, "readwrite");
    const store = transaction.objectStore(STORE_NAME);
    const request = store.put(value, key);
    request.onerror = () => reject(request.error);
    transaction.oncomplete = () => {
      database.close();
      resolve();
    };
    transaction.onerror = () => reject(transaction.error);
  });
}

function isOfflineCacheableQuery(query: Query): boolean {
  return (
    query.state.status === "success" &&
    isOfflineCacheableQueryKey(query.queryKey)
  );
}

export function isOfflineCacheableQueryKey(queryKey: QueryKey): boolean {
  const root = Array.isArray(queryKey) ? queryKey[0] : undefined;
  return typeof root === "string" && OFFLINE_QUERY_ROOTS.has(root);
}

export async function loadPersistedQuerySnapshot(): Promise<PersistedQuerySnapshot | null> {
  if (typeof window === "undefined" || !("indexedDB" in window)) {
    return null;
  }
  const snapshot =
    await readStoreValue<PersistedQuerySnapshot>(QUERY_CACHE_KEY);
  if (!snapshot || snapshot.version !== QUERY_CACHE_VERSION) {
    return null;
  }
  return snapshot;
}

export function hydrateOfflineQueryCache(
  queryClient: QueryClient,
  snapshot: PersistedQuerySnapshot | null,
): void {
  if (!snapshot) {
    return;
  }
  hydrate(queryClient, snapshot.state);
}

export async function persistOfflineQueryCache(
  queryClient: QueryClient,
): Promise<number | null> {
  if (typeof window === "undefined" || !("indexedDB" in window)) {
    return null;
  }

  const snapshot: PersistedQuerySnapshot = {
    version: QUERY_CACHE_VERSION,
    updatedAt: Date.now(),
    state: dehydrate(queryClient, {
      shouldDehydrateQuery: isOfflineCacheableQuery,
    }),
  };

  await writeStoreValue(QUERY_CACHE_KEY, snapshot);
  return snapshot.updatedAt;
}

export async function loadPendingOfflineMutations(): Promise<
  OfflineMutationEntry[]
> {
  if (typeof window === "undefined" || !("indexedDB" in window)) {
    return [];
  }

  const persisted =
    await readStoreValue<PersistedMutationQueue>(MUTATION_QUEUE_KEY);
  if (!persisted || persisted.version !== MUTATION_QUEUE_VERSION) {
    return [];
  }
  return persisted.mutations;
}

export async function enqueueOfflineMutation(
  entry: OfflineMutationEntry,
): Promise<void> {
  if (typeof window === "undefined" || !("indexedDB" in window)) {
    return;
  }

  const existing = await loadPendingOfflineMutations();
  const queue: PersistedMutationQueue = {
    version: MUTATION_QUEUE_VERSION,
    updatedAt: Date.now(),
    mutations: [...existing, { ...entry, status: entry.status ?? "pending" }],
  };
  await writeStoreValue(MUTATION_QUEUE_KEY, queue);
}

export async function markOfflineMutationFailed(
  id: string,
  message: string,
): Promise<void> {
  if (typeof window === "undefined" || !("indexedDB" in window)) {
    return;
  }

  const existing = await loadPendingOfflineMutations();
  const now = Date.now();
  const queue: PersistedMutationQueue = {
    version: MUTATION_QUEUE_VERSION,
    updatedAt: now,
    mutations: existing.map((entry) => {
      if (entry.id !== id) {
        return entry;
      }
      return {
        ...entry,
        status: "failed",
        updatedAt: now,
        lastError: message,
      };
    }),
  };
  await writeStoreValue(MUTATION_QUEUE_KEY, queue);
}

export async function removeOfflineMutation(id: string): Promise<void> {
  if (typeof window === "undefined" || !("indexedDB" in window)) {
    return;
  }

  const existing = await loadPendingOfflineMutations();
  const queue: PersistedMutationQueue = {
    version: MUTATION_QUEUE_VERSION,
    updatedAt: Date.now(),
    mutations: existing.filter((entry) => entry.id !== id),
  };
  await writeStoreValue(MUTATION_QUEUE_KEY, queue);
}

export async function setOfflineItemIDMapping(
  temporaryID: string,
  serverID: string,
): Promise<void> {
  if (typeof window === "undefined" || !("indexedDB" in window)) {
    return;
  }

  const existing = await readStoreValue<PersistedItemIDMap>(ITEM_ID_MAP_KEY);
  const queue: PersistedItemIDMap = {
    version: 1,
    updatedAt: Date.now(),
    mappings: {
      ...(existing?.mappings ?? {}),
      [temporaryID]: serverID,
    },
  };
  await writeStoreValue(ITEM_ID_MAP_KEY, queue);
}

export async function resolveOfflineItemID(id: string): Promise<string> {
  if (typeof window === "undefined" || !("indexedDB" in window)) {
    return id;
  }

  const existing = await readStoreValue<PersistedItemIDMap>(ITEM_ID_MAP_KEY);
  if (!existing || existing.version !== 1) {
    return id;
  }
  return existing.mappings[id] ?? id;
}
