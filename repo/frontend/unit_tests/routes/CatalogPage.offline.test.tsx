import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import CatalogPage from '@/routes/CatalogPage';

const mockUseAuth = vi.fn();
const mockUseItemList = vi.fn();
const mockUseBatchEdit = vi.fn();
const mockUseOfflineStatus = vi.fn();

vi.mock('@/lib/auth', async () => {
  const actual = await vi.importActual<typeof import('@/lib/auth')>('@/lib/auth');
  return {
    ...actual,
    useAuth: () => mockUseAuth(),
    RequireRole: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  };
});

vi.mock('@/lib/hooks/useItems', () => ({
  useItemList: () => mockUseItemList(),
  useBatchEdit: () => mockUseBatchEdit(),
}));

vi.mock('@/lib/notifications', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn(), warning: vi.fn() }),
}));

vi.mock('@/lib/offline', () => ({
  useOfflineStatus: () => mockUseOfflineStatus(),
  OFFLINE_MUTATION_MESSAGE: 'Reconnect to make changes.',
}));

describe('CatalogPage offline behavior', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseAuth.mockReturnValue({
      user: { id: 'admin-1', role: 'administrator' },
      isAuthenticated: true,
      isLoading: false,
    });
    mockUseBatchEdit.mockReturnValue({
      mutateAsync: vi.fn(),
      isPending: false,
    });
    mockUseOfflineStatus.mockReturnValue({
      isOnline: false,
      isOffline: true,
      lastSyncAt: Date.UTC(2026, 3, 11, 10, 30, 0),
    });
  });

  it('shows cached catalog rows while offline', () => {
    mockUseItemList.mockReturnValue({
      data: {
        data: [
          {
            id: 'item-1',
            name: 'Kettlebell Pro',
            category: 'strength',
            brand: 'FitCommerce',
            condition: 'new',
            billing_model: 'one_time',
            unit_price: 99,
            refundable_deposit: 50,
            status: 'published',
          },
        ],
        pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
      },
      isLoading: false,
      error: { code: 'NETWORK_OFFLINE', status: 0, message: 'offline' },
      dataUpdatedAt: Date.UTC(2026, 3, 11, 10, 30, 0),
    });

    render(
      <MemoryRouter initialEntries={['/catalog']}>
        <CatalogPage />
      </MemoryRouter>,
    );

    expect(screen.getByText(/offline mode is active\. showing cached data/i)).toBeInTheDocument();
    expect(screen.getByText('Kettlebell Pro')).toBeInTheDocument();
  });

  it('shows a first-load offline notice when no cached rows exist', () => {
    mockUseItemList.mockReturnValue({
      data: {
        data: [],
        pagination: { page: 1, page_size: 20, total_count: 0, total_pages: 0 },
      },
      isLoading: false,
      error: { code: 'NETWORK_OFFLINE', status: 0, message: 'offline' },
      dataUpdatedAt: 0,
    });

    render(
      <MemoryRouter initialEntries={['/catalog']}>
        <CatalogPage />
      </MemoryRouter>,
    );

    expect(screen.getByText(/has no cached data yet/i)).toBeInTheDocument();
  });
});
