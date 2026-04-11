import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import PurchaseOrderDetailPage from '@/routes/PurchaseOrderDetailPage';

vi.mock('@/lib/auth', async () => {
  const actual = await vi.importActual<typeof import('@/lib/auth')>('@/lib/auth');
  return {
    ...actual,
    RequireRole: ({ children, roles }: { children: React.ReactNode; roles: string[] }) =>
      roles.includes('administrator') ? <>{children}</> : null,
  };
});

vi.mock('@/lib/offline', () => ({
  useOfflineStatus: () => ({ isOnline: true, isOffline: false, lastSyncAt: null }),
  OFFLINE_MUTATION_MESSAGE: 'Reconnect to make changes.',
}));

const mockUsePO = vi.fn();
const mockReceiveMutateAsync = vi.fn();

vi.mock('@/lib/hooks/useProcurement', () => ({
  usePO: (id: string | undefined) => mockUsePO(id),
  useApprovePO: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useReceivePO: () => ({ mutateAsync: mockReceiveMutateAsync, isPending: false }),
  useReturnPO: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useVoidPO: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useVarianceList: () => ({ data: [], isLoading: false, error: null }),
  useResolveVariance: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock('@/lib/notifications', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn() }),
}));

const PO_ID = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';
const LINE_ID = 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb';

const MOCK_PO = {
  id: PO_ID,
  supplier_id: 'cccccccc-cccc-cccc-cccc-cccccccccccc',
  status: 'approved',
  total_amount: 112.5,
  approved_by: 'dddddddd-dddd-dddd-dddd-dddddddddddd',
  approved_at: '2026-04-09T12:00:00Z',
  received_at: null,
  created_by: 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
  created_at: '2026-04-09T10:00:00Z',
  version: 3,
  lines: [
    {
      id: LINE_ID,
      purchase_order_id: PO_ID,
      item_id: 'ffffffff-ffff-ffff-ffff-ffffffffffff',
      ordered_quantity: 5,
      ordered_unit_price: 22.5,
      received_quantity: null,
      received_unit_price: null,
    },
  ],
};

function renderPage() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[`/procurement/purchase-orders/${PO_ID}`]}>
        <Routes>
          <Route path="/procurement/purchase-orders/:id" element={<PurchaseOrderDetailPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('PurchaseOrderDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUsePO.mockReturnValue({ data: MOCK_PO, isLoading: false, error: null });
    mockReceiveMutateAsync.mockResolvedValue({});
  });

  it('renders backend-aligned PO line fields', () => {
    renderPage();

    expect(screen.getByText('$112.50')).toBeInTheDocument();
    expect(screen.getByText('$22.50')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^receive$/i })).toBeInTheDocument();
  });

  it('submits receive payload with received_quantity and received_unit_price', async () => {
    renderPage();
    const user = userEvent.setup();

    await user.click(screen.getByRole('button', { name: /^receive$/i }));

    const dialog = await screen.findByRole('dialog');
    const [quantityInput, unitPriceInput] = within(dialog).getAllByRole('spinbutton');

    await user.clear(quantityInput);
    await user.type(quantityInput, '3');
    await user.clear(unitPriceInput);
    await user.type(unitPriceInput, '23.75');

    await user.click(within(dialog).getByRole('button', { name: /confirm receipt/i }));

    await waitFor(() => {
      expect(mockReceiveMutateAsync).toHaveBeenCalledWith({
        id: PO_ID,
        lines: [
          {
            po_line_id: LINE_ID,
            received_quantity: 3,
            received_unit_price: 23.75,
          },
        ],
      });
    });
  });
});
