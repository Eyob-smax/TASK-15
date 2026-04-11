import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import OrderDetailPage from "@/routes/OrderDetailPage";

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
    }: {
      children: React.ReactNode;
      roles: string[];
    }) => {
      const user = mockUseAuth();
      return roles.includes(user?.user?.role) ? <>{children}</> : null;
    },
  };
});

vi.mock("@/lib/offline", () => ({
  useOfflineStatus: () => ({ isOnline: true, isOffline: false, lastSyncAt: null }),
  OFFLINE_MUTATION_MESSAGE: "Reconnect to make changes.",
}));

const mockUseOrder = vi.fn();
const mockUseOrderTimeline = vi.fn();
const mockCancelOrder = vi.fn();
const mockRefundOrder = vi.fn();
const mockPayOrder = vi.fn();
const mockAddOrderNote = vi.fn();
const mockSplitOrder = vi.fn();

vi.mock("@/lib/hooks/useOrders", () => ({
  useOrder: (id: string) => mockUseOrder(id),
  useOrderTimeline: (id: string) => mockUseOrderTimeline(id),
  useCancelOrder: () => ({ mutateAsync: mockCancelOrder, isPending: false }),
  useRefundOrder: () => ({ mutateAsync: mockRefundOrder, isPending: false }),
  usePayOrder: () => ({ mutateAsync: mockPayOrder, isPending: false }),
  useAddOrderNote: () => ({ mutateAsync: mockAddOrderNote, isPending: false }),
  useSplitOrder: () => ({ mutateAsync: mockSplitOrder, isPending: false }),
}));

vi.mock("@/lib/notifications", () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn() }),
}));

const ORDER_ID = "cccccccc-cccc-cccc-cccc-cccccccccccc";

const MOCK_ORDER = {
  id: ORDER_ID,
  user_id: "user-1",
  item_id: "item-1111-1111-1111-111111111111",
  campaign_id: null,
  quantity: 2,
  unit_price: 50,
  total_amount: 100,
  status: "created",
  settlement_marker: "",
  notes: "",
  auto_close_at: null,
  paid_at: null,
  cancelled_at: null,
  refunded_at: null,
  created_at: "2026-04-09T10:00:00Z",
  updated_at: "2026-04-09T10:00:00Z",
};

const MOCK_TIMELINE = [
  {
    id: "tl-1",
    order_id: ORDER_ID,
    action: "created",
    performed_by: "user-1",
    description: "Order placed",
    created_at: "2026-04-09T10:00:00Z",
  },
];

function renderOrderDetail(role = "administrator") {
  mockUseAuth.mockReturnValue({
    user: { id: "user-1", role, display_name: "Test User" },
    isAuthenticated: true,
  });
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[`/orders/${ORDER_ID}`]}>
        <Routes>
          <Route path="/orders/:id" element={<OrderDetailPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("OrderDetailPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseOrder.mockReturnValue({
      data: MOCK_ORDER,
      isLoading: false,
      error: null,
    });
    mockUseOrderTimeline.mockReturnValue({
      data: MOCK_TIMELINE,
      isLoading: false,
    });
  });

  it("renders order status chip", () => {
    renderOrderDetail();
    expect(screen.getAllByText(/created/i).length).toBeGreaterThan(0);
  });

  it("renders order quantity and total amount", () => {
    renderOrderDetail();
    expect(screen.getByText("$100.00")).toBeInTheDocument();
  });

  it("renders timeline entry description", () => {
    renderOrderDetail();
    expect(screen.getByText(/order placed/i)).toBeInTheDocument();
  });

  it("shows Cancel Order button for member on own Created order", () => {
    mockUseAuth.mockReturnValue({
      user: { id: "user-1", role: "member" },
      isAuthenticated: true,
    });
    renderOrderDetail("member");
    expect(
      screen.getByRole("button", { name: /cancel order/i }),
    ).toBeInTheDocument();
  });

  it("hides Cancel Order button for member on own Paid order", () => {
    mockUseAuth.mockReturnValue({
      user: { id: "user-1", role: "member" },
      isAuthenticated: true,
    });
    mockUseOrder.mockReturnValue({
      data: { ...MOCK_ORDER, status: "paid" },
      isLoading: false,
      error: null,
    });
    renderOrderDetail("member");
    expect(
      screen.queryByRole("button", { name: /cancel order/i }),
    ).not.toBeInTheDocument();
  });

  it("shows Record Payment and Add Note actions for admin", () => {
    renderOrderDetail("administrator");
    expect(
      screen.getByRole("button", { name: /record payment/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /add note/i }),
    ).toBeInTheDocument();
  });

  it("shows error alert when order fails to load", () => {
    mockUseOrder.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error("fail"),
    });
    renderOrderDetail();
    expect(
      screen.getByText(/failed to load order details/i),
    ).toBeInTheDocument();
  });

  it("shows loading skeleton when order is loading", () => {
    mockUseOrder.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
    });
    renderOrderDetail();
    expect(screen.queryByText(/order details/i)).toBeInTheDocument();
    expect(screen.queryByText("$100.00")).not.toBeInTheDocument();
  });
});
