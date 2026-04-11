import { describe, it, expect, vi, beforeAll, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { AuthProvider, useAuth } from '@/lib/auth';

/**
 * Tests that AuthProvider correctly unwraps the backend envelope format:
 *   { data: { user: User, session: Session } }
 * where Session has idle_expires_at and absolute_expires_at fields.
 */

let originalFetch: typeof fetch;

afterEach(() => {
  globalThis.fetch = originalFetch;
  vi.restoreAllMocks();
});

function AuthDisplay() {
  const { user, session, isAuthenticated, isLoading, login } = useAuth();

  if (isLoading) return <div data-testid="loading">Loading</div>;

  if (!isAuthenticated) {
    return (
      <div>
        <div data-testid="unauthenticated">Not authenticated</div>
        <button onClick={() => login('admin@test.com', 'password123')}>Login</button>
      </div>
    );
  }

  return (
    <div>
      <div data-testid="user-email">{user?.email}</div>
      <div data-testid="user-role">{user?.role}</div>
      <div data-testid="user-display-name">{user?.display_name}</div>
      <div data-testid="session-idle">{session?.idle_expires_at}</div>
      <div data-testid="session-absolute">{session?.absolute_expires_at}</div>
    </div>
  );
}

function renderWithAuth() {
  return render(
    <MemoryRouter>
      <AuthProvider>
        <AuthDisplay />
      </AuthProvider>
    </MemoryRouter>,
  );
}

describe('Auth contract: backend envelope unwrapping', () => {
  beforeAll(() => {
    originalFetch = globalThis.fetch;
  });

  it('checkSession unwraps { data: { user, session } } envelope on mount', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({
        data: {
          user: {
            id: 'user-1',
            email: 'admin@test.com',
            role: 'administrator',
            display_name: 'Admin User',
            status: 'active',
          },
          session: {
            idle_expires_at: '2026-04-10T12:00:00Z',
            absolute_expires_at: '2026-04-10T20:00:00Z',
          },
        },
      }),
    } as Response);

    renderWithAuth();

    await waitFor(() => {
      expect(screen.getByTestId('user-email')).toHaveTextContent('admin@test.com');
    });

    expect(screen.getByTestId('user-role')).toHaveTextContent('administrator');
    expect(screen.getByTestId('user-display-name')).toHaveTextContent('Admin User');
    expect(screen.getByTestId('session-idle')).toHaveTextContent('2026-04-10T12:00:00Z');
    expect(screen.getByTestId('session-absolute')).toHaveTextContent('2026-04-10T20:00:00Z');
  });

  it('login unwraps { data: { user, session } } envelope', async () => {
    // First call: checkSession returns 401 (not logged in)
    // Second call: login returns success envelope
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({
        ok: false,
        status: 401,
        statusText: 'Unauthorized',
        json: async () => ({ error: { code: 'UNAUTHORIZED', message: 'no session' } }),
      } as Response)
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          data: {
            user: {
              id: 'user-2',
              email: 'coach@test.com',
              role: 'coach',
              display_name: 'Coach User',
              status: 'active',
            },
            session: {
              idle_expires_at: '2026-04-10T13:00:00Z',
              absolute_expires_at: '2026-04-10T21:00:00Z',
            },
          },
        }),
      } as Response);

    globalThis.fetch = fetchMock;

    renderWithAuth();

    // Wait for initial session check to complete (will fail with 401)
    await waitFor(() => {
      expect(screen.getByTestId('unauthenticated')).toBeInTheDocument();
    });

    // Click login button
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: /login/i }));

    await waitFor(() => {
      expect(screen.getByTestId('user-email')).toHaveTextContent('coach@test.com');
    });

    expect(screen.getByTestId('user-role')).toHaveTextContent('coach');
    expect(screen.getByTestId('session-idle')).toHaveTextContent('2026-04-10T13:00:00Z');
    expect(screen.getByTestId('session-absolute')).toHaveTextContent('2026-04-10T21:00:00Z');
  });

  it('session fields match backend SessionMetaResponse shape (not legacy fields)', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({
        data: {
          user: {
            id: 'user-3',
            email: 'member@test.com',
            role: 'member',
            display_name: 'Member User',
            status: 'active',
          },
          session: {
            idle_expires_at: '2026-04-10T14:00:00Z',
            absolute_expires_at: '2026-04-10T22:00:00Z',
            // These legacy fields should NOT be present in the type,
            // but even if present in wire data, we verify only the correct fields render.
          },
        },
      }),
    } as Response);

    renderWithAuth();

    await waitFor(() => {
      expect(screen.getByTestId('session-idle')).toHaveTextContent('2026-04-10T14:00:00Z');
    });
    expect(screen.getByTestId('session-absolute')).toHaveTextContent('2026-04-10T22:00:00Z');
  });
});
