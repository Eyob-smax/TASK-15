import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import BiometricPage from '@/routes/BiometricPage';

const mockUseBiometric = vi.fn();
const mockUseEncryptionKeys = vi.fn();
const mockRegisterMutateAsync = vi.fn();
const mockRevokeMutateAsync = vi.fn();
const mockRotateMutateAsync = vi.fn();

vi.mock('@/lib/hooks/useAdmin', () => ({
  useBiometric: (userId: string | undefined) => mockUseBiometric(userId),
  useEncryptionKeys: () => mockUseEncryptionKeys(),
  useRegisterBiometric: () => ({
    mutateAsync: mockRegisterMutateAsync,
    isPending: false,
  }),
  useRevokeBiometric: () => ({
    mutateAsync: mockRevokeMutateAsync,
    isPending: false,
  }),
  useRotateKey: () => ({ mutateAsync: mockRotateMutateAsync, isPending: false }),
}));

vi.mock('@/lib/notifications', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn(), info: vi.fn() }),
}));

function renderPage() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <BiometricPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

// Page has two "User ID" inputs (lookup + register) and one "Template Ref"
// input. All are positional — [0]=lookup user id, [1]=register user id,
// [2]=register template ref.
function getInputs() {
  const userIdInputs = screen.getAllByLabelText(/user id/i, { selector: 'input' });
  const templateInput = screen.getByLabelText(/template ref/i, { selector: 'input' });
  return {
    lookupUserId: userIdInputs[0],
    registerUserId: userIdInputs[1],
    registerTemplateRef: templateInput,
  };
}

describe('BiometricPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseBiometric.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    });
    mockUseEncryptionKeys.mockReturnValue({
      data: [],
      isLoading: false,
      error: null,
    });
  });

  it('renders page title and section headings', () => {
    renderPage();
    expect(screen.getByRole('heading', { name: /biometric management/i })).toBeInTheDocument();
    expect(screen.getByText(/user lookup/i)).toBeInTheDocument();
    expect(screen.getByText(/register biometric/i)).toBeInTheDocument();
    expect(screen.getAllByText(/encryption keys/i).length).toBeGreaterThan(0);
  });

  it('Lookup button is disabled until a user id is typed', () => {
    renderPage();
    const lookup = screen.getByRole('button', { name: /^lookup$/i });
    expect(lookup).toBeDisabled();

    const { lookupUserId } = getInputs();
    fireEvent.change(lookupUserId, { target: { value: 'user-123' } });
    expect(lookup).toBeEnabled();
  });

  it('displays enrollment card when biometric is found', async () => {
    // Mock biometric data BEFORE render so it renders immediately when
    // the user id state is set.
    mockUseBiometric.mockReturnValue({
      data: {
        id: 'b1',
        user_id: 'user-123',
        template_ref: 'tpl-ref-456',
        is_active: true,
        created_at: '2026-03-01T12:00:00Z',
        updated_at: '2026-03-01T12:00:00Z',
      },
      isLoading: false,
      error: null,
    });
    renderPage();
    const user = userEvent.setup();

    const { lookupUserId } = getInputs();
    fireEvent.change(lookupUserId, { target: { value: 'user-123' } });
    await user.click(screen.getByRole('button', { name: /^lookup$/i }));

    await waitFor(() =>
      expect(screen.getByText(/enrollment record/i)).toBeInTheDocument(),
    );
    expect(screen.getByText('tpl-ref-456')).toBeInTheDocument();
    expect(screen.getByText('Active')).toBeInTheDocument();
  });

  it('shows module-disabled alert when biometric lookup returns 501', async () => {
    mockUseBiometric.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('HTTP 501 MODULE_DISABLED'),
    });
    renderPage();
    const user = userEvent.setup();

    const { lookupUserId } = getInputs();
    fireEvent.change(lookupUserId, { target: { value: 'user-123' } });
    await user.click(screen.getByRole('button', { name: /^lookup$/i }));

    await waitFor(() =>
      expect(
        screen.getAllByText(/biometric module is not enabled/i).length,
      ).toBeGreaterThan(0),
    );
  });

  it('Register button is disabled until both fields are filled', () => {
    renderPage();
    const registerBtn = screen.getByRole('button', { name: /^register$/i });
    expect(registerBtn).toBeDisabled();

    const { registerUserId, registerTemplateRef } = getInputs();
    fireEvent.change(registerUserId, { target: { value: 'user-1' } });
    expect(registerBtn).toBeDisabled();

    fireEvent.change(registerTemplateRef, { target: { value: 'tpl-1' } });
    expect(registerBtn).toBeEnabled();
  });

  it('submits register with trimmed values', async () => {
    mockRegisterMutateAsync.mockResolvedValue({});
    renderPage();
    const user = userEvent.setup();

    const { registerUserId, registerTemplateRef } = getInputs();
    fireEvent.change(registerUserId, { target: { value: '  user-1  ' } });
    fireEvent.change(registerTemplateRef, { target: { value: '  tpl-1  ' } });

    await user.click(screen.getByRole('button', { name: /^register$/i }));

    await waitFor(() =>
      expect(mockRegisterMutateAsync).toHaveBeenCalledWith({
        user_id: 'user-1',
        template_ref: 'tpl-1',
      }),
    );
  });

  it('rotate key button triggers rotate mutation', async () => {
    mockRotateMutateAsync.mockResolvedValue({});
    renderPage();
    const user = userEvent.setup();

    await user.click(screen.getByRole('button', { name: /rotate key/i }));
    await waitFor(() => expect(mockRotateMutateAsync).toHaveBeenCalled());
  });

  it('renders encryption keys list when present', () => {
    mockUseEncryptionKeys.mockReturnValue({
      data: [
        { id: 'k1', purpose: 'biometric', is_active: true, created_at: '2026-03-01T00:00:00Z' },
      ],
      isLoading: false,
      error: null,
    });
    renderPage();
    expect(screen.getByText('k1')).toBeInTheDocument();
  });

  it('shows empty-keys message when no keys present', () => {
    mockUseEncryptionKeys.mockReturnValue({
      data: [],
      isLoading: false,
      error: null,
    });
    renderPage();
    expect(screen.getByText(/no encryption keys found/i)).toBeInTheDocument();
  });

  it('shows module-disabled alert in keys section on 501', () => {
    mockUseEncryptionKeys.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('HTTP 501 MODULE_DISABLED'),
    });
    renderPage();
    expect(screen.getAllByText(/biometric module is not enabled/i).length).toBeGreaterThan(0);
  });

  it('opens and cancels the revoke confirm dialog', async () => {
    mockUseBiometric.mockReturnValue({
      data: {
        id: 'b1',
        user_id: 'user-123',
        template_ref: 'tpl-ref-456',
        is_active: true,
        created_at: '2026-03-01T00:00:00Z',
        updated_at: '2026-03-01T00:00:00Z',
      },
      isLoading: false,
      error: null,
    });
    renderPage();
    const user = userEvent.setup();

    const { lookupUserId } = getInputs();
    fireEvent.change(lookupUserId, { target: { value: 'user-123' } });
    await user.click(screen.getByRole('button', { name: /^lookup$/i }));

    const revokeBtn = await screen.findByRole('button', { name: /^revoke$/i });
    await user.click(revokeBtn);

    const dialog = await screen.findByRole('dialog');
    expect(within(dialog).getByText(/revoke biometric/i)).toBeInTheDocument();

    const cancelBtn = within(dialog).getByRole('button', { name: /cancel/i });
    await user.click(cancelBtn);
  });
});
