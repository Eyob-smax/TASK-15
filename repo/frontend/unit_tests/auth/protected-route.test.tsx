import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import type { ComponentProps, ReactElement } from 'react';
import { AuthContext, ProtectedRoute } from '@/lib/auth';
import type { User } from '@/lib/types';

type AuthValue = NonNullable<ComponentProps<typeof AuthContext.Provider>['value']>;

function makeUser(role: User['role']): User {
  return {
    id: 'user-1',
    email: 'user@test.com',
    display_name: 'Test User',
    role,
    status: 'active',
    location_id: null,
    failed_login_attempts: 0,
    locked_until: null,
    last_login_at: null,
    password_changed_at: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  };
}

function makeAuthValue(overrides: Partial<AuthValue> = {}): AuthValue {
  return {
    user: null,
    session: null,
    isLoading: false,
    isAuthenticated: false,
    captchaState: null,
    lockoutState: null,
    login: vi.fn(),
    logout: vi.fn(),
    refreshSession: vi.fn(),
    verifyCaptcha: vi.fn(),
    clearCaptchaState: vi.fn(),
    ...overrides,
  };
}

function renderWithRouter(ui: ReactElement, authValue: AuthValue, initialPath = '/protected') {
  return render(
    <AuthContext.Provider value={authValue}>
      <MemoryRouter initialEntries={[initialPath]}>
        <Routes>
          <Route path="/login" element={<div data-testid="login-page">Login</div>} />
          <Route path="/dashboard" element={<div data-testid="dashboard-page">Dashboard</div>} />
          <Route path="/protected" element={ui} />
        </Routes>
      </MemoryRouter>
    </AuthContext.Provider>,
  );
}

describe('ProtectedRoute', () => {
  it('renders children when authenticated', () => {
    renderWithRouter(
      <ProtectedRoute>
        <div data-testid="content">Protected Content</div>
      </ProtectedRoute>,
      makeAuthValue({
        isAuthenticated: true,
        isLoading: false,
        user: makeUser('administrator'),
      }),
    );

    expect(screen.getByTestId('content')).toBeInTheDocument();
  });

  it('renders null (loading state) while loading', () => {
    const { container } = renderWithRouter(
      <ProtectedRoute>
        <div data-testid="content">Protected Content</div>
      </ProtectedRoute>,
      makeAuthValue({
        isAuthenticated: false,
        isLoading: true,
        user: null,
      }),
    );

    expect(container.firstChild).toBeNull();
    expect(screen.queryByTestId('content')).not.toBeInTheDocument();
  });

  it('redirects to /login when unauthenticated', () => {
    renderWithRouter(
      <ProtectedRoute>
        <div data-testid="content">Protected Content</div>
      </ProtectedRoute>,
      makeAuthValue({
        isAuthenticated: false,
        isLoading: false,
        user: null,
      }),
    );

    expect(screen.queryByTestId('content')).not.toBeInTheDocument();
    expect(screen.getByTestId('login-page')).toBeInTheDocument();
  });

  it('redirects to /dashboard when role is not allowed', () => {
    renderWithRouter(
      <ProtectedRoute allowedRoles={['administrator']}>
        <div data-testid="admin-content">Admin Only</div>
      </ProtectedRoute>,
      makeAuthValue({
        isAuthenticated: true,
        isLoading: false,
        user: makeUser('member'),
      }),
    );

    expect(screen.queryByTestId('admin-content')).not.toBeInTheDocument();
    expect(screen.getByTestId('dashboard-page')).toBeInTheDocument();
  });
});
