import { describe, it, expect, beforeEach } from "vitest";
import { QueryClient } from "@tanstack/react-query";
import {
  isOfflineCacheableQueryKey,
  hydrateOfflineQueryCache,
  loadPersistedQuerySnapshot,
  persistOfflineQueryCache,
  loadPendingOfflineMutations,
  loadFailedOfflineMutations,
  enqueueOfflineMutation,
  markOfflineMutationFailed,
  clearFailedOfflineMutations,
  removeOfflineMutation,
  setOfflineItemIDMapping,
  resolveOfflineItemID,
} from "@/lib/offline-cache";

// ─── Existing operational roots (regression guard) ────────────────────────────

describe("isOfflineCacheableQueryKey — existing operational roots", () => {
  const operationalRoots = [
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
  ] as const;

  for (const root of operationalRoots) {
    it(`caches queries with root "${root}"`, () => {
      expect(isOfflineCacheableQueryKey([root])).toBe(true);
      expect(isOfflineCacheableQueryKey([root, { page: 1 }])).toBe(true);
    });
  }
});

// ─── Admin read domains (offline coverage fix) ────────────────────────────────

describe("isOfflineCacheableQueryKey — admin read domains", () => {
  it('caches "users" queries (useUserList, useUser)', () => {
    expect(isOfflineCacheableQueryKey(["users"])).toBe(true);
    expect(
      isOfflineCacheableQueryKey(["users", { page: 1, role: "coach" }]),
    ).toBe(true);
    expect(isOfflineCacheableQueryKey(["users", "user-uuid-123"])).toBe(true);
  });

  it('caches "audit" queries (useAuditLog, useSecurityEvents)', () => {
    expect(isOfflineCacheableQueryKey(["audit"])).toBe(true);
    expect(isOfflineCacheableQueryKey(["audit", { page: 1 }])).toBe(true);
    expect(isOfflineCacheableQueryKey(["audit", "security", { page: 1 }])).toBe(
      true,
    );
  });

  it('caches "backups" queries (useBackupList)', () => {
    expect(isOfflineCacheableQueryKey(["backups"])).toBe(true);
    expect(isOfflineCacheableQueryKey(["backups", { page: 1 }])).toBe(true);
  });

  it('caches "biometrics" queries (useBiometric)', () => {
    expect(isOfflineCacheableQueryKey(["biometrics"])).toBe(true);
    expect(isOfflineCacheableQueryKey(["biometrics", "user-uuid-456"])).toBe(
      true,
    );
  });

  it('caches "encryption-keys" queries (useEncryptionKeys)', () => {
    expect(isOfflineCacheableQueryKey(["encryption-keys"])).toBe(true);
  });

  it('caches "retention-policies" queries (useRetentionPolicies)', () => {
    expect(isOfflineCacheableQueryKey(["retention-policies"])).toBe(true);
  });
});

// ─── Non-cacheable roots ──────────────────────────────────────────────────────

describe("isOfflineCacheableQueryKey — non-cacheable roots", () => {
  it("rejects unknown roots", () => {
    expect(isOfflineCacheableQueryKey(["unknown-root"])).toBe(false);
    expect(isOfflineCacheableQueryKey(["admin"])).toBe(false);
    expect(isOfflineCacheableQueryKey(["auth"])).toBe(false);
    expect(isOfflineCacheableQueryKey(["session"])).toBe(false);
  });
});

// ─── Edge cases ───────────────────────────────────────────────────────────────

describe("isOfflineCacheableQueryKey — edge cases", () => {
  it("rejects an empty array key", () => {
    expect(isOfflineCacheableQueryKey([])).toBe(false);
  });

  it("rejects a non-array scalar key", () => {
    const scalarKey = "orders" as unknown as readonly unknown[];
    expect(isOfflineCacheableQueryKey(scalarKey)).toBe(false);
  });

  it("rejects an array whose first element is not a string", () => {
    expect(isOfflineCacheableQueryKey([42])).toBe(false);
    expect(isOfflineCacheableQueryKey([null])).toBe(false);
    expect(isOfflineCacheableQueryKey([{ root: "orders" }])).toBe(false);
  });
});

// ─── hydrateOfflineQueryCache ────────────────────────────────────────────────

describe("hydrateOfflineQueryCache", () => {
  it("no-ops when snapshot is null", () => {
    const qc = new QueryClient();
    expect(() => hydrateOfflineQueryCache(qc, null)).not.toThrow();
  });
});

// ─── IDB-backed operations without IndexedDB ─────────────────────────────────
//
// jsdom does not provide window.indexedDB. Every IDB-backed function should
// hit its early-return branch and resolve without error. These tests exercise
// those fallback paths to raise coverage without needing a real polyfill.

describe("IDB-backed operations fall back cleanly when IndexedDB is absent", () => {
  beforeEach(() => {
    // Ensure the early-return branch fires even if some upstream setup added a
    // stub.
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    if ("indexedDB" in window) delete (window as any).indexedDB;
  });

  it("loadPersistedQuerySnapshot returns null", async () => {
    await expect(loadPersistedQuerySnapshot()).resolves.toBeNull();
  });

  it("persistOfflineQueryCache returns null", async () => {
    const qc = new QueryClient();
    await expect(persistOfflineQueryCache(qc)).resolves.toBeNull();
  });

  it("loadPendingOfflineMutations returns empty array", async () => {
    await expect(loadPendingOfflineMutations()).resolves.toEqual([]);
  });

  it("loadFailedOfflineMutations returns empty array", async () => {
    await expect(loadFailedOfflineMutations()).resolves.toEqual([]);
  });

  it("enqueueOfflineMutation resolves silently", async () => {
    await expect(
      enqueueOfflineMutation({
        id: "1",
        type: "create-item",
        payload: { name: "x" },
        createdAt: Date.now(),
      }),
    ).resolves.toBeUndefined();
  });

  it("markOfflineMutationFailed resolves silently", async () => {
    await expect(markOfflineMutationFailed("1", "boom")).resolves.toBeUndefined();
  });

  it("clearFailedOfflineMutations resolves silently", async () => {
    await expect(clearFailedOfflineMutations()).resolves.toBeUndefined();
  });

  it("removeOfflineMutation resolves silently", async () => {
    await expect(removeOfflineMutation("1")).resolves.toBeUndefined();
  });

  it("setOfflineItemIDMapping resolves silently", async () => {
    await expect(setOfflineItemIDMapping("temp", "real")).resolves.toBeUndefined();
  });

  it("resolveOfflineItemID returns the original id", async () => {
    await expect(resolveOfflineItemID("temp-123")).resolves.toBe("temp-123");
  });
});
