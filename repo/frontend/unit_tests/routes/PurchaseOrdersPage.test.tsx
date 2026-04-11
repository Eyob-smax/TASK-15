import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import PurchaseOrdersPage from '@/routes/PurchaseOrdersPage';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: () => mockNavigate };
});

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

const mockUsePOList = vi.fn();
vi.mock('@/lib/hooks/useProcurement', () => ({
  usePOList: (params: unknown) => mockUsePOList(params),
  useSupplierList: () => ({ data: { data: [], pagination: { page: 1, page_size: 20, total_count: 0, total_pages: 0 } } }),
  useCreatePO: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useApprovePO: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useVoidPO: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

const MOCK_PO = {
  id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  supplier_id: 'supplier-1',
  status: 'created',
  total_amount: 1500,
  created_by: 'user-1',
  approved_by: null,
  created_at: '2026-04-09T10:00:00Z',
  updated_at: '2026-04-09T10:00:00Z',
};

const MOCK_PO_LIST = [MOCK_PO];

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
        <PurchaseOrdersPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('PurchaseOrdersPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUsePOList.mockReturnValue({
      data: {
        data: MOCK_PO_LIST,
        pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
      },
      isLoading: false,
      error: null,
    });
  });

  it('renders page title', () => {
    renderPage();
    expect(screen.getByRole('heading', { name: 'Purchase Orders' })).toBeInTheDocument();
  });

  it('shows purchase order rows', () => {
    renderPage();
    // The table renders a shortened ID (first 8 chars + ellipsis) and status
    const idCell = screen.getByText('aaaaaaaa…');
    expect(idCell).toBeInTheDocument();

    const row = idCell.closest('tr');
    expect(row).not.toBeNull();
    expect(within(row as HTMLTableRowElement).getByText(/^created$/i)).toBeInTheDocument();
  });

  it('shows loading skeleton when isLoading', () => {
    mockUsePOList.mockReturnValue({ data: undefined, isLoading: true, error: null });
    renderPage();
    // No data rows rendered while loading
    expect(screen.queryByText('aaaaaaaa…')).not.toBeInTheDocument();
  });

  it('shows error alert on error', () => {
    mockUsePOList.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('fail'),
    });
    renderPage();
    expect(screen.getByText(/failed to load purchase orders/i)).toBeInTheDocument();
  });

  it('status filter renders in FilterBar', () => {
    renderPage();
    // The FilterBar renders the Status select field
    expect(screen.getByLabelText(/status/i)).toBeInTheDocument();
  });

  it('navigates to detail on row PO ID click', async () => {
    renderPage();
    const user = userEvent.setup();
    const idCell = screen.getByText('aaaaaaaa…');
    await user.click(idCell);
    expect(mockNavigate).toHaveBeenCalledWith(
      `/procurement/purchase-orders/${MOCK_PO.id}`,
    );
  });

  it('create PO button visible for manage_procurement role', () => {
    renderPage('administrator');
    expect(screen.getByRole('button', { name: /new po/i })).toBeInTheDocument();
  });

  it('pagination controls present when totalCount > 0', () => {
    renderPage();
    // MUI TablePagination renders rows-per-page and page info
    expect(screen.getByText(/rows per page/i)).toBeInTheDocument();
  });
});
