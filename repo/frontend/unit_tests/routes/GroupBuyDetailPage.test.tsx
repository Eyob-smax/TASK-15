import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import GroupBuyDetailPage from "@/routes/GroupBuyDetailPage";

const mockUseAuth = vi.fn();
vi.mock("@/lib/auth", async () => {
  const actual =
    await vi.importActual<typeof import("@/lib/auth")>("@/lib/auth");
  return {
    ...actual,
    useAuth: () => mockUseAuth(),
    RequireRole: ({
      children,
      roles,
      fallback,
    }: {
      children: React.ReactNode;
      roles: string[];
      fallback?: React.ReactNode;
    }) => {
      const u = mockUseAuth();
      return roles.includes(u?.user?.role) ? (
        <>{children}</>
      ) : (
        <>{fallback ?? null}</>
      );
    },
  };
});

const mockUseCampaign = vi.fn();
const mockUseItem = vi.fn();
const mockJoinCampaign = vi.fn();
const mockCancelCampaign = vi.fn();
const mockEvaluateCampaign = vi.fn();

vi.mock("@/lib/hooks/useCampaigns", () => ({
  useCampaign: (id: string) => mockUseCampaign(id),
  useJoinCampaign: () => ({ mutateAsync: mockJoinCampaign, isPending: false }),
  useCancelCampaign: () => ({
    mutateAsync: mockCancelCampaign,
    isPending: false,
  }),
  useEvaluateCampaign: () => ({
    mutateAsync: mockEvaluateCampaign,
    isPending: false,
  }),
}));

vi.mock("@/lib/hooks/useItems", () => ({
  useItem: () => mockUseItem(),
}));

vi.mock("@/lib/notifications", () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn() }),
}));

vi.mock("@/lib/offline", () => ({
  useOfflineStatus: () => ({
    isOnline: true,
    isOffline: false,
    lastSyncAt: null,
  }),
}));

const CAMPAIGN_ID = "dddddddd-dddd-dddd-dddd-dddddddddddd";

const MOCK_CAMPAIGN = {
  id: CAMPAIGN_ID,
  item_id: "item-1111-1111-1111-111111111111",
  min_quantity: 10,
  current_committed_qty: 6,
  status: "active",
  cutoff_time: "2026-05-01T00:00:00Z",
  created_by: "user-admin",
  created_at: "2026-04-01T00:00:00Z",
  evaluated_at: null,
};

const MOCK_ITEM = {
  id: MOCK_CAMPAIGN.item_id,
  name: "Spring Kettlebell",
  category: "strength",
  refundable_deposit: 50,
};

function renderGroupBuyDetail(role = "member") {
  mockUseAuth.mockReturnValue({
    user: { id: "user-1", role },
    isAuthenticated: true,
  });
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[`/group-buys/${CAMPAIGN_ID}`]}>
        <Routes>
          <Route path="/group-buys/:id" element={<GroupBuyDetailPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("GroupBuyDetailPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseCampaign.mockReturnValue({
      data: MOCK_CAMPAIGN,
      isLoading: false,
      error: null,
    });
    mockUseItem.mockReturnValue({
      data: MOCK_ITEM,
      isLoading: false,
      error: null,
    });
  });

  it("renders linked item name in the campaign title", () => {
    renderGroupBuyDetail();
    expect(
      screen.getAllByText("Spring Kettlebell Group Buy").length,
    ).toBeGreaterThan(0);
  });

  it("renders campaign status chip", () => {
    renderGroupBuyDetail();
    expect(screen.getByText(/active/i)).toBeInTheDocument();
  });

  it("renders progress information", () => {
    renderGroupBuyDetail();
    expect(screen.getByText(/6\s*\/\s*10 min/i)).toBeInTheDocument();
  });

  it("renders linked item details", () => {
    renderGroupBuyDetail();
    expect(screen.getByText("Spring Kettlebell")).toBeInTheDocument();
    expect(screen.getByText("strength")).toBeInTheDocument();
    expect(screen.getByText("$50.00")).toBeInTheDocument();
  });

  it("shows join form for members", () => {
    renderGroupBuyDetail("member");
    expect(screen.getByText(/join this campaign/i)).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /join campaign/i }),
    ).toBeInTheDocument();
  });

  it("hides join form for admin (not a member)", () => {
    renderGroupBuyDetail("administrator");
    expect(
      screen.queryByRole("button", { name: /join campaign/i }),
    ).not.toBeInTheDocument();
    expect(screen.getByText(/only members can join/i)).toBeInTheDocument();
  });

  it("shows Evaluate and Cancel buttons for admin on active campaign", () => {
    renderGroupBuyDetail("administrator");
    expect(
      screen.getByRole("button", { name: /evaluate/i }),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /cancel/i })).toBeInTheDocument();
  });

  it("hides staff actions for member", () => {
    renderGroupBuyDetail("member");
    expect(
      screen.queryByRole("button", { name: /evaluate/i }),
    ).not.toBeInTheDocument();
  });

  it("shows error alert when campaign fails to load", () => {
    mockUseCampaign.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error("fail"),
    });
    renderGroupBuyDetail();
    expect(
      screen.getByText(/failed to load campaign details/i),
    ).toBeInTheDocument();
  });
});
