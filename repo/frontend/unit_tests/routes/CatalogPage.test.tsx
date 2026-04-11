import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import CatalogPage from '@/routes/CatalogPage';

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

const mockUseItemList = vi.fn();
const mockBatchEditMutateAsync = vi.fn();
vi.mock('@/lib/hooks/useItems', () => ({
  useItemList: (params: unknown) => mockUseItemList(params),
  useBatchEdit: () => ({ mutateAsync: mockBatchEditMutateAsync, isPending: false }),
}));

vi.mock('@/lib/notifications', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn(), warning: vi.fn(), info: vi.fn() }),
}));

const MOCK_ITEMS = [
  {
    id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    name: 'Test Barbell',
    category: 'Weights',
    brand: 'PowerBar',
    condition: 'new',
    billing_model: 'one_time',
    unit_price: 199.99,
    refundable_deposit: 50,
    status: 'published',
  },
  {
    id: 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    name: 'Yoga Mat',
    category: 'Accessories',
    brand: 'FlexMat',
    condition: 'used',
    billing_model: 'monthly_rental',
    unit_price: 29.99,
    refundable_deposit: 10,
    status: 'draft',
  },
];

function renderCatalogPage(role = 'administrator') {
  mockUseAuth.mockReturnValue({
    user: { id: 'user-1', role, display_name: 'Test User', email: 'test@test.com' },
    isAuthenticated: true,
    isLoading: false,
  });
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <CatalogPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('CatalogPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseItemList.mockReturnValue({
      data: {
        data: MOCK_ITEMS,
        pagination: { page: 1, page_size: 20, total_count: 2, total_pages: 1 },
      },
      isLoading: false,
      error: null,
    });
    mockBatchEditMutateAsync.mockResolvedValue({
      job_id: 'job-1',
      total_rows: 1,
      success_count: 1,
      failure_count: 0,
      results: [
        {
          item_id: MOCK_ITEMS[0].id,
          field: 'availability_windows',
          old_value: '[]',
          new_value: '[2026-05-01T09:00:00Z|2026-05-01T12:00:00Z]',
          success: true,
        },
      ],
    });
  });

  it('renders page title', () => {
    renderCatalogPage();
    expect(screen.getByRole('heading', { name: 'Catalog' })).toBeInTheDocument();
  });

  it('renders filter bar fields', () => {
    renderCatalogPage();
    expect(screen.getByLabelText(/category/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/brand/i)).toBeInTheDocument();
  });

  it('renders item rows from API response', () => {
    renderCatalogPage();
    expect(screen.getByText('Test Barbell')).toBeInTheDocument();
    expect(screen.getByText('Yoga Mat')).toBeInTheDocument();
  });

  it('shows New Item button for admin', () => {
    renderCatalogPage('administrator');
    expect(screen.getByRole('button', { name: /new item/i })).toBeInTheDocument();
  });

  it('shows Batch Edit button for admin', () => {
    renderCatalogPage('administrator');
    expect(screen.getByRole('button', { name: /batch edit/i })).toBeInTheDocument();
  });

  it('hides New Item button for member', () => {
    renderCatalogPage('member');
    expect(screen.queryByRole('button', { name: /new item/i })).not.toBeInTheDocument();
  });

  it('hides Batch Edit button for member', () => {
    renderCatalogPage('member');
    expect(screen.queryByRole('button', { name: /batch edit/i })).not.toBeInTheDocument();
  });

  it('navigates to new item form when button clicked', async () => {
    renderCatalogPage('administrator');
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: /new item/i }));
    expect(mockNavigate).toHaveBeenCalledWith('/catalog/new');
  });

  it('submits availability-window batch edits from the dialog', async () => {
    renderCatalogPage('administrator');

    fireEvent.click(screen.getByRole('button', { name: /batch edit/i }));
    const dialog = await screen.findByRole('dialog');

    fireEvent.mouseDown(within(dialog).getByRole('combobox', { name: /field/i }));
    fireEvent.click(await screen.findByRole('option', { name: /availability windows/i }));
    fireEvent.click(within(dialog).getByRole('button', { name: /add availability window/i }));

    const startField = within(dialog).getByLabelText(/^start$/i);
    const endField = within(dialog).getByLabelText(/^end$/i);
    fireEvent.change(startField, { target: { value: '2026-05-01T09:00' } });
    fireEvent.change(endField, { target: { value: '2026-05-01T12:00' } });

    fireEvent.click(within(dialog).getByRole('button', { name: /apply batch edit/i }));

    await waitFor(() => {
      expect(mockBatchEditMutateAsync).toHaveBeenCalledWith([
        {
          item_id: MOCK_ITEMS[0].id,
          field: 'availability_windows',
          availability_windows: [
            {
              start_time: new Date('2026-05-01T09:00').toISOString(),
              end_time: new Date('2026-05-01T12:00').toISOString(),
            },
          ],
        },
      ]);
    });
  }, 10000);

  it('shows loading skeleton when isLoading', () => {
    mockUseItemList.mockReturnValue({ data: undefined, isLoading: true, error: null });
    renderCatalogPage();
    // DataTable shows skeletons — no rows rendered
    expect(screen.queryByText('Test Barbell')).not.toBeInTheDocument();
  });

  it('shows error alert when query fails', () => {
    mockUseItemList.mockReturnValue({ data: undefined, isLoading: false, error: new Error('fail') });
    renderCatalogPage();
    expect(screen.getByText(/failed to load items/i)).toBeInTheDocument();
  });

  it('calls useItemList with filter values when filters change', async () => {
    renderCatalogPage();
    const categoryInput = screen.getByLabelText(/category/i);
    fireEvent.change(categoryInput, { target: { value: 'Weights' } });
    await waitFor(() => {
      expect(mockUseItemList).toHaveBeenCalledWith(
        expect.objectContaining({ category: 'Weights' }),
      );
    });
  });
});
