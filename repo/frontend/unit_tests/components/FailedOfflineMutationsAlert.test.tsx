import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { FailedOfflineMutationsAlert } from "@/components/FailedOfflineMutationsAlert";
import { OfflineStatusProvider } from "@/lib/offline";
import type { ReactNode } from "react";

const mockLoadFailed = vi.fn();
const mockRemove = vi.fn();

vi.mock("@/lib/offline-cache", () => ({
  loadFailedOfflineMutations: (...args: unknown[]) => mockLoadFailed(...args),
  removeOfflineMutation: (...args: unknown[]) => mockRemove(...args),
}));

function Wrapper({ children }: { children: ReactNode }) {
  return (
    <OfflineStatusProvider lastSyncAt={null}>{children}</OfflineStatusProvider>
  );
}

describe("FailedOfflineMutationsAlert", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockRemove.mockResolvedValue(undefined);
  });

  it("renders nothing when there are no mutations", async () => {
    mockLoadFailed.mockResolvedValue([]);

    const { container } = render(
      <Wrapper>
        <FailedOfflineMutationsAlert />
      </Wrapper>,
    );

    await waitFor(() => expect(mockLoadFailed).toHaveBeenCalled());
    expect(container.firstChild).toBeNull();
  });

  it("renders failed entries even when no error message is present", async () => {
    mockLoadFailed.mockResolvedValue([
      {
        id: "q1",
        type: "create-item",
        payload: {},
        createdAt: Date.now(),
        status: "failed",
      },
    ]);

    render(
      <Wrapper>
        <FailedOfflineMutationsAlert />
      </Wrapper>,
    );

    await waitFor(() => expect(mockLoadFailed).toHaveBeenCalled());
    expect(screen.getByText(/create-item/i)).toBeInTheDocument();
  });

  it("renders alert with details when failed entries exist", async () => {
    mockLoadFailed.mockResolvedValue([
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
      </Wrapper>,
    );

    await waitFor(() =>
      expect(
        screen.getByText(/queued action.*failed to sync/i),
      ).toBeInTheDocument(),
    );
    expect(screen.getByText(/run-export/i)).toBeInTheDocument();
    expect(screen.getByText(/server rejected/i)).toBeInTheDocument();
  });

  it("calls removeOfflineMutation and hides the entry when dismissed", async () => {
    mockLoadFailed.mockResolvedValue([
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
      </Wrapper>,
    );

    await waitFor(() =>
      expect(screen.getByText(/dismiss/i)).toBeInTheDocument(),
    );

    fireEvent.click(screen.getByText(/dismiss/i));

    await waitFor(() => expect(mockRemove).toHaveBeenCalledWith("q3"));
    await waitFor(() =>
      expect(screen.queryByText(/update-item/i)).not.toBeInTheDocument(),
    );
  });
});
