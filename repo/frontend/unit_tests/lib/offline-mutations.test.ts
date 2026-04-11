import { beforeEach, describe, expect, it, vi } from "vitest";
import { QueryClient } from "@tanstack/react-query";
import { replayOfflineMutations } from "@/lib/offline-mutations";

const mockPost = vi.fn();
const mockPut = vi.fn();
const mockLoadPending = vi.fn();
const mockRemove = vi.fn();
const mockMarkFailed = vi.fn();
const mockSetIDMap = vi.fn();

vi.mock("@/lib/api-client", () => ({
  apiClient: {
    post: (...args: unknown[]) => mockPost(...args),
    put: (...args: unknown[]) => mockPut(...args),
  },
  isOfflineApiError: (error: unknown) => {
    const candidate = error as { status?: number; code?: string };
    return (
      candidate?.status === 0 &&
      (candidate?.code === "NETWORK_OFFLINE" ||
        candidate?.code === "NETWORK_ERROR")
    );
  },
}));

vi.mock("@/lib/offline-cache", () => ({
  loadPendingOfflineMutations: (...args: unknown[]) => mockLoadPending(...args),
  removeOfflineMutation: (...args: unknown[]) => mockRemove(...args),
  markOfflineMutationFailed: (...args: unknown[]) => mockMarkFailed(...args),
  setOfflineItemIDMapping: (...args: unknown[]) => mockSetIDMap(...args),
}));

describe("replayOfflineMutations", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(window.navigator, "onLine", {
      configurable: true,
      value: true,
    });
  });

  it("processes pending create-item entries and stores temp->server ID mapping", async () => {
    mockLoadPending.mockResolvedValue([
      {
        id: "q1",
        type: "create-item",
        createdAt: 1,
        status: "pending",
        payload: {
          name: "Kettlebell",
          __offline_temp_id: "temp-item-1",
        },
      },
    ]);
    mockPost.mockResolvedValue({ data: { id: "server-item-1" } });

    const queryClient = new QueryClient();
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

    const processed = await replayOfflineMutations(queryClient);

    expect(processed).toBe(1);
    expect(mockPost).toHaveBeenCalledWith("/items", { name: "Kettlebell" });
    expect(mockSetIDMap).toHaveBeenCalledWith("temp-item-1", "server-item-1");
    expect(mockRemove).toHaveBeenCalledWith("q1");
    expect(mockMarkFailed).not.toHaveBeenCalled();
    expect(invalidateSpy).toHaveBeenCalled();
  });

  it("marks non-offline failures as failed and keeps the queue entry", async () => {
    mockLoadPending.mockResolvedValue([
      {
        id: "q2",
        type: "run-export",
        createdAt: 2,
        status: "pending",
        payload: { report_id: "r1", format: "csv" },
      },
    ]);
    mockPost.mockRejectedValue({
      status: 422,
      code: "VALIDATION_ERROR",
      message: "invalid filters",
    });

    const queryClient = new QueryClient();
    const processed = await replayOfflineMutations(queryClient);

    expect(processed).toBe(0);
    expect(mockMarkFailed).toHaveBeenCalledWith("q2", "invalid filters");
    expect(mockRemove).not.toHaveBeenCalled();
  });

  it("stops replay on offline errors and leaves current entry untouched", async () => {
    mockLoadPending.mockResolvedValue([
      {
        id: "q3",
        type: "run-export",
        createdAt: 3,
        status: "pending",
        payload: { report_id: "r2", format: "pdf" },
      },
      {
        id: "q4",
        type: "update-item",
        createdAt: 4,
        status: "pending",
        payload: { id: "item-1", body: { name: "Updated" } },
      },
    ]);
    mockPost.mockRejectedValue({
      status: 0,
      code: "NETWORK_OFFLINE",
      message: "offline",
    });

    const queryClient = new QueryClient();
    const processed = await replayOfflineMutations(queryClient);

    expect(processed).toBe(0);
    expect(mockRemove).not.toHaveBeenCalled();
    expect(mockMarkFailed).not.toHaveBeenCalled();
    expect(mockPut).not.toHaveBeenCalled();
  });

  it.each([
    [
      "refund-order",
      { id: "o1" },
      "/orders/o1/refund",
      {},
    ],
    [
      "add-order-note",
      { id: "o1", note: "hello" },
      "/orders/o1/notes",
      { note: "hello" },
    ],
    [
      "split-order",
      { id: "o1", quantities: [1, 2] },
      "/orders/o1/split",
      { quantities: [1, 2] },
    ],
    [
      "merge-order",
      { order_ids: ["o1", "o2"] },
      "/orders/merge",
      { order_ids: ["o1", "o2"] },
    ],
    [
      "evaluate-campaign",
      { id: "c1" },
      "/campaigns/c1/evaluate",
      {},
    ],
    [
      "create-campaign",
      { item_id: "i1", min_quantity: 5, cutoff_time: "2026-06-01T00:00:00Z" },
      "/campaigns",
      { item_id: "i1", min_quantity: 5, cutoff_time: "2026-06-01T00:00:00Z" },
    ],
    [
      "create-adjustment",
      { item_id: "i1", quantity_change: -2, reason: "damaged" },
      "/inventory/adjustments",
      { item_id: "i1", quantity_change: -2, reason: "damaged" },
    ],
    [
      "return-po",
      { id: "po1" },
      "/purchase-orders/po1/return",
      {},
    ],
    [
      "resolve-variance",
      { id: "v1", action: "adjustment", resolution_notes: "ok", quantity_change: 1 },
      "/variances/v1/resolve",
      { action: "adjustment", resolution_notes: "ok", quantity_change: 1 },
    ],
  ])(
    "replays %s mutation with correct path and body",
    async (type, payload, expectedPath, expectedBody) => {
      mockLoadPending.mockResolvedValue([
        { id: "q1", type, payload, createdAt: 1, status: "pending" },
      ]);
      mockPost.mockResolvedValue({});

      const queryClient = new QueryClient();
      const count = await replayOfflineMutations(queryClient);

      expect(count).toBe(1);
      expect(mockPost).toHaveBeenCalledWith(expectedPath, expectedBody);
      expect(mockRemove).toHaveBeenCalledWith("q1");
      expect(mockMarkFailed).not.toHaveBeenCalled();
    },
  );

  it("skips entries already marked as failed", async () => {
    mockLoadPending.mockResolvedValue([
      {
        id: "q5",
        type: "run-export",
        createdAt: 5,
        status: "failed",
        payload: { report_id: "r3", format: "csv" },
      },
      {
        id: "q6",
        type: "update-item",
        createdAt: 6,
        status: "pending",
        payload: { id: "item-2", body: { brand: "Reconciled" } },
      },
    ]);
    mockPut.mockResolvedValue({ data: { id: "item-2" } });

    const queryClient = new QueryClient();
    const processed = await replayOfflineMutations(queryClient);

    expect(processed).toBe(1);
    expect(mockPost).not.toHaveBeenCalled();
    expect(mockPut).toHaveBeenCalledWith("/items/item-2", {
      brand: "Reconciled",
    });
    expect(mockRemove).toHaveBeenCalledWith("q6");
  });
});
