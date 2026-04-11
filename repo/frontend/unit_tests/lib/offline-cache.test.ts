import { describe, it, expect } from "vitest";
import { isOfflineCacheableQueryKey } from "@/lib/offline-cache";

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
