import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import InventoryPage from '@/routes/InventoryPage';

const mockUseAuth = vi.fn();
vi.mock('@/lib/auth', async () => {
  const actual = await vi.importActual<typeof import('@/lib/auth')>('@/lib/auth');
  return {
    ...actual,
    useAuth: () => mockUseAuth(),
    RequireRole: ({ children, roles }: { children: React.ReactNode; roles: string[] }) => {
      const user = mockUseAuth();
      return roles.includes(user?.user?.role) ? <>{children}</> : null;
    },
  };
});

vi.mock('@/lib/offline', () => ({
  useOfflineStatus: () => ({ isOnline: true, isOffline: false, lastSyncAt: null }),
  OFFLINE_MUTATION_MESSAGE: 'Reconnect to make changes.',
}));

const mockUseInventorySnapshots = vi.fn();
const mockUseInventoryAdjustments = vi.fn();
const mockCreateAdjustmentMutateAsync = vi.fn();

vi.mock('@/lib/hooks/useInventory', () => ({
  useInventorySnapshots: (params: unknown) => mockUseInventorySnapshots(params),
  useInventoryAdjustments: (params: unknown) => mockUseInventoryAdjustments(params),
  useCreateAdjustment: () => ({ mutateAsync: mockCreateAdjustmentMutateAsync, isPending: false }),
}));

vi.mock('@/lib/notifications', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn() }),
}));

const MOCK_SNAPSHOT = {
  id: 'ssssssss-ssss-ssss-ssss-ssssssssssss',
  item_id: 'iiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii',
  location_id: 'llllllll-llll-llll-llll-llllllllllll',
  quantity_on_hand: 50,
  quantity_reserved: 5,
  quantity_available: 45,
  snapshot_date: '2026-04-01T00:00:00Z',
};

const MOCK_ADJUSTMENT = {
  id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  item_id: 'iiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii',
  quantity_change: 10,
  reason: 'po-receipt',
  created_by: 'user-1',
  created_at: '2026-04-01T00:00:00Z',
};

function renderPage(role = 'administrator') {
  mockUseAuth.mockReturnValue({
    user: { id: 'user-1', role, display_name: 'Test User', email: 'test@test.com' },
    isAuthenticated: true,
    isLoading: false,
  });
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <InventoryPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('InventoryPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseInventorySnapshots.mockReturnValue({
      data: { data: [MOCK_SNAPSHOT] },
      isLoading: false,
      error: null,
    });
    mockUseInventoryAdjustments.mockReturnValue({
      data: { data: [MOCK_ADJUSTMENT], pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 } },
      isLoading: false,
      error: null,
    });
    mockCreateAdjustmentMutateAsync.mockResolvedValue(MOCK_ADJUSTMENT);
  });

  it('renders page title', () => {
    renderPage();
    expect(screen.getByRole('heading', { name: /inventory/i })).toBeInTheDocument();
  });

  it('renders Snapshots and Adjustments tabs', () => {
    renderPage();
    expect(screen.getByRole('tab', { name: /snapshots/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /adjustments/i })).toBeInTheDocument();
  });

  it('renders snapshot row with On Hand quantity', () => {
    renderPage();
    expect(screen.getByText('50')).toBeInTheDocument();
  });

  it('Add Adjustment button visible for manage_inventory role', () => {
    renderPage('administrator');
    expect(screen.getByRole('button', { name: /create adjustment/i })).toBeInTheDocument();
  });

  it('Add Adjustment button hidden for coach role', () => {
    renderPage('coach');
    expect(screen.queryByRole('button', { name: /create adjustment/i })).not.toBeInTheDocument();
  });

  it('switches to adjustments tab on click', async () => {
    renderPage();
    const user = userEvent.setup();
    const adjTab = screen.getByRole('tab', { name: /adjustments/i });
    await user.click(adjTab);

    await waitFor(() => {
      expect(screen.getByText('po-receipt')).toBeInTheDocument();
    });
  });

  it('opens adjustment dialog and submits', async () => {
    renderPage('administrator');
    const user = userEvent.setup();

    const addBtn = screen.getByRole('button', { name: /create adjustment/i });
    await user.click(addBtn);

    const itemField = await screen.findByLabelText(/item id/i);
    await user.type(itemField, '99999999-9999-9999-9999-999999999999');

    const qtyField = await screen.findByLabelText(/quantity change/i);
    await user.clear(qtyField);
    await user.type(qtyField, '5');

    const reasonField = screen.getByLabelText(/reason/i);
    await user.type(reasonField, 'manual-recount');

    const createBtn = screen.getByRole('button', { name: /^create$/i });
    await user.click(createBtn);

    await waitFor(() => {
      expect(mockCreateAdjustmentMutateAsync).toHaveBeenCalled();
    });
  });
});
