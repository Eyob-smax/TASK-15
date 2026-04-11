import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import VariancesPage from '@/routes/VariancesPage';

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

const mockUseVarianceList = vi.fn();
const mockResolveMutateAsync = vi.fn();

vi.mock('@/lib/hooks/useProcurement', () => ({
  useVarianceList: (params: unknown) => mockUseVarianceList(params),
  useResolveVariance: () => ({ mutateAsync: mockResolveMutateAsync, isPending: false }),
}));

vi.mock('@/lib/notifications', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn() }),
}));

const MOCK_VARIANCE = {
  id: 'vvvvvvvv-vvvv-vvvv-vvvv-vvvvvvvvvvvv',
  po_line_id: 'llllllll-llll-llll-llll-llllllllllll',
  type: 'shortage',
  expected_value: 10,
  actual_value: 7,
  difference_amount: 3,
  status: 'open',
  resolution_action: '',
  resolution_notes: '',
  quantity_change: null,
  resolved_at: null,
  requires_escalation: false,
  is_overdue: false,
  created_at: '2026-04-01T00:00:00Z',
};

const MOCK_OVERDUE_VARIANCE = {
  ...MOCK_VARIANCE,
  id: 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
  is_overdue: true,
};

const MOCK_ESCALATED_VARIANCE = {
  ...MOCK_VARIANCE,
  id: 'ssssssss-ssss-ssss-ssss-ssssssssssss',
  status: 'escalated',
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
        <VariancesPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('VariancesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseVarianceList.mockReturnValue({
      data: { data: [MOCK_VARIANCE], pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 } },
      isLoading: false,
      error: null,
    });
    mockResolveMutateAsync.mockResolvedValue({});
  });

  it('renders page title', () => {
    renderPage();
    expect(screen.getByRole('heading', { name: /variances/i })).toBeInTheDocument();
  });

  it('renders variance rows with type and status', () => {
    renderPage();
    expect(screen.getByText('shortage')).toBeInTheDocument();
    expect(screen.getByText('open')).toBeInTheDocument();
  });

  it('shows Overdue chip when variance is overdue', () => {
    mockUseVarianceList.mockReturnValue({
      data: { data: [MOCK_OVERDUE_VARIANCE], pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 } },
      isLoading: false,
      error: null,
    });
    renderPage();
    expect(screen.getByText('Overdue')).toBeInTheDocument();
  });

  it('Resolve button is visible for administrator role', () => {
    renderPage('administrator');
    expect(screen.getByRole('button', { name: /resolve/i })).toBeInTheDocument();
  });

  it('Resolve button remains visible for escalated variances', () => {
    mockUseVarianceList.mockReturnValue({
      data: { data: [MOCK_ESCALATED_VARIANCE], pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 } },
      isLoading: false,
      error: null,
    });
    renderPage('administrator');
    expect(screen.getByRole('button', { name: /resolve/i })).toBeInTheDocument();
    expect(screen.getByText(/escalated/i)).toBeInTheDocument();
  });

  it('Resolve button is hidden for member role', () => {
    renderPage('member');
    expect(screen.queryByRole('button', { name: /resolve/i })).not.toBeInTheDocument();
  });

  it('opens resolve dialog and submits adjustment payload', async () => {
    renderPage('administrator');
    const resolveBtn = screen.getByRole('button', { name: /resolve/i });
    fireEvent.click(resolveBtn);

    const dialog = await screen.findByRole('dialog');

    const quantityField = within(dialog).getByLabelText(/quantity change/i);
    fireEvent.change(quantityField, { target: { value: '2' } });

    const notesField = within(dialog).getByLabelText(/resolution notes/i);
    fireEvent.change(notesField, { target: { value: 'ok' } });

    const submitBtn = within(dialog).getByRole('button', { name: /^resolve$/i });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mockResolveMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          id: MOCK_VARIANCE.id,
          action: 'adjustment',
          resolution_notes: 'ok',
          quantity_change: 2,
        }),
      );
    });
  }, 10000);

  it('shows error alert when variance list fails to load', () => {
    mockUseVarianceList.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('network error'),
    });
    renderPage();
    expect(screen.getByText(/failed to load variances/i)).toBeInTheDocument();
  });
});
