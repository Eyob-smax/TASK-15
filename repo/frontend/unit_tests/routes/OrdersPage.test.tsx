import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, within, waitFor, cleanup } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import OrdersPage from '@/routes/OrdersPage';

const mockUseAuth = vi.fn();
vi.mock('@/lib/auth', async () => {
  const actual = await vi.importActual<typeof import('@/lib/auth')>('@/lib/auth');
  return {
    ...actual,
    useAuth: () => mockUseAuth(),
  };
});

vi.mock('@/lib/offline', () => ({
  useOfflineStatus: () => ({ isOnline: true, isOffline: false, lastSyncAt: null }),
}));

const mockUseOrderList = vi.fn();
const mockMergeMutateAsync = vi.fn();
vi.mock('@/lib/hooks/useOrders', () => ({
  useOrderList: (params: unknown) => mockUseOrderList(params),
  useMergeOrder: () => ({ mutateAsync: mockMergeMutateAsync, isPending: false }),
}));

vi.mock('@/lib/notifications', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn(), info: vi.fn() }),
}));

const MOCK_ORDER_CREATED = {
  id: 'oooooooo-oooo-oooo-oooo-oooooooooooo',
  member_id: 'm1',
  item_id: 'iiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii',
  quantity: 2,
  total_amount: 150.5,
  status: 'created',
  created_at: '2026-04-01T00:00:00Z',
};

const MOCK_ORDER_CANCELLED = {
  ...MOCK_ORDER_CREATED,
  id: 'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
  status: 'cancelled',
};

function renderPage(role = 'administrator') {
  mockUseAuth.mockReturnValue({
    user: { id: 'u1', role, display_name: 'Test User', email: 't@t.com' },
    isAuthenticated: true,
    isLoading: false,
  });
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <OrdersPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('OrdersPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseOrderList.mockReturnValue({
      data: {
        data: [MOCK_ORDER_CREATED, MOCK_ORDER_CANCELLED],
        pagination: { page: 1, page_size: 20, total_count: 2, total_pages: 1 },
      },
      isLoading: false,
      error: null,
      dataUpdatedAt: Date.now(),
    });
  });

  it('renders page title', () => {
    renderPage();
    expect(screen.getByRole('heading', { name: /orders/i })).toBeInTheDocument();
  });

  it('renders order rows with total and short ids', () => {
    renderPage();
    expect(screen.getAllByText(/oooooooo/).length).toBeGreaterThan(0);
    // Both mock orders share total_amount 150.5 — assert at least one match.
    expect(screen.getAllByText('$150.50').length).toBeGreaterThan(0);
  });

  it('shows error alert when query fails', () => {
    mockUseOrderList.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('boom'),
      dataUpdatedAt: 0,
    });
    renderPage();
    expect(screen.getByText(/failed to load orders/i)).toBeInTheDocument();
  });

  it('admin sees merge controls', () => {
    renderPage('administrator');
    expect(screen.getByRole('button', { name: /select page/i })).toBeInTheDocument();
  });

  it('member does not see merge controls', () => {
    cleanup();
    renderPage('member');
    expect(screen.queryByRole('button', { name: /select page/i })).not.toBeInTheDocument();
  });

  it('merge button is disabled until at least two orders are selected', async () => {
    renderPage('administrator');
    const user = userEvent.setup();

    const mergeBtn = screen.getByRole('button', { name: /merge selected/i });
    expect(mergeBtn).toBeDisabled();

    // Selecting the one eligible row (created) leaves count at 1 — still disabled.
    const checkboxes = screen.getAllByRole('checkbox');
    await user.click(checkboxes[0]);
    expect(screen.getByRole('button', { name: /merge selected \(1\)/i })).toBeDisabled();
  });

  it('Select Page toggles all eligible rows and clears with Clear Selection', async () => {
    renderPage('administrator');
    const user = userEvent.setup();

    await user.click(screen.getByRole('button', { name: /^select page$/i }));
    // After selecting, one eligible order is checked — counter should show 1
    expect(screen.getByRole('button', { name: /merge selected \(1\)/i })).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: /clear selection/i }));
    expect(screen.getByRole('button', { name: /merge selected \(0\)/i })).toBeInTheDocument();
  });

  it('opens the merge dialog when two selectable orders are present', async () => {
    mockUseOrderList.mockReturnValue({
      data: {
        data: [
          MOCK_ORDER_CREATED,
          { ...MOCK_ORDER_CREATED, id: 'yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy' },
        ],
        pagination: { page: 1, page_size: 20, total_count: 2, total_pages: 1 },
      },
      isLoading: false,
      error: null,
      dataUpdatedAt: Date.now(),
    });
    renderPage('administrator');
    const user = userEvent.setup();

    await user.click(screen.getByRole('button', { name: /^select page$/i }));
    await user.click(screen.getByRole('button', { name: /merge selected \(2\)/i }));

    const dialog = await screen.findByRole('dialog');
    expect(within(dialog).getByText(/merge orders/i)).toBeInTheDocument();
  });

  it('calls mergeMutation when confirming the merge dialog', async () => {
    mockUseOrderList.mockReturnValue({
      data: {
        data: [
          MOCK_ORDER_CREATED,
          { ...MOCK_ORDER_CREATED, id: 'yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy' },
        ],
        pagination: { page: 1, page_size: 20, total_count: 2, total_pages: 1 },
      },
      isLoading: false,
      error: null,
      dataUpdatedAt: Date.now(),
    });
    mockMergeMutateAsync.mockResolvedValue({ id: 'merged-id' });
    renderPage('administrator');
    const user = userEvent.setup();

    await user.click(screen.getByRole('button', { name: /^select page$/i }));
    await user.click(screen.getByRole('button', { name: /merge selected \(2\)/i }));

    const dialog = await screen.findByRole('dialog');
    const confirmBtn = within(dialog).getByRole('button', { name: /^merge$/i });
    await user.click(confirmBtn);

    await waitFor(() => expect(mockMergeMutateAsync).toHaveBeenCalled());
  });
});
