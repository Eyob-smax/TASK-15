import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import SuppliersPage from '@/routes/SuppliersPage';

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

const mockUseSupplierList = vi.fn();
const mockCreateMutateAsync = vi.fn();

vi.mock('@/lib/hooks/useProcurement', () => ({
  useSupplierList: (params: unknown) => mockUseSupplierList(params),
  useCreateSupplier: () => ({ mutateAsync: mockCreateMutateAsync, isPending: false }),
}));

vi.mock('@/lib/notifications', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn() }),
}));

const MOCK_SUPPLIER = {
  id: 'ssssssss-ssss-ssss-ssss-ssssssssssss',
  name: 'Acme Fitness Supplies',
  contact_name: 'Jane Smith',
  contact_email: 'jane@acme.test',
  contact_phone: '555-0100',
  address: '123 Main St',
  payment_terms: 'Net 30',
  lead_time_days: 14,
  notes: '',
  is_active: true,
  created_at: '2026-04-01T00:00:00Z',
  updated_at: '2026-04-01T00:00:00Z',
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
        <SuppliersPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('SuppliersPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseSupplierList.mockReturnValue({
      data: {
        data: [MOCK_SUPPLIER],
        pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
      },
      isLoading: false,
      error: null,
    });
    mockCreateMutateAsync.mockResolvedValue(MOCK_SUPPLIER);
  });

  it('renders page title', () => {
    renderPage();
    expect(screen.getByRole('heading', { name: /suppliers/i })).toBeInTheDocument();
  });

  it('renders supplier name in table', () => {
    renderPage();
    expect(screen.getByText('Acme Fitness Supplies')).toBeInTheDocument();
  });

  it('shows Active chip for active supplier', () => {
    renderPage();
    expect(screen.getAllByText('Active').length).toBeGreaterThan(0);
  });

  it('Add Supplier button visible for manage_procurement role', () => {
    renderPage('administrator');
    expect(screen.getByRole('button', { name: /new supplier/i })).toBeInTheDocument();
  });

  it('Add Supplier button hidden for member role', () => {
    renderPage('member');
    expect(screen.queryByRole('button', { name: /new supplier/i })).not.toBeInTheDocument();
  });

  it('shows error alert when supplier list fails', () => {
    mockUseSupplierList.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('network error'),
    });
    renderPage();
    expect(screen.getByText(/failed to load suppliers/i)).toBeInTheDocument();
  });

  it('opens create dialog and submits new supplier', async () => {
    renderPage('administrator');
    const user = userEvent.setup();

    const addBtn = screen.getByRole('button', { name: /new supplier/i });
    await user.click(addBtn);

    const dialog = await screen.findByRole('dialog');
    const nameField = within(dialog).getByRole('textbox', { name: /^name/i });
    fireEvent.change(nameField, { target: { value: 'New Supplier Co' } });

    const saveBtn = within(dialog).getByRole('button', { name: /^create$/i });
    expect(saveBtn).toBeEnabled();
    await user.click(saveBtn);

    await waitFor(() => {
      expect(mockCreateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ name: 'New Supplier Co' }),
      );
    });
  });
});
