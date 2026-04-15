import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import MembersPage from '@/routes/MembersPage';

const mockGet = vi.fn();
vi.mock('@/lib/api-client', () => ({
  apiClient: {
    get: (...args: unknown[]) => mockGet(...args),
  },
}));

const MOCK_MEMBER = {
  id: 'mmmmmmmm-mmmm-mmmm-mmmm-mmmmmmmmmmmm',
  user_id: 'uuuuuuuu-uuuu-uuuu-uuuu-uuuuuuuuuuuu',
  location_id: 'llllllll-llll-llll-llll-llllllllllll',
  membership_status: 'active',
  joined_at: '2026-01-01T00:00:00Z',
  renewal_date: '2027-01-01T00:00:00Z',
};

function renderPage() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <MembersPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('MembersPage', () => {
  beforeEach(() => vi.clearAllMocks());

  it('renders page title', () => {
    mockGet.mockResolvedValue({
      data: [],
      pagination: { page: 1, page_size: 20, total_count: 0, total_pages: 0 },
    });
    renderPage();
    expect(screen.getByRole('heading', { name: /members/i })).toBeInTheDocument();
  });

  it('calls /members with pagination', async () => {
    mockGet.mockResolvedValue({
      data: [],
      pagination: { page: 1, page_size: 20, total_count: 0, total_pages: 0 },
    });
    renderPage();
    await waitFor(() =>
      expect(mockGet).toHaveBeenCalledWith('/members', { page: 1, page_size: 20 }),
    );
  });

  it('renders member row with status chip', async () => {
    mockGet.mockResolvedValue({
      data: [MOCK_MEMBER],
      pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
    });
    renderPage();
    await waitFor(() => expect(screen.getAllByText(/active/i).length).toBeGreaterThan(0));
  });

  it('renders em dash when renewal_date is absent', async () => {
    mockGet.mockResolvedValue({
      data: [{ ...MOCK_MEMBER, renewal_date: null }],
      pagination: { page: 1, page_size: 20, total_count: 1, total_pages: 1 },
    });
    renderPage();
    await waitFor(() => expect(screen.getByText('—')).toBeInTheDocument());
  });
});
