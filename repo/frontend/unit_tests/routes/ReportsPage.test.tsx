import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import ReportsPage from "@/routes/ReportsPage";

const mockUseAuth = vi.fn();
vi.mock("@/lib/auth", async () => {
  const actual =
    await vi.importActual<typeof import("@/lib/auth")>("@/lib/auth");
  return {
    ...actual,
    useAuth: () => mockUseAuth(),
  };
});

vi.mock("@/lib/offline", () => ({
  useOfflineStatus: () => ({
    isOnline: true,
    isOffline: false,
    lastSyncAt: null,
  }),
  OFFLINE_MUTATION_MESSAGE: "Reconnect to make changes.",
}));

const mockUseReportList = vi.fn();
const mockRunExportMutateAsync = vi.fn();
const mockDownloadFile = vi.fn();

vi.mock("@/lib/hooks/useReports", () => ({
  useReportList: () => mockUseReportList(),
  useRunExport: () => ({
    mutateAsync: mockRunExportMutateAsync,
    isPending: false,
  }),
}));

vi.mock("@/lib/api-client", async () => {
  const actual =
    await vi.importActual<typeof import("@/lib/api-client")>(
      "@/lib/api-client",
    );
  return {
    ...actual,
    downloadFile: (...args: Parameters<typeof actual.downloadFile>) =>
      mockDownloadFile(...args),
  };
});

vi.mock("@/lib/notifications", () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn() }),
}));

const MOCK_REPORT = {
  id: "rrrrrrrr-rrrr-rrrr-rrrr-rrrrrrrrrrrr",
  name: "Items Summary",
  description: "Catalog with stock counts",
  report_type: "items_summary",
  allowed_roles: ["administrator", "operations_manager"],
  created_at: "2026-04-01T00:00:00Z",
};

const MOCK_EXPORT_JOB = {
  id: "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee",
  report_id: MOCK_REPORT.id,
  format: "csv",
  filename: "items_summary_20260401_120000.csv",
  status: "completed",
  created_by: "user-1",
  created_at: "2026-04-01T12:00:00Z",
  completed_at: "2026-04-01T12:00:05Z",
};

function renderPage(role = "administrator") {
  mockUseAuth.mockReturnValue({
    user: {
      id: "user-1",
      role,
      display_name: "Test User",
      email: "test@test.com",
    },
    isAuthenticated: true,
    isLoading: false,
  });
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <ReportsPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("ReportsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseReportList.mockReturnValue({
      data: [MOCK_REPORT],
      isLoading: false,
      error: null,
    });
    mockRunExportMutateAsync.mockResolvedValue(MOCK_EXPORT_JOB);
    mockDownloadFile.mockResolvedValue(MOCK_EXPORT_JOB.filename);
  });

  it("renders page title", () => {
    renderPage();
    expect(
      screen.getByRole("heading", { name: /reports/i }),
    ).toBeInTheDocument();
  });

  it("renders report card with name and type", () => {
    renderPage();
    expect(screen.getByText("Items Summary")).toBeInTheDocument();
    expect(screen.getByText("items_summary")).toBeInTheDocument();
  });

  it("renders Export CSV and Export PDF buttons per report", () => {
    renderPage();
    expect(
      screen.getByRole("button", { name: /export csv/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /export pdf/i }),
    ).toBeInTheDocument();
  });

  it("clicking Export CSV calls useRunExport with csv format and downloads the file", async () => {
    renderPage();
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: /export csv/i }));

    await waitFor(() => {
      expect(mockRunExportMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          report_id: MOCK_REPORT.id,
          format: "csv",
          parameters: {},
        }),
      );
    });

    await waitFor(() => {
      expect(mockDownloadFile).toHaveBeenCalledWith(
        `/exports/${MOCK_EXPORT_JOB.id}/download`,
        MOCK_EXPORT_JOB.filename,
      );
    });
  });

  it("forwards selected filters with export requests", async () => {
    renderPage();
    const user = userEvent.setup();

    fireEvent.change(screen.getByLabelText(/location id/i), {
      target: { value: "loc-1" },
    });
    fireEvent.change(screen.getByLabelText(/category/i), {
      target: { value: "supplements" },
    });
    fireEvent.change(screen.getByLabelText(/^from$/i), {
      target: { value: "2026-04-01" },
    });
    await user.click(screen.getByRole("button", { name: /export csv/i }));

    await waitFor(() => {
      expect(mockRunExportMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          report_id: MOCK_REPORT.id,
          format: "csv",
          parameters: expect.objectContaining({
            location_id: "loc-1",
            category: "supplements",
            from: "2026-04-01",
          }),
        }),
      );
    });
  });

  it("clicking Export PDF calls useRunExport with pdf format", async () => {
    mockRunExportMutateAsync.mockResolvedValueOnce({
      ...MOCK_EXPORT_JOB,
      id: "ffffffff-ffff-ffff-ffff-ffffffffffff",
      format: "pdf",
      filename: "items_summary_20260401_120000.pdf",
    });
    renderPage();
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: /export pdf/i }));

    await waitFor(() => {
      expect(mockRunExportMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          report_id: MOCK_REPORT.id,
          format: "pdf",
          parameters: {},
        }),
      );
    });
  });

  it("shows completed exports in the recent jobs list with a download action", async () => {
    renderPage();
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: /export csv/i }));

    expect(
      await screen.findByText(MOCK_EXPORT_JOB.filename),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /download/i }),
    ).toBeInTheDocument();
  });

  it("shows loading text while reports are loading", () => {
    mockUseReportList.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
    });
    renderPage();
    expect(screen.getByText(/loading reports/i)).toBeInTheDocument();
  });

  it("shows no reports message when list is empty", () => {
    mockUseReportList.mockReturnValue({
      data: [],
      isLoading: false,
      error: null,
    });
    renderPage();
    expect(screen.getByText(/no reports are available/i)).toBeInTheDocument();
  });

  it("shows error alert when reports fail to load", () => {
    mockUseReportList.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error("server error"),
    });
    renderPage();
    expect(screen.getByRole("alert")).toBeInTheDocument();
  });
});
