import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import BackupsPage from '@/routes/BackupsPage';

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

const mockUseBackupList = vi.fn();
const mockUseTriggerBackup = vi.fn();

vi.mock('@/lib/hooks/useAdmin', () => ({
  useBackupList: (params: unknown) => mockUseBackupList(params),
  useTriggerBackup: () => mockUseTriggerBackup(),
}));

vi.mock('@/lib/notifications', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn() }),
}));

const MOCK_BACKUP: import('@/lib/types').BackupRun = {
  id: 'backup-001',
  archive_path: '/backups/2026-04-10.tar.gz',
  checksum: 'abc123def456',
  checksum_algorithm: 'sha256',
  status: 'completed',
  file_size: 1048576,
  started_at: '2026-04-10T08:00:00Z',
  completed_at: '2026-04-10T08:05:00Z',
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
        <BackupsPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('BackupsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseBackupList.mockReturnValue({
      data: {
        data: [MOCK_BACKUP],
        pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
      },
      isLoading: false,
      error: null,
    });
    mockUseTriggerBackup.mockReturnValue({
      mutateAsync: vi.fn(),
      isPending: false,
    });
  });

  it('renders the Backups title', () => {
    renderPage();
    expect(screen.getByRole('heading', { name: 'Backups' })).toBeInTheDocument();
  });

  it('renders the file_size column with formatted bytes', () => {
    renderPage();
    // 1048576 bytes = 1024.0 KB = 1.0 MB
    expect(screen.getByText('1.0 MB')).toBeInTheDocument();
  });

  it('renders the checksum column', () => {
    renderPage();
    // shortStr('abc123def456', 12) = 'abc123def456'
    expect(screen.getByText('abc123def456')).toBeInTheDocument();
  });

  it('renders trigger backup button', () => {
    renderPage();
    expect(screen.getByRole('button', { name: /trigger backup/i })).toBeInTheDocument();
  });

  it('shows loading state', () => {
    mockUseBackupList.mockReturnValue({ data: undefined, isLoading: true, error: null });
    renderPage();
    expect(screen.queryByText('1.0 MB')).not.toBeInTheDocument();
  });

  it('shows error state', () => {
    mockUseBackupList.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('fail'),
    });
    renderPage();
    expect(screen.getByText(/failed to load backups/i)).toBeInTheDocument();
  });
});
