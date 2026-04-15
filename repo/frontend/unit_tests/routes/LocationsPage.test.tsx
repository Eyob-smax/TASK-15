import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import LocationsPage from '@/routes/LocationsPage';

const mockGet = vi.fn();
vi.mock('@/lib/api-client', () => ({
  apiClient: {
    get: (...args: unknown[]) => mockGet(...args),
  },
}));

const MOCK_LOCATION = {
  id: 'llllllll-llll-llll-llll-llllllllllll',
  name: 'Main Gym',
  address: '1 Fitness Ave',
  timezone: 'UTC',
  is_active: true,
  created_at: '2026-03-01T00:00:00Z',
  updated_at: '2026-03-01T00:00:00Z',
};

function renderPage() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <LocationsPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('LocationsPage', () => {
  beforeEach(() => vi.clearAllMocks());

  it('renders page title', () => {
    mockGet.mockResolvedValue({
      data: [],
      pagination: { page: 1, page_size: 20, total_count: 0, total_pages: 0 },
    });
    renderPage();
    expect(screen.getByRole('heading', { name: /locations/i })).toBeInTheDocument();
  });

  it('calls /locations with pagination', async () => {
    mockGet.mockResolvedValue({
      data: [],
      pagination: { page: 1, page_size: 20, total_count: 0, total_pages: 0 },
    });
    renderPage();
    await waitFor(() =>
      expect(mockGet).toHaveBeenCalledWith('/locations', { page: 1, page_size: 20 }),
    );
  });

  it('renders location row', async () => {
    mockGet.mockResolvedValue({
      data: [MOCK_LOCATION],
      pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
    });
    renderPage();
    await waitFor(() => expect(screen.getByText('Main Gym')).toBeInTheDocument());
    expect(screen.getByText('1 Fitness Ave')).toBeInTheDocument();
    expect(screen.getByText('UTC')).toBeInTheDocument();
  });

  it('shows empty state when no locations', async () => {
    mockGet.mockResolvedValue({
      data: [],
      pagination: { page: 1, page_size: 20, total_count: 0, total_pages: 0 },
    });
    renderPage();
    await waitFor(() => expect(screen.getByText(/no locations found/i)).toBeInTheDocument());
  });
});
