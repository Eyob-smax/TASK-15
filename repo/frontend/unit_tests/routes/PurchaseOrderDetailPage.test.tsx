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

const mockOffline = vi.fn();
vi.mock('@/lib/offline', () => ({
  useOfflineStatus: () => mockOffline(),
  OFFLINE_MUTATION_MESSAGE: 'Reconnect to make changes.',
}));

const mockUsePO = vi.fn();
const mockApprove = vi.fn();
const mockReceive = vi.fn();
const mockReturn = vi.fn();
const mockVoid = vi.fn();
const mockVarianceList = vi.fn();
const mockResolveVariance = vi.fn();

vi.mock('@/lib/hooks/useProcurement', () => ({
  usePO: (id: string | undefined) => mockUsePO(id),
  useApprovePO: () => ({ mutateAsync: mockApprove, isPending: false }),
  useReceivePO: () => ({ mutateAsync: mockReceive, isPending: false }),
  useReturnPO: () => ({ mutateAsync: mockReturn, isPending: false }),
  useVoidPO: () => ({ mutateAsync: mockVoid, isPending: false }),
  useVarianceList: () => mockVarianceList(),
  useResolveVariance: () => ({ mutateAsync: mockResolveVariance, isPending: false }),
}));

vi.mock('@/lib/notifications', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn(), info: vi.fn() }),
}));

const PO_ID = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';
const LINE_ID = 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb';
const VAR_ID = 'dddddddd-dddd-dddd-dddd-dddddddddddd';

const BASE_PO = {
  id: PO_ID,
  supplier_id: 'cccccccc-cccc-cccc-cccc-cccccccccccc',
  status: 'approved',
  total_amount: 112.5,
  approved_by: 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
  approved_at: '2026-04-09T12:00:00Z',
  received_at: null,
  created_by: 'ffffffff-ffff-ffff-ffff-ffffffffffff',
  created_at: '2026-04-09T10:00:00Z',
  version: 3,
  lines: [
    {
      id: LINE_ID,
      purchase_order_id: PO_ID,
      item_id: '11111111-1111-1111-1111-111111111111',
      ordered_quantity: 5,
      ordered_unit_price: 22.5,
      received_quantity: null,
      received_unit_price: null,
    },
  ],
};

const OPEN_VARIANCE = {
  id: VAR_ID,
  po_line_id: LINE_ID,
  type: 'quantity',
  expected_value: 5,
  actual_value: 4,
  status: 'open',
  is_overdue: false,
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
    mockOffline.mockReturnValue({ isOnline: true, isOffline: false, lastSyncAt: null });
    mockUsePO.mockReturnValue({ data: BASE_PO, isLoading: false, error: null, dataUpdatedAt: Date.now() });
    mockVarianceList.mockReturnValue({ data: [], isLoading: false, error: null });
  });

  it('renders backend-aligned PO line fields', () => {
    renderPage();
    expect(screen.getByText('$112.50')).toBeInTheDocument();
    expect(screen.getByText('$22.50')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^receive$/i })).toBeInTheDocument();
  });

  it('renders skeleton while loading', () => {
    mockUsePO.mockReturnValue({ data: undefined, isLoading: true, error: null, dataUpdatedAt: 0 });
    const { container } = renderPage();
    expect(container.querySelector('.MuiSkeleton-root')).not.toBeNull();
  });

  it('renders error alert when PO fails to load', () => {
    mockUsePO.mockReturnValue({ data: undefined, isLoading: false, error: null, dataUpdatedAt: 0 });
    renderPage();
    expect(screen.getByText(/failed to load purchase order/i)).toBeInTheDocument();
  });

  it('shows Approve button when status is created; calls approve mutation', async () => {
    mockUsePO.mockReturnValue({
      data: { ...BASE_PO, status: 'created' },
      isLoading: false,
      error: null,
      dataUpdatedAt: Date.now(),
    });
    mockApprove.mockResolvedValue({});
    renderPage();
    const user = userEvent.setup();

    const approveBtn = screen.getByRole('button', { name: /^approve$/i });
    await user.click(approveBtn);
    await waitFor(() => expect(mockApprove).toHaveBeenCalledWith(PO_ID));
  });

  it('shows Return button when status is received; confirms return', async () => {
    mockUsePO.mockReturnValue({
      data: { ...BASE_PO, status: 'received', received_at: '2026-04-10T00:00:00Z' },
      isLoading: false,
      error: null,
      dataUpdatedAt: Date.now(),
    });
    mockReturn.mockResolvedValue({});
    renderPage();
    const user = userEvent.setup();

    await user.click(screen.getByRole('button', { name: /^return$/i }));
    const dialog = await screen.findByRole('dialog');
    await user.click(within(dialog).getByRole('button', { name: /^return$/i }));
    await waitFor(() => expect(mockReturn).toHaveBeenCalledWith(PO_ID));
  });

  it('shows Void button on non-terminal PO; confirms void', async () => {
    mockVoid.mockResolvedValue({});
    renderPage();
    const user = userEvent.setup();

    const voidBtn = screen.getByRole('button', { name: /^void$/i });
    await user.click(voidBtn);
    const dialog = await screen.findByRole('dialog');
    await user.click(within(dialog).getByRole('button', { name: /^void$/i }));
    await waitFor(() => expect(mockVoid).toHaveBeenCalledWith(PO_ID));
  });

  it('hides Void button on terminal status', () => {
    mockUsePO.mockReturnValue({
      data: { ...BASE_PO, status: 'voided' },
      isLoading: false,
      error: null,
      dataUpdatedAt: Date.now(),
    });
    renderPage();
    expect(screen.queryByRole('button', { name: /^void$/i })).not.toBeInTheDocument();
  });

  it('renders "No lines found" when lines array is empty', () => {
    mockUsePO.mockReturnValue({
      data: { ...BASE_PO, lines: [] },
      isLoading: false,
      error: null,
      dataUpdatedAt: Date.now(),
    });
    renderPage();
    expect(screen.getByText(/no lines found/i)).toBeInTheDocument();
  });

  it('renders variance rows when variance list contains a matching po_line_id', () => {
    mockVarianceList.mockReturnValue({
      data: [OPEN_VARIANCE],
      isLoading: false,
      error: null,
    });
    renderPage();
    expect(screen.getByText('quantity')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^resolve$/i })).toBeInTheDocument();
  });

  it('shows Overdue chip for overdue variance', () => {
    mockVarianceList.mockReturnValue({
      data: [{ ...OPEN_VARIANCE, is_overdue: true }],
      isLoading: false,
      error: null,
    });
    renderPage();
    expect(screen.getAllByText(/overdue/i).length).toBeGreaterThan(0);
  });

  it('accepts variance list as a bare array (non-paginated)', () => {
    // Some hooks return an array directly instead of { data: [...] }.
    mockVarianceList.mockReturnValue({ data: [OPEN_VARIANCE] });
    renderPage();
    expect(screen.getByText('quantity')).toBeInTheDocument();
  });

  it('shows "No variances recorded" when none match', () => {
    mockVarianceList.mockReturnValue({ data: [], isLoading: false, error: null });
    renderPage();
    expect(screen.getByText(/no variances recorded/i)).toBeInTheDocument();
  });

  it('opens resolve variance dialog; validates and calls resolveMutation for adjustment', async () => {
    mockVarianceList.mockReturnValue({ data: [OPEN_VARIANCE] });
    mockResolveVariance.mockResolvedValue({});
    renderPage();
    const user = userEvent.setup();

    await user.click(screen.getByRole('button', { name: /^resolve$/i }));
    const dialog = await screen.findByRole('dialog');

    const resolveBtn = within(dialog).getByRole('button', { name: /^resolve$/i });
    // Both quantity change and notes empty — button disabled.
    expect(resolveBtn).toBeDisabled();

    const qtyInput = within(dialog).getByLabelText(/quantity change/i);
    await user.type(qtyInput, '1');
    expect(resolveBtn).toBeDisabled(); // notes still empty

    const notesInput = within(dialog).getByLabelText(/resolution notes/i);
    await user.type(notesInput, 'partial ship adjustment');
    expect(resolveBtn).toBeEnabled();

    await user.click(resolveBtn);
    await waitFor(() =>
      expect(mockResolveVariance).toHaveBeenCalledWith({
        id: VAR_ID,
        action: 'adjustment',
        resolution_notes: 'partial ship adjustment',
        quantity_change: 1,
      }),
    );
  });

  it('resolve dialog hides Quantity Change field when action is return', async () => {
    mockVarianceList.mockReturnValue({ data: [OPEN_VARIANCE] });
    mockResolveVariance.mockResolvedValue({});
    renderPage();
    const user = userEvent.setup();

    await user.click(screen.getByRole('button', { name: /^resolve$/i }));
    const dialog = await screen.findByRole('dialog');

    const select = within(dialog).getByLabelText(/resolution action/i);
    await user.click(select);
    const returnOption = await screen.findByRole('option', { name: /return/i });
    await user.click(returnOption);

    // Quantity field should no longer be rendered.
    expect(within(dialog).queryByLabelText(/quantity change/i)).not.toBeInTheDocument();

    const notesInput = within(dialog).getByLabelText(/resolution notes/i);
    await user.type(notesInput, 'return to supplier');
    await user.click(within(dialog).getByRole('button', { name: /^resolve$/i }));
    await waitFor(() =>
      expect(mockResolveVariance).toHaveBeenCalledWith({
        id: VAR_ID,
        action: 'return',
        resolution_notes: 'return to supplier',
        quantity_change: undefined,
      }),
    );
  });

  it('shows sync warning alert when PO query has stale error', () => {
    mockUsePO.mockReturnValue({
      data: BASE_PO,
      isLoading: false,
      error: new Error('sync error'),
      dataUpdatedAt: Date.now(),
    });
    renderPage();
    expect(screen.getByText(/purchase order sync is temporarily unavailable/i)).toBeInTheDocument();
  });

  it('receive dialog shows offline notice when offline', async () => {
    mockOffline.mockReturnValue({ isOnline: false, isOffline: true, lastSyncAt: null });
    renderPage();
    const user = userEvent.setup();

    await user.click(screen.getByRole('button', { name: /^receive$/i }));
    const dialog = await screen.findByRole('dialog');
    expect(within(dialog).getByText(/offline — this action will be queued/i)).toBeInTheDocument();
  });

  it('submits receive payload with received_quantity and received_unit_price', async () => {
    mockReceive.mockResolvedValue({});
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
      expect(mockReceive).toHaveBeenCalledWith({
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
