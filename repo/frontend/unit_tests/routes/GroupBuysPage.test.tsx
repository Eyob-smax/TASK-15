import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import GroupBuysPage from '@/routes/GroupBuysPage';

const mockUseAuth = vi.fn();
vi.mock('@/lib/auth', async () => {
  const actual = await vi.importActual<typeof import('@/lib/auth')>('@/lib/auth');
  return {
    ...actual,
    useAuth: () => mockUseAuth(),
    RequireRole: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  };
});

const mockUseCampaignList = vi.fn();
const mockCreateCampaignMutateAsync = vi.fn();
const mockUseOfflineStatus = vi.fn();

vi.mock('@/lib/hooks/useCampaigns', () => ({
  useCampaignList: () => mockUseCampaignList(),
  useCreateCampaign: () => ({ mutateAsync: mockCreateCampaignMutateAsync, isPending: false }),
}));

vi.mock('@/lib/notifications', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn() }),
}));

vi.mock('@/lib/offline', () => ({
  useOfflineStatus: () => mockUseOfflineStatus(),
  OFFLINE_MUTATION_MESSAGE: 'Reconnect to make changes.',
}));

const MOCK_CAMPAIGN = {
  id: 'cccccccc-cccc-cccc-cccc-cccccccccccc',
  item_id: '11111111-1111-1111-1111-111111111111',
  min_quantity: 10,
  current_committed_qty: 4,
  cutoff_time: '2026-05-01T00:00:00Z',
  status: 'active',
  created_by: 'admin-user',
  created_at: '2026-04-01T00:00:00Z',
  evaluated_at: null,
};

function renderPage(role = 'administrator', initialEntries = ['/group-buys']) {
  mockUseAuth.mockReturnValue({
    user: { id: 'user-1', role },
    isAuthenticated: true,
    isLoading: false,
  });
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={initialEntries}>
        <GroupBuysPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('GroupBuysPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseOfflineStatus.mockReturnValue({
      isOnline: true,
      isOffline: false,
      lastSyncAt: null,
    });
    mockUseCampaignList.mockReturnValue({
      data: {
        data: [MOCK_CAMPAIGN],
        pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
      },
      isLoading: false,
      error: null,
    });
  });

  it('renders backend-shaped campaign rows', () => {
    renderPage();
    expect(screen.getByText('Item 11111111')).toBeInTheDocument();
    expect(screen.getByText('4/10')).toBeInTheDocument();
  });

  it('create dialog only shows supported campaign fields', async () => {
    renderPage();
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: /create campaign/i }));

    expect(screen.getByLabelText(/item id/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/minimum quantity/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/cutoff time/i)).toBeInTheDocument();
    expect(screen.queryByLabelText(/title/i)).not.toBeInTheDocument();
    expect(screen.queryByLabelText(/discount/i)).not.toBeInTheDocument();
  });

  it('hides the generic create button for members', () => {
    renderPage('member');
    expect(screen.queryByRole('button', { name: /create campaign/i })).not.toBeInTheDocument();
  });

  it('opens the member start flow from an item context and locks the item id', async () => {
    renderPage('member', ['/group-buys?item_id=11111111-1111-1111-1111-111111111111']);

    expect(await screen.findByText(/start group buy/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/item id/i)).toBeDisabled();
    expect(screen.getByRole('button', { name: /^start$/i })).toBeInTheDocument();
  });

  it('disables campaign creation while offline', async () => {
    mockUseOfflineStatus.mockReturnValue({
      isOnline: false,
      isOffline: true,
      lastSyncAt: Date.UTC(2026, 3, 11, 10, 30, 0),
    });
    renderPage();
    expect(screen.getByRole('button', { name: /create campaign/i })).toBeDisabled();
  });
});
