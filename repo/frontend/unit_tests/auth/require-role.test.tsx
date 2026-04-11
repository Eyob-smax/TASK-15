import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import type { ComponentProps } from 'react';
import { AuthContext, RequireRole } from '@/lib/auth';
import type { User, UserRole } from '@/lib/types';

type AuthValue = NonNullable<ComponentProps<typeof AuthContext.Provider>['value']>;

function makeUser(role: UserRole): User {
  return {
    id: 'u1',
    email: 'test@example.com',
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

function renderWithAuth(ui: React.ReactNode, authValue: AuthValue) {
  return render(
    <AuthContext.Provider value={authValue}>
      {ui}
    </AuthContext.Provider>,
  );
}

describe('RequireRole', () => {
  it('renders children when user has the required role', () => {
    renderWithAuth(
      <RequireRole roles={['administrator']}>
        <div data-testid="admin-content">Admin Area</div>
      </RequireRole>,
      makeAuthValue({ user: makeUser('administrator') }),
    );

    expect(screen.getByTestId('admin-content')).toBeInTheDocument();
  });

  it('renders fallback when user does not have the required role', () => {
    renderWithAuth(
      <RequireRole roles={['administrator']} fallback={<div data-testid="fallback">No access</div>}>
        <div data-testid="admin-content">Admin Area</div>
      </RequireRole>,
      makeAuthValue({ user: makeUser('member') }),
    );

    expect(screen.queryByTestId('admin-content')).not.toBeInTheDocument();
    expect(screen.getByTestId('fallback')).toBeInTheDocument();
  });

  it('renders null (default fallback) when user role does not match', () => {
    const { container } = renderWithAuth(
      <RequireRole roles={['administrator', 'operations_manager']}>
        <div data-testid="staff-content">Staff Only</div>
      </RequireRole>,
      makeAuthValue({ user: makeUser('coach') }),
    );

    expect(screen.queryByTestId('staff-content')).not.toBeInTheDocument();
    expect(container.firstChild).toBeNull();
  });

  it('renders children when user has any of the allowed roles', () => {
    renderWithAuth(
      <RequireRole roles={['administrator', 'operations_manager']}>
        <div data-testid="mgmt-content">Management</div>
      </RequireRole>,
      makeAuthValue({ user: makeUser('operations_manager') }),
    );

    expect(screen.getByTestId('mgmt-content')).toBeInTheDocument();
  });

  it('renders fallback when user is null', () => {
    renderWithAuth(
      <RequireRole roles={['administrator']} fallback={<span data-testid="no-user">No user</span>}>
        <div>Admin</div>
      </RequireRole>,
      makeAuthValue({ user: null }),
    );

    expect(screen.getByTestId('no-user')).toBeInTheDocument();
  });

  it('administrator only passes administrator checks', () => {
    const roles: UserRole[] = ['administrator', 'operations_manager', 'procurement_specialist', 'coach', 'member'];

    for (const role of roles) {
      const { unmount } = renderWithAuth(
        <RequireRole roles={[role]} fallback={<div data-testid={`blocked-${role}`}>Blocked</div>}>
          <div data-testid={`role-${role}`}>{role}</div>
        </RequireRole>,
        makeAuthValue({ user: makeUser('administrator') }),
      );

      if (role === 'administrator') {
        expect(screen.getByTestId(`role-${role}`)).toBeInTheDocument();
      } else {
        expect(screen.getByTestId(`blocked-${role}`)).toBeInTheDocument();
      }

      unmount();
    }
  });
});
