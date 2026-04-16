import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, within, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import CatalogDetailPage from '@/routes/CatalogDetailPage';

const mockUseAuth = vi.fn();
vi.mock('@/lib/auth', async () => {
  const actual = await vi.importActual<typeof import('@/lib/auth')>('@/lib/auth');
  return {
    ...actual,
    useAuth: () => mockUseAuth(),
    RequireRole: ({ children, roles }: { children: React.ReactNode; roles: string[] }) => {
      const u = mockUseAuth();
      return roles.includes(u?.user?.role) ? <>{children}</> : null;
    },
  };
});

const mockOffline = vi.fn();
vi.mock('@/lib/offline', () => ({
  useOfflineStatus: () => mockOffline(),
}));

const mockUseItem = vi.fn();
const mockPublishMutate = vi.fn();
const mockUnpublishMutate = vi.fn();
vi.mock('@/lib/hooks/useItems', () => ({
  useItem: (id: string | undefined) => mockUseItem(id),
  usePublishItem: () => ({ mutateAsync: mockPublishMutate, isPending: false }),
  useUnpublishItem: () => ({ mutateAsync: mockUnpublishMutate, isPending: false }),
}));

vi.mock('@/lib/notifications', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn(), info: vi.fn() }),
}));

const ITEM_ID = 'iiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii';

const BASE_ITEM = {
  id: ITEM_ID,
  name: 'Mountain Bike',
  description: 'A sturdy trail bike.',
  category: 'bikes',
  brand: 'TrailMaster',
  sku: 'SKU-MB-001',
  condition: 'new',
  billing_model: 'flat_rate',
  unit_price: 499.99,
  refundable_deposit: 50,
  quantity: 10,
  location_id: 'loc-1',
  version: 3,
  status: 'draft',
  availability_windows: [],
  blackout_windows: [],
};

function renderPage(role = 'administrator') {
  mockUseAuth.mockReturnValue({
    user: { id: 'u1', role, display_name: 'Test', email: 't@t.com' },
    isAuthenticated: true,
    isLoading: false,
  });
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[`/catalog/${ITEM_ID}`]}>
        <Routes>
          <Route path="/catalog/:id" element={<CatalogDetailPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('CatalogDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockOffline.mockReturnValue({ isOnline: true, isOffline: false, lastSyncAt: null });
    mockUseItem.mockReturnValue({
      data: BASE_ITEM,
      isLoading: false,
      error: null,
      dataUpdatedAt: Date.now(),
    });
  });

  it('renders skeleton while loading', () => {
    mockUseItem.mockReturnValue({ data: undefined, isLoading: true, error: null, dataUpdatedAt: 0 });
    const { container } = renderPage();
    expect(container.querySelector('.MuiSkeleton-root')).not.toBeNull();
  });

  it('renders error alert when item cannot be loaded', () => {
    mockUseItem.mockReturnValue({ data: undefined, isLoading: false, error: null, dataUpdatedAt: 0 });
    renderPage();
    expect(screen.getByText(/failed to load item details/i)).toBeInTheDocument();
  });

  it('renders item fields with condition/billing labels and prices', () => {
    renderPage();
    expect(screen.getAllByText(/mountain bike/i).length).toBeGreaterThan(0);
    expect(screen.getByText('A sturdy trail bike.')).toBeInTheDocument();
    expect(screen.getByText('SKU-MB-001')).toBeInTheDocument();
    expect(screen.getByText('$499.99')).toBeInTheDocument();
    expect(screen.getByText('$50.00')).toBeInTheDocument();
  });

  it('admin sees Publish (draft) and Edit buttons', () => {
    renderPage('administrator');
    expect(screen.getByRole('button', { name: /^publish$/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^edit$/i })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /^unpublish$/i })).not.toBeInTheDocument();
  });

  it('admin sees Unpublish when item is published', () => {
    mockUseItem.mockReturnValue({
      data: { ...BASE_ITEM, status: 'published' },
      isLoading: false,
      error: null,
      dataUpdatedAt: Date.now(),
    });
    renderPage('administrator');
    expect(screen.getByRole('button', { name: /^unpublish$/i })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /^publish$/i })).not.toBeInTheDocument();
  });

  it('member sees Start Group Buy when published, not when draft', () => {
    mockUseItem.mockReturnValue({
      data: { ...BASE_ITEM, status: 'published' },
      isLoading: false,
      error: null,
      dataUpdatedAt: Date.now(),
    });
    renderPage('member');
    expect(screen.getByRole('button', { name: /start group buy/i })).toBeInTheDocument();
  });

  it('member does not see Start Group Buy when draft', () => {
    renderPage('member');
    expect(screen.queryByRole('button', { name: /start group buy/i })).not.toBeInTheDocument();
  });

  it('admin action buttons are disabled when offline', () => {
    mockOffline.mockReturnValue({ isOnline: false, isOffline: true, lastSyncAt: null });
    renderPage('administrator');
    expect(screen.getByRole('button', { name: /^publish$/i })).toBeDisabled();
    expect(screen.getByRole('button', { name: /^edit$/i })).toBeDisabled();
  });

  it('opens publish confirm and calls publish mutation', async () => {
    mockPublishMutate.mockResolvedValue({});
    renderPage('administrator');
    const user = userEvent.setup();

    await user.click(screen.getByRole('button', { name: /^publish$/i }));
    const dialog = await screen.findByRole('dialog');
    expect(within(dialog).getByText(/publish item/i)).toBeInTheDocument();

    const confirm = within(dialog).getByRole('button', { name: /^publish$/i });
    await user.click(confirm);
    await waitFor(() => expect(mockPublishMutate).toHaveBeenCalledWith(ITEM_ID));
  });

  it('opens unpublish confirm and calls unpublish mutation', async () => {
    mockUseItem.mockReturnValue({
      data: { ...BASE_ITEM, status: 'published' },
      isLoading: false,
      error: null,
      dataUpdatedAt: Date.now(),
    });
    mockUnpublishMutate.mockResolvedValue({});
    renderPage('administrator');
    const user = userEvent.setup();

    await user.click(screen.getByRole('button', { name: /^unpublish$/i }));
    const dialog = await screen.findByRole('dialog');
    expect(within(dialog).getByText(/unpublish item/i)).toBeInTheDocument();

    const confirm = within(dialog).getByRole('button', { name: /^unpublish$/i });
    await user.click(confirm);
    await waitFor(() => expect(mockUnpublishMutate).toHaveBeenCalledWith(ITEM_ID));
  });

  it('renders availability and blackout window sections when present', () => {
    mockUseItem.mockReturnValue({
      data: {
        ...BASE_ITEM,
        availability_windows: [
          { id: 'aw1', start_time: '2026-04-01T10:00:00Z', end_time: '2026-04-01T18:00:00Z' },
        ],
        blackout_windows: [
          { id: 'bw1', start_time: '2026-05-01T00:00:00Z', end_time: '2026-05-02T00:00:00Z' },
        ],
      },
      isLoading: false,
      error: null,
      dataUpdatedAt: Date.now(),
    });
    renderPage();
    expect(screen.getByText(/availability windows/i)).toBeInTheDocument();
    expect(screen.getByText(/blackout windows/i)).toBeInTheDocument();
  });

  it('shows a stale-data warning alert when sync error is present', () => {
    mockUseItem.mockReturnValue({
      data: BASE_ITEM,
      isLoading: false,
      error: new Error('sync failed'),
      dataUpdatedAt: Date.now(),
    });
    renderPage();
    expect(screen.getByText(/catalog sync is temporarily unavailable/i)).toBeInTheDocument();
  });
});
