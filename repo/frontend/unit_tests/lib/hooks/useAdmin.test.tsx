import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import {
  useUserList,
  useUser,
  useCreateUser,
  useUpdateUser,
  useDeactivateUser,
  useAuditLog,
  useSecurityEvents,
  useBackupList,
  useTriggerBackup,
  useBiometric,
  useRegisterBiometric,
  useRevokeBiometric,
  useRotateKey,
  useEncryptionKeys,
  useRetentionPolicies,
  useUpdateRetentionPolicy,
} from '@/lib/hooks/useAdmin';

const mockGet = vi.fn();
const mockPost = vi.fn();
const mockPut = vi.fn();

vi.mock('@/lib/api-client', () => ({
  apiClient: {
    get: (...args: unknown[]) => mockGet(...args),
    post: (...args: unknown[]) => mockPost(...args),
    put: (...args: unknown[]) => mockPut(...args),
  },
}));

function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

describe('useUserList / useUser', () => {
  beforeEach(() => vi.clearAllMocks());

  it('useUserList fetches with defaults', async () => {
    mockGet.mockResolvedValue({ data: [], total: 0 });
    renderHook(() => useUserList(), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith(
      '/admin/users',
      expect.objectContaining({ page: 1, page_size: 20 }),
    );
  });

  it('useUserList forwards filter params', async () => {
    mockGet.mockResolvedValue({ data: [], total: 0 });
    renderHook(
      () => useUserList({ role: 'admin', status: 'active', search: 'alice', page: 3 }),
      { wrapper },
    );
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith(
      '/admin/users',
      expect.objectContaining({
        role: 'admin',
        status: 'active',
        search: 'alice',
        page: 3,
      }),
    );
  });

  it('useUser disabled without id', () => {
    const { result } = renderHook(() => useUser(undefined), { wrapper });
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('useUser fetches and unwraps data', async () => {
    mockGet.mockResolvedValue({ data: { id: 'u1' } });
    const { result } = renderHook(() => useUser('u1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockGet).toHaveBeenCalledWith('/admin/users/u1');
    expect(result.current.data).toEqual({ id: 'u1' });
  });
});

describe('useCreateUser / useUpdateUser / useDeactivateUser', () => {
  beforeEach(() => vi.clearAllMocks());

  it('useCreateUser posts to /admin/users', async () => {
    mockPost.mockResolvedValue({ data: { id: 'u1' } });
    const { result } = renderHook(() => useCreateUser(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({
        email: 'a@b.c',
        display_name: 'Alice',
        role: 'admin',
        password: 'pw',
      });
    });
    expect(mockPost).toHaveBeenCalledWith(
      '/admin/users',
      expect.objectContaining({ email: 'a@b.c', display_name: 'Alice' }),
    );
  });

  it('useUpdateUser puts to /admin/users/:id', async () => {
    mockPut.mockResolvedValue({ data: { id: 'u1' } });
    const { result } = renderHook(() => useUpdateUser(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ id: 'u1', body: { display_name: 'Updated' } });
    });
    expect(mockPut).toHaveBeenCalledWith('/admin/users/u1', { display_name: 'Updated' });
  });

  it('useDeactivateUser posts to deactivate endpoint', async () => {
    mockPost.mockResolvedValue({ data: { id: 'u1' } });
    const { result } = renderHook(() => useDeactivateUser(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync('u1');
    });
    expect(mockPost).toHaveBeenCalledWith('/admin/users/u1/deactivate', {});
  });
});

describe('useAuditLog / useSecurityEvents', () => {
  beforeEach(() => vi.clearAllMocks());

  it('useAuditLog forwards event_type and defaults', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(() => useAuditLog({ event_type: 'login' }), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith(
      '/admin/audit-log',
      expect.objectContaining({ event_type: 'login', page: 1, page_size: 20 }),
    );
  });

  it('useSecurityEvents uses security endpoint', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(() => useSecurityEvents({ page: 2 }), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith(
      '/admin/audit-log/security',
      expect.objectContaining({ page: 2 }),
    );
  });
});

describe('useBackupList / useTriggerBackup', () => {
  beforeEach(() => vi.clearAllMocks());

  it('useBackupList fetches with defaults', async () => {
    mockGet.mockResolvedValue({ data: [] });
    renderHook(() => useBackupList(), { wrapper });
    await waitFor(() => expect(mockGet).toHaveBeenCalled());
    expect(mockGet).toHaveBeenCalledWith(
      '/admin/backups',
      expect.objectContaining({ page: 1, page_size: 20 }),
    );
  });

  it('useTriggerBackup posts empty body', async () => {
    mockPost.mockResolvedValue({ data: { id: 'b1' } });
    const { result } = renderHook(() => useTriggerBackup(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync();
    });
    expect(mockPost).toHaveBeenCalledWith('/admin/backups', {});
  });
});

describe('biometric hooks', () => {
  beforeEach(() => vi.clearAllMocks());

  it('useBiometric disabled without id', () => {
    const { result } = renderHook(() => useBiometric(undefined), { wrapper });
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('useBiometric fetches by userId', async () => {
    mockGet.mockResolvedValue({ data: { id: 'b1', user_id: 'u1' } });
    const { result } = renderHook(() => useBiometric('u1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockGet).toHaveBeenCalledWith('/admin/biometrics/u1');
    expect(result.current.data).toMatchObject({ user_id: 'u1' });
  });

  it('useRegisterBiometric posts body', async () => {
    mockPost.mockResolvedValue({ data: {} });
    const { result } = renderHook(() => useRegisterBiometric(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ user_id: 'u1', template_ref: 't1' });
    });
    expect(mockPost).toHaveBeenCalledWith('/admin/biometrics', {
      user_id: 'u1',
      template_ref: 't1',
    });
  });

  it('useRevokeBiometric posts to revoke endpoint', async () => {
    mockPost.mockResolvedValue({ data: {} });
    const { result } = renderHook(() => useRevokeBiometric(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync('u1');
    });
    expect(mockPost).toHaveBeenCalledWith('/admin/biometrics/u1/revoke', {});
  });

  it('useRotateKey posts to rotate-key endpoint', async () => {
    mockPost.mockResolvedValue({ data: {} });
    const { result } = renderHook(() => useRotateKey(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync();
    });
    expect(mockPost).toHaveBeenCalledWith('/admin/biometrics/rotate-key', {});
  });

  it('useEncryptionKeys fetches and unwraps data', async () => {
    mockGet.mockResolvedValue({ data: [{ id: 'k1' }] });
    const { result } = renderHook(() => useEncryptionKeys(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockGet).toHaveBeenCalledWith('/admin/biometrics/keys');
    expect(result.current.data).toEqual([{ id: 'k1' }]);
  });
});

describe('retention hooks', () => {
  beforeEach(() => vi.clearAllMocks());

  it('useRetentionPolicies fetches and unwraps', async () => {
    mockGet.mockResolvedValue({ data: [{ entity_type: 'audit', retention_days: 365 }] });
    const { result } = renderHook(() => useRetentionPolicies(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockGet).toHaveBeenCalledWith('/admin/retention-policies');
    expect(result.current.data).toEqual([{ entity_type: 'audit', retention_days: 365 }]);
  });

  it('useUpdateRetentionPolicy puts by entity_type', async () => {
    mockPut.mockResolvedValue({ data: { entity_type: 'audit' } });
    const { result } = renderHook(() => useUpdateRetentionPolicy(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ entity_type: 'audit', body: { retention_days: 180 } });
    });
    expect(mockPut).toHaveBeenCalledWith('/admin/retention-policies/audit', { retention_days: 180 });
  });
});
