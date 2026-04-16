import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import CoachesPage from '@/routes/CoachesPage';

const mockGet = vi.fn();
vi.mock('@/lib/api-client', () => ({
  apiClient: {
    get: (...args: unknown[]) => mockGet(...args),
  },
}));

const MOCK_COACH = {
  id: 'cccccccc-cccc-cccc-cccc-cccccccccccc',
  user_id: 'uuuuuuuu-uuuu-uuuu-uuuu-uuuuuuuuuuuu',
  location_id: 'llllllll-llll-llll-llll-llllllllllll',
  specialization: 'strength',
  is_active: true,
  created_at: '2026-03-01T00:00:00Z',
};

function renderPage() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <CoachesPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('CoachesPage', () => {
  beforeEach(() => vi.clearAllMocks());

  it('renders page title', async () => {
    mockGet.mockResolvedValue({
      data: [MOCK_COACH],
      pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
    });
    renderPage();
    expect(screen.getByRole('heading', { name: /coaches/i })).toBeInTheDocument();
  });

  it('calls /coaches with pagination', async () => {
    mockGet.mockResolvedValue({
      data: [],
      pagination: { page: 1, page_size: 20, total_count: 0, total_pages: 0 },
    });
    renderPage();
    await waitFor(() =>
      expect(mockGet).toHaveBeenCalledWith('/coaches', { page: 1, page_size: 20 }),
    );
  });

  it('renders coach row with specialization', async () => {
    mockGet.mockResolvedValue({
      data: [MOCK_COACH],
      pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
    });
    renderPage();
    await waitFor(() => expect(screen.getByText('strength')).toBeInTheDocument());
  });

  it('shows Yes for active coach', async () => {
    mockGet.mockResolvedValue({
      data: [MOCK_COACH],
      pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
    });
    renderPage();
    await waitFor(() => expect(screen.getAllByText('Yes').length).toBeGreaterThan(0));
  });
});
