import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, act, waitFor } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { AuthProvider, useAuth, ProtectedRoute, RequireRole } from '@/lib/auth';

// All of auth.tsx's API calls flow through apiClient. We replace its get/post
// methods so we can control session checks, logins, logouts, and captcha.
const mockGet = vi.fn();
const mockPost = vi.fn();
vi.mock('@/lib/api-client', async () => {
  const actual = await vi.importActual<typeof import('@/lib/api-client')>('@/lib/api-client');
  return {
    ...actual,
    apiClient: {
      get: (...args: unknown[]) => mockGet(...args),
      post: (...args: unknown[]) => mockPost(...args),
      put: vi.fn(),
      delete: vi.fn(),
    },
  };
});

const MOCK_USER = {
  id: 'u1',
  email: 't@t.com',
  display_name: 'T User',
  role: 'administrator',
};

// Session expiries set far in the future so the snapshot is treated as valid.
const MOCK_SESSION = {
  id: 's1',
  user_id: 'u1',
  absolute_expires_at: '2099-01-01T00:00:00Z',
  idle_expires_at: '2099-01-01T00:00:00Z',
};

function WhoAmI() {
  const { user, isAuthenticated, isLoading, captchaState, lockoutState } = useAuth();
  return (
    <div>
      <div data-testid="loading">{String(isLoading)}</div>
      <div data-testid="authed">{String(isAuthenticated)}</div>
      <div data-testid="user">{user?.email ?? '-'}</div>
      <div data-testid="captcha">{captchaState?.challengeId ?? '-'}</div>
      <div data-testid="lockout">{lockoutState?.lockedUntil ?? '-'}</div>
    </div>
  );
}

function LoginHarness() {
  const { login, logout, verifyCaptcha, clearCaptchaState, refreshSession } = useAuth();
  return (
    <div>
      <button onClick={() => login('a@b.com', 'pw').catch(() => {})} data-testid="login">login</button>
      <button onClick={() => logout().catch(() => {})} data-testid="logout">logout</button>
      <button onClick={() => verifyCaptcha('cid', 'ans')} data-testid="captcha-verify">cv</button>
      <button onClick={() => clearCaptchaState()} data-testid="clear-captcha">cc</button>
      <button onClick={() => refreshSession()} data-testid="refresh">r</button>
    </div>
  );
}

function renderWith(children: React.ReactNode) {
  return render(
    <MemoryRouter>
      <AuthProvider>
        <WhoAmI />
        <LoginHarness />
        {children}
      </AuthProvider>
    </MemoryRouter>,
  );
}

describe('AuthProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
  });

  afterEach(() => {
    window.localStorage.clear();
  });

  it('checkSession on mount: sets user/session on success and writes snapshot', async () => {
    mockGet.mockResolvedValueOnce({ data: { user: MOCK_USER, session: MOCK_SESSION } });
    renderWith(null);
    await waitFor(() => expect(screen.getByTestId('authed').textContent).toBe('true'));
    expect(screen.getByTestId('user').textContent).toBe('t@t.com');
    expect(window.localStorage.getItem('fitcommerce.auth.snapshot')).not.toBeNull();
  });

  it('checkSession failure clears auth state', async () => {
    mockGet.mockRejectedValueOnce({ status: 401, code: 'UNAUTH', message: 'no' });
    renderWith(null);
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('false'));
    expect(screen.getByTestId('authed').textContent).toBe('false');
  });

  it('checkSession offline error falls back to valid local snapshot', async () => {
    window.localStorage.setItem(
      'fitcommerce.auth.snapshot',
      JSON.stringify({ user: MOCK_USER, session: MOCK_SESSION }),
    );
    mockGet.mockRejectedValueOnce({ code: 'NETWORK_OFFLINE', status: 0, message: 'off' });
    renderWith(null);
    await waitFor(() => expect(screen.getByTestId('authed').textContent).toBe('true'));
    expect(screen.getByTestId('user').textContent).toBe('t@t.com');
  });

  it('expired local snapshot is cleared and not used on offline checkSession', async () => {
    window.localStorage.setItem(
      'fitcommerce.auth.snapshot',
      JSON.stringify({
        user: MOCK_USER,
        session: {
          ...MOCK_SESSION,
          absolute_expires_at: '2000-01-01T00:00:00Z',
          idle_expires_at: '2000-01-01T00:00:00Z',
        },
      }),
    );
    mockGet.mockRejectedValueOnce({ code: 'NETWORK_OFFLINE', status: 0, message: 'off' });
    renderWith(null);
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('false'));
    // Snapshot expired: auth stays false.
    expect(screen.getByTestId('authed').textContent).toBe('false');
    expect(window.localStorage.getItem('fitcommerce.auth.snapshot')).toBeNull();
  });

  it('login success sets state and writes snapshot', async () => {
    mockGet.mockRejectedValueOnce({ status: 401, code: 'UNAUTH', message: 'no' });
    mockPost.mockResolvedValueOnce({ data: { user: MOCK_USER, session: MOCK_SESSION } });
    renderWith(null);
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('false'));

    await act(async () => {
      screen.getByTestId('login').click();
    });
    await waitFor(() => expect(screen.getByTestId('authed').textContent).toBe('true'));
    expect(window.localStorage.getItem('fitcommerce.auth.snapshot')).not.toBeNull();
  });

  it('login 423 sets lockoutState', async () => {
    mockGet.mockRejectedValueOnce({ status: 401, code: 'UNAUTH', message: 'no' });
    mockPost.mockRejectedValueOnce({
      status: 423,
      code: 'ACCOUNT_LOCKED',
      message: 'locked',
      details: [{ field: 'locked_until', message: '2026-05-01T00:00:00Z' }],
    });
    renderWith(null);
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('false'));

    await act(async () => {
      screen.getByTestId('login').click();
    });
    await waitFor(() =>
      expect(screen.getByTestId('lockout').textContent).toBe('2026-05-01T00:00:00Z'),
    );
  });

  it('login 403 CAPTCHA_REQUIRED sets captchaState', async () => {
    mockGet.mockRejectedValueOnce({ status: 401, code: 'UNAUTH', message: 'no' });
    mockPost.mockRejectedValueOnce({
      status: 403,
      code: 'CAPTCHA_REQUIRED',
      message: 'captcha',
      details: [
        { field: 'challenge_id', message: 'chal-1' },
        { field: 'challenge_data', message: 'data-x' },
      ],
    });
    renderWith(null);
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('false'));

    await act(async () => {
      screen.getByTestId('login').click();
    });
    await waitFor(() => expect(screen.getByTestId('captcha').textContent).toBe('chal-1'));
  });

  it('logout clears state even when the API call fails', async () => {
    mockGet.mockResolvedValueOnce({ data: { user: MOCK_USER, session: MOCK_SESSION } });
    mockPost.mockRejectedValueOnce(new Error('boom')); // logout itself fails
    renderWith(null);
    await waitFor(() => expect(screen.getByTestId('authed').textContent).toBe('true'));

    await act(async () => {
      screen.getByTestId('logout').click();
    });
    await waitFor(() => expect(screen.getByTestId('authed').textContent).toBe('false'));
    expect(window.localStorage.getItem('fitcommerce.auth.snapshot')).toBeNull();
  });

  it('verifyCaptcha clears captchaState on success', async () => {
    mockGet.mockRejectedValueOnce({ status: 401, code: 'UNAUTH', message: 'no' });
    mockPost.mockRejectedValueOnce({
      status: 403,
      code: 'CAPTCHA_REQUIRED',
      message: 'captcha',
      details: [
        { field: 'challenge_id', message: 'chal-1' },
        { field: 'challenge_data', message: 'data-x' },
      ],
    });
    renderWith(null);
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('false'));

    await act(async () => {
      screen.getByTestId('login').click();
    });
    await waitFor(() => expect(screen.getByTestId('captcha').textContent).toBe('chal-1'));

    mockPost.mockResolvedValueOnce({ ok: true });
    await act(async () => {
      screen.getByTestId('captcha-verify').click();
    });
    await waitFor(() => expect(screen.getByTestId('captcha').textContent).toBe('-'));
  });

  it('clearCaptchaState resets captcha and lockout', async () => {
    mockGet.mockRejectedValueOnce({ status: 401, code: 'UNAUTH', message: 'no' });
    mockPost.mockRejectedValueOnce({
      status: 423,
      code: 'ACCOUNT_LOCKED',
      message: 'locked',
      details: [{ field: 'locked_until', message: '2026-05-01T00:00:00Z' }],
    });
    renderWith(null);
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('false'));

    await act(async () => {
      screen.getByTestId('login').click();
    });
    await waitFor(() => expect(screen.getByTestId('lockout').textContent).not.toBe('-'));

    await act(async () => {
      screen.getByTestId('clear-captcha').click();
    });
    expect(screen.getByTestId('lockout').textContent).toBe('-');
  });

  it('refreshSession re-runs checkSession', async () => {
    mockGet
      .mockRejectedValueOnce({ status: 401, code: 'UNAUTH', message: 'no' })
      .mockResolvedValueOnce({ data: { user: MOCK_USER, session: MOCK_SESSION } });
    renderWith(null);
    await waitFor(() => expect(screen.getByTestId('authed').textContent).toBe('false'));

    await act(async () => {
      screen.getByTestId('refresh').click();
    });
    await waitFor(() => expect(screen.getByTestId('authed').textContent).toBe('true'));
  });

  it('auth:session-expired event clears auth state', async () => {
    mockGet.mockResolvedValueOnce({ data: { user: MOCK_USER, session: MOCK_SESSION } });
    renderWith(null);
    await waitFor(() => expect(screen.getByTestId('authed').textContent).toBe('true'));

    await act(async () => {
      window.dispatchEvent(new Event('auth:session-expired'));
    });
    await waitFor(() => expect(screen.getByTestId('authed').textContent).toBe('false'));
  });

  it('ignores malformed snapshot json', async () => {
    window.localStorage.setItem('fitcommerce.auth.snapshot', 'not-json');
    mockGet.mockRejectedValueOnce({ code: 'NETWORK_OFFLINE', status: 0, message: 'off' });
    renderWith(null);
    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('false'));
    expect(screen.getByTestId('authed').textContent).toBe('false');
    expect(window.localStorage.getItem('fitcommerce.auth.snapshot')).toBeNull();
  });
});

describe('useAuth outside provider', () => {
  it('throws when rendered without AuthProvider', () => {
    // Suppress React's console.error for the thrown render.
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
    expect(() =>
      render(<WhoAmI />),
    ).toThrow(/useAuth must be used within an AuthProvider/);
    spy.mockRestore();
  });
});

describe('ProtectedRoute', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
  });

  it('redirects to /login when unauthenticated', async () => {
    mockGet.mockRejectedValueOnce({ status: 401, code: 'UNAUTH', message: 'no' });
    render(
      <MemoryRouter initialEntries={['/private']}>
        <AuthProvider>
          <Routes>
            <Route
              path="/private"
              element={
                <ProtectedRoute>
                  <div data-testid="private">private</div>
                </ProtectedRoute>
              }
            />
            <Route path="/login" element={<div data-testid="login-page">login</div>} />
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByTestId('login-page')).toBeInTheDocument());
  });

  it('renders children when authenticated and role allowed', async () => {
    mockGet.mockResolvedValueOnce({ data: { user: MOCK_USER, session: MOCK_SESSION } });
    render(
      <MemoryRouter initialEntries={['/private']}>
        <AuthProvider>
          <Routes>
            <Route
              path="/private"
              element={
                <ProtectedRoute allowedRoles={['administrator']}>
                  <div data-testid="private">private</div>
                </ProtectedRoute>
              }
            />
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByTestId('private')).toBeInTheDocument());
  });

  it('redirects to /dashboard when role is not in allowedRoles', async () => {
    mockGet.mockResolvedValueOnce({ data: { user: MOCK_USER, session: MOCK_SESSION } });
    render(
      <MemoryRouter initialEntries={['/private']}>
        <AuthProvider>
          <Routes>
            <Route
              path="/private"
              element={
                <ProtectedRoute allowedRoles={['member']}>
                  <div data-testid="private">private</div>
                </ProtectedRoute>
              }
            />
            <Route path="/dashboard" element={<div data-testid="dashboard">dash</div>} />
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByTestId('dashboard')).toBeInTheDocument());
  });
});

describe('RequireRole', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
  });

  it('hides children when role is not allowed; shows fallback', async () => {
    mockGet.mockResolvedValueOnce({ data: { user: MOCK_USER, session: MOCK_SESSION } });
    render(
      <MemoryRouter>
        <AuthProvider>
          <RequireRole roles={['member']} fallback={<span data-testid="fallback">nope</span>}>
            <span data-testid="gated">gated</span>
          </RequireRole>
        </AuthProvider>
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByTestId('fallback')).toBeInTheDocument());
    expect(screen.queryByTestId('gated')).not.toBeInTheDocument();
  });

  it('renders children when role is allowed', async () => {
    mockGet.mockResolvedValueOnce({ data: { user: MOCK_USER, session: MOCK_SESSION } });
    render(
      <MemoryRouter>
        <AuthProvider>
          <RequireRole roles={['administrator']}>
            <span data-testid="gated">gated</span>
          </RequireRole>
        </AuthProvider>
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByTestId('gated')).toBeInTheDocument());
  });
});
