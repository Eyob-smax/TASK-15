import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import CatalogFormPage from '@/routes/CatalogFormPage';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: () => mockNavigate };
});

const mockUseAuth = vi.fn();
vi.mock('@/lib/auth', async () => {
  const actual = await vi.importActual<typeof import('@/lib/auth')>('@/lib/auth');
  return { ...actual, useAuth: () => mockUseAuth() };
});

vi.mock('@/lib/offline', () => ({
  useOfflineStatus: () => ({ isOnline: true, isOffline: false, lastSyncAt: null }),
  OFFLINE_MUTATION_MESSAGE: 'Reconnect to make changes.',
}));

const mockUseItem = vi.fn();
const mockCreateMutateAsync = vi.fn();
const mockUpdateMutateAsync = vi.fn();

vi.mock('@/lib/hooks/useItems', () => ({
  useItem: (id: string | undefined) => mockUseItem(id),
  useCreateItem: () => ({ mutateAsync: mockCreateMutateAsync, isPending: false }),
  useUpdateItem: () => ({ mutateAsync: mockUpdateMutateAsync, isPending: false }),
}));

vi.mock('@/lib/notifications', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn() }),
}));

const MOCK_ITEM = {
  id: 'iiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii',
  name: 'Resistance Band',
  description: 'Latex resistance band',
  category: 'Fitness',
  brand: 'FitPro',
  sku: 'RB-001',
  condition: 'new',
  billing_model: 'one_time',
  unit_price: 29.99,
  refundable_deposit: 50,
  quantity: 100,
  location_id: null,
  status: 'draft',
  version: 7,
  availability_windows: [
    {
      id: 'aw-1',
      start_time: '2026-04-15T09:00:00Z',
      end_time: '2026-04-15T17:00:00Z',
    },
  ],
  blackout_windows: [],
  created_at: '2026-04-01T00:00:00Z',
  updated_at: '2026-04-01T00:00:00Z',
};

function renderCreatePage() {
  mockUseAuth.mockReturnValue({
    user: { id: 'user-1', role: 'administrator', display_name: 'Admin', email: 'admin@test.com' },
    isAuthenticated: true,
  });
  mockUseItem.mockReturnValue({ data: undefined, isLoading: false, error: null });
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={['/catalog/new']}>
        <Routes>
          <Route path="/catalog/new" element={<CatalogFormPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

function renderEditPage() {
  mockUseAuth.mockReturnValue({
    user: { id: 'user-1', role: 'administrator', display_name: 'Admin', email: 'admin@test.com' },
    isAuthenticated: true,
  });
  mockUseItem.mockReturnValue({ data: MOCK_ITEM, isLoading: false, error: null });
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[`/catalog/${MOCK_ITEM.id}/edit`]}>
        <Routes>
          <Route path="/catalog/:id/edit" element={<CatalogFormPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('CatalogFormPage create mode', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateMutateAsync.mockResolvedValue({ ...MOCK_ITEM, id: 'new-id' });
  });

  it('renders Create Item heading', () => {
    renderCreatePage();
    expect(screen.getByText(/create item/i)).toBeInTheDocument();
  });

  it('shows validation error when name is missing', async () => {
    renderCreatePage();
    const user = userEvent.setup();

    await user.click(screen.getByRole('button', { name: /save draft/i }));

    await waitFor(() => {
      expect(screen.getByText(/name is required/i)).toBeInTheDocument();
    });
  });

  it('shows validation error when category is missing', async () => {
    renderCreatePage();
    const user = userEvent.setup();

    await user.click(screen.getByRole('button', { name: /save draft/i }));

    await waitFor(() => {
      expect(screen.getByText(/category is required/i)).toBeInTheDocument();
    });
  });

  it('submits form with valid data and calls createItem', async () => {
    renderCreatePage();
    const user = userEvent.setup();

    await user.type(screen.getByLabelText(/^name/i), 'Kettlebell 16kg');
    await user.type(screen.getByLabelText(/category/i), 'Free Weights');
    await user.type(screen.getByLabelText(/brand/i), 'IronGrip');
    await user.click(screen.getByRole('button', { name: /save draft/i }));

    await waitFor(() => {
      expect(mockCreateMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Kettlebell 16kg',
          category: 'Free Weights',
          brand: 'IronGrip',
          unit_price: 0,
          refundable_deposit: 50,
          quantity: 0,
          billing_model: 'one_time',
        }),
      );
    });
  });
});

describe('CatalogFormPage edit mode', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUpdateMutateAsync.mockResolvedValue(MOCK_ITEM);
  });

  it('renders the item-specific edit heading', () => {
    renderEditPage();
    expect(screen.getByText(/edit resistance band/i)).toBeInTheDocument();
  });

  it('pre-fills name field with existing item name', async () => {
    renderEditPage();
    await waitFor(() => {
      const nameInput = screen.getByLabelText(/^name/i) as HTMLInputElement;
      expect(nameInput.value).toBe('Resistance Band');
    });
  });

  it('calls updateItem with a numeric version on save', async () => {
    renderEditPage();
    const user = userEvent.setup();

    await waitFor(() => {
      expect((screen.getByLabelText(/^name/i) as HTMLInputElement).value).toBe('Resistance Band');
    });

    await user.click(screen.getByRole('button', { name: /save changes/i }));

    await waitFor(() => {
      expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
        id: MOCK_ITEM.id,
        body: expect.objectContaining({
          name: 'Resistance Band',
          unit_price: 29.99,
          quantity: 100,
          version: 7,
        }),
      });
    });
  });
});
