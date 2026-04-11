import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { apiClient, isOfflineApiError, type ApiError } from '@/lib/api-client';
import type { User, Session, UserRole } from '@/lib/types';

interface CaptchaState {
  required: boolean;
  challengeId: string;
  challengeData: string;
}

interface LockoutState {
  lockedUntil: string;
}

interface AuthState {
  user: User | null;
  session: Session | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  captchaState: CaptchaState | null;
  lockoutState: LockoutState | null;
}

interface AuthContextValue extends AuthState {
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshSession: () => Promise<void>;
  verifyCaptcha: (challengeId: string, answer: string) => Promise<void>;
  clearCaptchaState: () => void;
}

export const AuthContext = createContext<AuthContextValue | null>(null);

interface AuthProviderProps {
  children: ReactNode;
}

const AUTH_SNAPSHOT_KEY = 'fitcommerce.auth.snapshot';

function readStoredAuthSnapshot(): Pick<AuthState, 'user' | 'session' | 'isAuthenticated'> | null {
  if (typeof window === 'undefined') {
    return null;
  }

  const raw = window.localStorage.getItem(AUTH_SNAPSHOT_KEY);
  if (!raw) {
    return null;
  }

  try {
    const parsed = JSON.parse(raw) as { user: User; session: Session };
    if (!parsed?.user || !parsed?.session) {
      return null;
    }

    const absoluteExpiry = Date.parse(parsed.session.absolute_expires_at);
    const idleExpiry = Date.parse(parsed.session.idle_expires_at);
    const now = Date.now();
    if (Number.isNaN(absoluteExpiry) || Number.isNaN(idleExpiry) || now > absoluteExpiry || now > idleExpiry) {
      window.localStorage.removeItem(AUTH_SNAPSHOT_KEY);
      return null;
    }

    return {
      user: parsed.user,
      session: parsed.session,
      isAuthenticated: true,
    };
  } catch {
    window.localStorage.removeItem(AUTH_SNAPSHOT_KEY);
    return null;
  }
}

function writeStoredAuthSnapshot(user: User, session: Session): void {
  if (typeof window === 'undefined') {
    return;
  }
  window.localStorage.setItem(AUTH_SNAPSHOT_KEY, JSON.stringify({ user, session }));
}

function clearStoredAuthSnapshot(): void {
  if (typeof window === 'undefined') {
    return;
  }
  window.localStorage.removeItem(AUTH_SNAPSHOT_KEY);
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [state, setState] = useState<AuthState>({
    user: null,
    session: null,
    isLoading: true,
    isAuthenticated: false,
    captchaState: null,
    lockoutState: null,
  });

  const clearAuth = useCallback(() => {
    clearStoredAuthSnapshot();
    setState(prev => ({
      ...prev,
      user: null,
      session: null,
      isLoading: false,
      isAuthenticated: false,
    }));
  }, []);

  const checkSession = useCallback(async () => {
    try {
      const response = await apiClient.get<{ data: { user: User; session: Session } }>('/auth/session');
      writeStoredAuthSnapshot(response.data.user, response.data.session);
      setState(prev => ({
        ...prev,
        user: response.data.user,
        session: response.data.session,
        isLoading: false,
        isAuthenticated: true,
      }));
    } catch (error) {
      if (isOfflineApiError(error)) {
        const snapshot = readStoredAuthSnapshot();
        setState(prev => ({
          ...prev,
          user: snapshot?.user ?? prev.user,
          session: snapshot?.session ?? prev.session,
          isAuthenticated: snapshot?.isAuthenticated ?? prev.isAuthenticated,
          isLoading: false,
        }));
        return;
      }

      setState(prev => ({
        ...prev,
        user: null,
        session: null,
        isLoading: false,
        isAuthenticated: false,
      }));
    }
  }, []);

  useEffect(() => {
    checkSession();
  }, [checkSession]);

  // Subscribe to session-expiry events fired by the API client on 401 responses.
  useEffect(() => {
    const handler = () => {
      clearAuth();
    };
    window.addEventListener('auth:session-expired', handler);
    return () => window.removeEventListener('auth:session-expired', handler);
  }, [clearAuth]);

  const login = useCallback(async (email: string, password: string) => {
    setState(prev => ({ ...prev, captchaState: null, lockoutState: null }));
    try {
      const response = await apiClient.post<{ data: { user: User; session: Session } }>('/auth/login', {
        email,
        password,
      });
      writeStoredAuthSnapshot(response.data.user, response.data.session);
      setState(prev => ({
        ...prev,
        user: response.data.user,
        session: response.data.session,
        isLoading: false,
        isAuthenticated: true,
        captchaState: null,
        lockoutState: null,
      }));
    } catch (err) {
      const apiErr = err as ApiError;
      if (apiErr.status === 423) {
        // Account locked — extract locked_until from details
        const lockedDetail = apiErr.details?.find(d => d.field === 'locked_until');
        setState(prev => ({
          ...prev,
          lockoutState: { lockedUntil: lockedDetail?.message ?? '' },
        }));
      } else if (apiErr.status === 403 && apiErr.code === 'CAPTCHA_REQUIRED') {
        // CAPTCHA gate — extract challenge info from details
        const idDetail = apiErr.details?.find(d => d.field === 'challenge_id');
        const dataDetail = apiErr.details?.find(d => d.field === 'challenge_data');
        setState(prev => ({
          ...prev,
          captchaState: {
            required: true,
            challengeId: idDetail?.message ?? '',
            challengeData: dataDetail?.message ?? '',
          },
        }));
      }
      throw err;
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      await apiClient.post('/auth/logout', {});
    } finally {
      clearStoredAuthSnapshot();
      setState({
        user: null,
        session: null,
        isLoading: false,
        isAuthenticated: false,
        captchaState: null,
        lockoutState: null,
      });
    }
  }, []);

  const refreshSession = useCallback(async () => {
    await checkSession();
  }, [checkSession]);

  const verifyCaptcha = useCallback(async (challengeId: string, answer: string) => {
    await apiClient.post('/auth/captcha/verify', { challenge_id: challengeId, answer });
    setState(prev => ({ ...prev, captchaState: null }));
  }, []);

  const clearCaptchaState = useCallback(() => {
    setState(prev => ({ ...prev, captchaState: null, lockoutState: null }));
  }, []);

  const value: AuthContextValue = {
    ...state,
    login,
    logout,
    refreshSession,
    verifyCaptcha,
    clearCaptchaState,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

interface ProtectedRouteProps {
  children: ReactNode;
  allowedRoles?: UserRole[];
}

export function ProtectedRoute({ children, allowedRoles }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading, user } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return null;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (allowedRoles && user && !allowedRoles.includes(user.role)) {
    return <Navigate to="/dashboard" replace />;
  }

  return <>{children}</>;
}

interface RequireRoleProps {
  children: ReactNode;
  roles: UserRole[];
  fallback?: ReactNode;
}

export function RequireRole({ children, roles, fallback = null }: RequireRoleProps) {
  const { user } = useAuth();

  if (!user || !roles.includes(user.role)) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
}
