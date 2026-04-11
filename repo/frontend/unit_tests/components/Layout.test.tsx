import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { Layout } from '@/components/Layout';

// Mock useAuth to control the logged-in user
const mockUseAuth = vi.fn();
const mockUseOfflineStatus = vi.fn();
vi.mock('@/lib/auth', async () => {
  const actual = await vi.importActual<typeof import('@/lib/auth')>('@/lib/auth');
  return {
    ...actual,
    useAuth: () => mockUseAuth(),
  };
});

vi.mock('@/lib/offline', () => ({
  useOfflineStatus: () => mockUseOfflineStatus(),
}));

function renderLayout(
  role: string,
  offlineState: { isOnline: boolean; isOffline: boolean; lastSyncAt: number | null } = {
    isOnline: true,
    isOffline: false,
    lastSyncAt: null,
  },
) {
  mockUseAuth.mockReturnValue({
    user: { id: '1', display_name: 'Test User', email: 'test@example.com', role },
    isAuthenticated: true,
    isLoading: false,
    logout: vi.fn(),
    captchaState: null,
    lockoutState: null,
  });
  mockUseOfflineStatus.mockReturnValue(offlineState);

  return render(
    <MemoryRouter initialEntries={['/dashboard']}>
      <Routes>
        <Route path="/*" element={<Layout />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe('Layout / NavSidebar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders app brand name', () => {
    renderLayout('administrator');
    expect(screen.getByText('FitCommerce')).toBeInTheDocument();
  });

  it('shows user display name in header', () => {
    renderLayout('administrator');
    expect(screen.getByText('Test User')).toBeInTheDocument();
  });

  it('administrator sees all nav items', () => {
    renderLayout('administrator');
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Catalog')).toBeInTheDocument();
    expect(screen.getByText('Inventory')).toBeInTheDocument();
    expect(screen.getByText('Group Buys')).toBeInTheDocument();
    expect(screen.getByText('Orders')).toBeInTheDocument();
    expect(screen.getByText('Procurement')).toBeInTheDocument();
    expect(screen.getByText('Reports')).toBeInTheDocument();
    expect(screen.getByText('Admin')).toBeInTheDocument();
  });

  it('member only sees allowed nav items', () => {
    renderLayout('member');
    expect(screen.getByText('Catalog')).toBeInTheDocument();
    expect(screen.getByText('Group Buys')).toBeInTheDocument();
    expect(screen.getByText('Orders')).toBeInTheDocument();
    // Member should NOT see inventory, procurement, admin
    expect(screen.queryByText('Inventory')).not.toBeInTheDocument();
    expect(screen.queryByText('Procurement')).not.toBeInTheDocument();
    expect(screen.queryByText('Admin')).not.toBeInTheDocument();
  });

  it('procurement_specialist sees catalog, procurement, reports', () => {
    renderLayout('procurement_specialist');
    expect(screen.getByText('Catalog')).toBeInTheDocument();
    expect(screen.getByText('Procurement')).toBeInTheDocument();
    expect(screen.getByText('Reports')).toBeInTheDocument();
    // Should NOT see inventory or admin
    expect(screen.queryByText('Inventory')).not.toBeInTheDocument();
    expect(screen.queryByText('Admin')).not.toBeInTheDocument();
  });

  it('coach sees dashboard, group buys, orders, reports', () => {
    renderLayout('coach');
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Group Buys')).toBeInTheDocument();
    expect(screen.getByText('Orders')).toBeInTheDocument();
    expect(screen.getByText('Reports')).toBeInTheDocument();
    expect(screen.queryByText('Inventory')).not.toBeInTheDocument();
    expect(screen.queryByText('Admin')).not.toBeInTheDocument();
  });

  it('shows the offline banner when the app is offline', () => {
    renderLayout('administrator', {
      isOnline: false,
      isOffline: true,
      lastSyncAt: Date.UTC(2026, 3, 11, 10, 30, 0),
    });
    expect(screen.getByText(/offline mode is active/i)).toBeInTheDocument();
  });
});
