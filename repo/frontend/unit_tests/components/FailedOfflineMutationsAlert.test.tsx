import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { FailedOfflineMutationsAlert } from "@/components/FailedOfflineMutationsAlert";
import { OfflineStatusProvider } from "@/lib/offline";
import type { ReactNode } from "react";

const mockLoadPending = vi.fn();
const mockRemove = vi.fn();

vi.mock("@/lib/offline-cache", () => ({
  loadPendingOfflineMutations: (...args: unknown[]) => mockLoadPending(...args),
  removeOfflineMutation: (...args: unknown[]) => mockRemove(...args),
}));

function Wrapper({ children }: { children: ReactNode }) {
  return <OfflineStatusProvider lastSyncAt={null}>{children}</OfflineStatusProvider>;
}

describe("FailedOfflineMutationsAlert", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockRemove.mockResolvedValue(undefined);
  });

  it("renders nothing when there are no mutations", async () => {
    mockLoadPending.mockResolvedValue([]);

    const { container } = render(
      <Wrapper>
        <FailedOfflineMutationsAlert />
      </Wrapper>
    );

    await waitFor(() => expect(mockLoadPending).toHaveBeenCalled());
    expect(container.firstChild).toBeNull();
  });

  it("renders nothing when all mutations are pending (not failed)", async () => {
    mockLoadPending.mockResolvedValue([
      {
        id: "q1",
        type: "create-item",
        payload: {},
        createdAt: Date.now(),
        status: "pending",
      },
    ]);

    const { container } = render(
      <Wrapper>
        <FailedOfflineMutationsAlert />
      </Wrapper>
    );

    await waitFor(() => expect(mockLoadPending).toHaveBeenCalled());
    expect(container.firstChild).toBeNull();
  });

  it("renders alert with details when failed entries exist", async () => {
    mockLoadPending.mockResolvedValue([
      {
        id: "q2",
        type: "run-export",
        payload: {},
        createdAt: Date.now(),
        status: "failed",
        lastError: "server rejected",
      },
    ]);

    render(
      <Wrapper>
        <FailedOfflineMutationsAlert />
      </Wrapper>
    );

    await waitFor(() =>
      expect(screen.getByText(/queued action.*failed to sync/i)).toBeInTheDocument()
    );
    expect(screen.getByText(/run-export/i)).toBeInTheDocument();
    expect(screen.getByText(/server rejected/i)).toBeInTheDocument();
  });

  it("calls removeOfflineMutation and hides the entry when dismissed", async () => {
    mockLoadPending.mockResolvedValue([
      {
        id: "q3",
        type: "update-item",
        payload: {},
        createdAt: Date.now(),
        status: "failed",
        lastError: "404 not found",
      },
    ]);

    render(
      <Wrapper>
        <FailedOfflineMutationsAlert />
      </Wrapper>
    );

    await waitFor(() => expect(screen.getByText(/dismiss/i)).toBeInTheDocument());

    fireEvent.click(screen.getByText(/dismiss/i));

    await waitFor(() => expect(mockRemove).toHaveBeenCalledWith("q3"));
    await waitFor(() => expect(screen.queryByText(/update-item/i)).not.toBeInTheDocument());
  });
});
