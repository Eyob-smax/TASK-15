import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import AuditPage from '@/routes/AuditPage';

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

const mockUseAuditLog = vi.fn();
const mockUseSecurityEvents = vi.fn();

vi.mock('@/lib/hooks/useAdmin', () => ({
  useAuditLog: (params: unknown) => mockUseAuditLog(params),
  useSecurityEvents: (params: unknown) => mockUseSecurityEvents(params),
  useUserList: () => ({ data: undefined, isLoading: false, error: null }),
  useCreateUser: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useDeactivateUser: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

const MOCK_AUDIT_EVENT = {
  id: 'audit-1',
  event_type: 'auth.login.success',
  entity_type: 'user',
  entity_id: 'user-1',
  actor_id: 'user-1',
  details: {},
  integrity_hash: 'abc123',
  created_at: '2026-04-09T10:00:00Z',
};

const MOCK_SECURITY_EVENT = {
  id: 'audit-2',
  event_type: 'auth.login.failure',
  entity_type: 'user',
  entity_id: 'user-2',
  actor_id: 'user-2',
  details: {},
  integrity_hash: 'def456',
  created_at: '2026-04-09T11:00:00Z',
};

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
        <AuditPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('AuditPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseAuditLog.mockReturnValue({
      data: {
        data: [MOCK_AUDIT_EVENT],
        pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
      },
      isLoading: false,
      error: null,
    });
    mockUseSecurityEvents.mockReturnValue({
      data: {
        data: [MOCK_SECURITY_EVENT],
        pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
      },
      isLoading: false,
      error: null,
    });
  });

  it('renders Audit Log title', () => {
    renderPage();
    expect(screen.getByRole('heading', { name: 'Audit Log' })).toBeInTheDocument();
  });

  it('shows All Events tab by default', () => {
    renderPage();
    expect(screen.getByRole('tab', { name: /all events/i })).toBeInTheDocument();
    // The All Events tab should be selected by default
    const allEventsTab = screen.getByRole('tab', { name: /all events/i });
    expect(allEventsTab).toHaveAttribute('aria-selected', 'true');
  });

  it('shows Security Events tab', () => {
    renderPage();
    expect(screen.getByRole('tab', { name: /security events/i })).toBeInTheDocument();
  });

  it('switching to security tab shows security events', async () => {
    renderPage();
    const user = userEvent.setup();

    // Initially showing all events
    expect(screen.getByText('auth.login.success')).toBeInTheDocument();

    // Click the Security Events tab
    await user.click(screen.getByRole('tab', { name: /security events/i }));

    // Security events content should now be visible
    await waitFor(() => {
      expect(screen.getByText('auth.login.failure')).toBeInTheDocument();
    });

    // All events content is no longer rendered (tab === 0 condition hides it)
    expect(screen.queryByText('auth.login.success')).not.toBeInTheDocument();
  });

  it('shows loading state', () => {
    mockUseAuditLog.mockReturnValue({ data: undefined, isLoading: true, error: null });
    renderPage();
    // No audit event rows rendered while loading
    expect(screen.queryByText('auth.login.success')).not.toBeInTheDocument();
  });

  it('shows error state', () => {
    mockUseAuditLog.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('fail'),
    });
    renderPage();
    expect(screen.getByText(/failed to load audit events/i)).toBeInTheDocument();
  });
});
