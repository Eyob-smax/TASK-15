import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import UsersPage from '@/routes/UsersPage';

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

const mockDeactivateMutateAsync = vi.fn();

const mockUseUserList = vi.fn();
const mockUseCreateUser = vi.fn();
const mockUseDeactivateUser = vi.fn();

vi.mock('@/lib/hooks/useAdmin', () => ({
  useUserList: (params: unknown) => mockUseUserList(params),
  useCreateUser: () => mockUseCreateUser(),
  useDeactivateUser: () => mockUseDeactivateUser(),
  useAuditLog: () => ({ data: undefined, isLoading: false, error: null }),
  useSecurityEvents: () => ({ data: undefined, isLoading: false, error: null }),
}));

vi.mock('@/lib/notifications', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn() }),
}));

const MOCK_USER = {
  id: 'user-1',
  email: 'admin@test.com',
  display_name: 'Admin User',
  role: 'administrator',
  status: 'active',
  location_id: null,
  created_at: '2026-04-09T10:00:00Z',
  updated_at: '2026-04-09T10:00:00Z',
};

const MOCK_USER_LIST = [MOCK_USER];

function renderPage() {
  mockUseAuth.mockReturnValue({
    user: { id: 'user-1', role: 'administrator', display_name: 'Admin User', email: 'admin@test.com' },
    isAuthenticated: true,
    isLoading: false,
  });
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <UsersPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('UsersPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseUserList.mockReturnValue({
      data: {
        data: MOCK_USER_LIST,
        pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
      },
      isLoading: false,
      error: null,
    });
    mockUseCreateUser.mockReturnValue({ mutateAsync: vi.fn(), isPending: false });
    mockUseDeactivateUser.mockReturnValue({
      mutateAsync: mockDeactivateMutateAsync,
      isPending: false,
    });
  });

  it('renders page title', () => {
    renderPage();
    expect(screen.getByRole('heading', { name: 'Users' })).toBeInTheDocument();
  });

  it('shows users in table', () => {
    renderPage();
    expect(screen.getByText('Admin User')).toBeInTheDocument();
    expect(screen.getByText('admin@test.com')).toBeInTheDocument();
  });

  it('shows loading state', () => {
    mockUseUserList.mockReturnValue({ data: undefined, isLoading: true, error: null });
    renderPage();
    // No user rows rendered while loading
    expect(screen.queryByText('Admin User')).not.toBeInTheDocument();
  });

  it('shows error state', () => {
    mockUseUserList.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('fail'),
    });
    renderPage();
    expect(screen.getByText(/failed to load users/i)).toBeInTheDocument();
  });

  it('create user button is present', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /create user/i })).toBeInTheDocument();
  });

  it('opens create user dialog on button click', async () => {
    renderPage();
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: /create user/i }));

    const dialog = await screen.findByRole('dialog');
    expect(within(dialog).getByRole('heading', { name: 'Create User' })).toBeInTheDocument();
  });

  it('deactivate button present per user row', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /deactivate/i })).toBeInTheDocument();
  });

  it('deactivate calls mutation after confirmation', async () => {
    renderPage();
    const user = userEvent.setup();

    // Click Deactivate button on the row to open the confirm dialog
    await user.click(screen.getByRole('button', { name: /deactivate/i }));

    // The ConfirmDialog should now be open — click the confirm button
    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument();
    });

    const confirmButton = screen.getByRole('button', { name: /^deactivate$/i });
    await user.click(confirmButton);

    await waitFor(() => {
      expect(mockDeactivateMutateAsync).toHaveBeenCalledWith(MOCK_USER.id);
    });
  });
});
