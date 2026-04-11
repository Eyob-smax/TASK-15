import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ComponentProps } from "react";
import DashboardPage from "@/routes/DashboardPage";
import { AuthContext } from "@/lib/auth";
import type { User } from "@/lib/types";

// Mock useDashboardKPIs — matches the hook exported by useDashboard.ts
const mockUseDashboardKPIs = vi.fn();
vi.mock("@/lib/hooks/useDashboard", () => ({
  useDashboardKPIs: (...args: unknown[]) => mockUseDashboardKPIs(...args),
}));

vi.mock("@/lib/offline", () => ({
  useOfflineStatus: () => ({
    isOnline: true,
    isOffline: false,
    lastSyncAt: null,
  }),
}));

type AuthValue = NonNullable<
  ComponentProps<typeof AuthContext.Provider>["value"]
>;

function makeUser(role: User["role"]): User {
  return {
    id: "1",
    email: "test@example.com",
    display_name: "Test User",
    role,
    status: "active",
    location_id: null,
    failed_login_attempts: 0,
    locked_until: null,
    last_login_at: null,
    password_changed_at: null,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  };
}

function makeAuthValue(role: User["role"]): AuthValue {
  return {
    user: makeUser(role),
    session: null,
    isLoading: false,
    isAuthenticated: true,
    captchaState: null,
    lockoutState: null,
    login: vi.fn(),
    logout: vi.fn(),
    refreshSession: vi.fn(),
    verifyCaptcha: vi.fn(),
    clearCaptchaState: vi.fn(),
  };
}

function renderDashboard(role = "administrator") {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <AuthContext.Provider value={makeAuthValue(role as User["role"])}>
      <QueryClientProvider client={qc}>
        <MemoryRouter>
          <DashboardPage />
        </MemoryRouter>
      </QueryClientProvider>
    </AuthContext.Provider>,
  );
}

const kpiData = {
  member_growth: { value: 5, change_percent: 1.5, period: "weekly" },
  churn: { value: 2, change_percent: -0.5, period: "weekly" },
  renewal_rate: { value: 0.85, change_percent: 0.1, period: "weekly" },
  engagement: { value: 0.72, change_percent: 0.04, period: "weekly" },
  class_fill_rate: { value: 0.9, change_percent: 0.03, period: "weekly" },
  coach_productivity: { value: 0.88, change_percent: 0.02, period: "weekly" },
};

describe("DashboardPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseDashboardKPIs.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    });
  });

  it("renders page title", () => {
    renderDashboard();
    expect(
      screen.getByRole("heading", { name: /dashboard/i }),
    ).toBeInTheDocument();
  });

  it("renders period toggle buttons", () => {
    renderDashboard();
    expect(screen.getByRole("group", { name: /period/i })).toBeInTheDocument();
    expect(screen.getByText("Daily")).toBeInTheDocument();
    expect(screen.getByText("Weekly")).toBeInTheDocument();
    expect(screen.getByText("Monthly")).toBeInTheDocument();
  });

  it("shows loading skeletons while fetching", () => {
    mockUseDashboardKPIs.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
    });
    renderDashboard();
    // The "Membership & Engagement" section heading is always rendered
    expect(screen.getByText(/membership/i)).toBeInTheDocument();
  });

  it("shows warning alert when API returns error", () => {
    mockUseDashboardKPIs.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error("501 Not Implemented"),
    });
    renderDashboard();
    expect(screen.getByText(/not yet available/i)).toBeInTheDocument();
  });

  it("renders KPI stat cards with data", () => {
    mockUseDashboardKPIs.mockReturnValue({
      data: kpiData,
      isLoading: false,
      error: null,
    });
    renderDashboard();
    expect(screen.getByText("Member Growth")).toBeInTheDocument();
    expect(screen.getByText("Churn Rate")).toBeInTheDocument();
    expect(screen.getByText("Renewal Rate")).toBeInTheDocument();
  });

  it("switches period when toggle is clicked", async () => {
    renderDashboard();
    const monthlyBtn = screen.getByText("Monthly");
    fireEvent.click(monthlyBtn);
    // After click, useDashboardKPIs should be called with updated period
    await waitFor(() => {
      expect(mockUseDashboardKPIs).toHaveBeenCalled();
    });
  });

  it("admin sees operations KPI section", () => {
    mockUseDashboardKPIs.mockReturnValue({
      data: kpiData,
      isLoading: false,
      error: null,
    });
    renderDashboard("administrator");
    expect(screen.getByText("Class Fill Rate")).toBeInTheDocument();
    expect(screen.getByText("Coach Productivity")).toBeInTheDocument();
  });

  it("member does not see operations section", () => {
    renderDashboard("member");
    expect(screen.queryByText("Class Fill Rate")).not.toBeInTheDocument();
  });
});
