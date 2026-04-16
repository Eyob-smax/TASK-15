import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import LandedCostsPage from '@/routes/LandedCostsPage';

const mockGet = vi.fn();
vi.mock('@/lib/api-client', () => ({
  apiClient: {
    get: (...args: unknown[]) => mockGet(...args),
  },
}));

const MOCK_ENTRY = {
  id: 'ee11ee11-ee11-ee11-ee11-ee11ee11ee11',
  item_id: 'iiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii',
  purchase_order_id: 'pppppppp-pppp-pppp-pppp-pppppppppppp',
  period: '2026-Q1',
  cost_component: 'shipping',
  raw_amount: 100,
  allocated_amount: 80.5,
  allocation_method: 'by_quantity',
};

function renderPage() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <LandedCostsPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('LandedCostsPage', () => {
  beforeEach(() => vi.clearAllMocks());

  it('renders title and prompts for item id on first load', () => {
    renderPage();
    expect(screen.getByRole('heading', { name: /landed costs/i })).toBeInTheDocument();
    expect(screen.getByText(/enter an item id to search/i)).toBeInTheDocument();
  });

  it('disables Search button when item id is empty', () => {
    renderPage();
    const btn = screen.getByRole('button', { name: /search/i });
    expect(btn).toBeDisabled();
  });

  it('does not fetch before search is submitted', () => {
    renderPage();
    expect(mockGet).not.toHaveBeenCalled();
  });

  it('fetches with item_id and period after Search click', async () => {
    mockGet.mockResolvedValue({ data: [MOCK_ENTRY] });
    renderPage();

    const user = userEvent.setup();
    const itemField = screen.getByLabelText(/item id/i);
    const periodField = screen.getByLabelText(/period/i);
    fireEvent.change(itemField, { target: { value: MOCK_ENTRY.item_id } });
    fireEvent.change(periodField, { target: { value: '2026-Q1' } });

    await user.click(screen.getByRole('button', { name: /search/i }));

    await waitFor(() =>
      expect(mockGet).toHaveBeenCalledWith('/procurement/landed-costs', {
        item_id: MOCK_ENTRY.item_id,
        period: '2026-Q1',
      }),
    );
  });

  it('omits period when empty', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderPage();

    const user = userEvent.setup();
    const itemField = screen.getByLabelText(/item id/i);
    fireEvent.change(itemField, { target: { value: MOCK_ENTRY.item_id } });

    await user.click(screen.getByRole('button', { name: /search/i }));

    await waitFor(() =>
      expect(mockGet).toHaveBeenCalledWith('/procurement/landed-costs', {
        item_id: MOCK_ENTRY.item_id,
        period: undefined,
      }),
    );
  });

  it('renders a landed-cost row after a successful search', async () => {
    mockGet.mockResolvedValue({ data: [MOCK_ENTRY] });
    renderPage();

    const user = userEvent.setup();
    const itemField = screen.getByLabelText(/item id/i);
    fireEvent.change(itemField, { target: { value: MOCK_ENTRY.item_id } });
    await user.click(screen.getByRole('button', { name: /search/i }));

    await waitFor(() => expect(screen.getByText('shipping')).toBeInTheDocument());
    expect(screen.getByText('$100.00')).toBeInTheDocument();
    expect(screen.getByText('$80.50')).toBeInTheDocument();
    expect(screen.getByText('by_quantity')).toBeInTheDocument();
  });
});
