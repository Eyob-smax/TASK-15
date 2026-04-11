import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, type ApiEnvelope, type PaginatedResponse } from '@/lib/api-client';
import type { User, UserRole, AuditEvent, BackupRun, RetentionPolicy } from '@/lib/types';

// ─── Shared ──────────────────────────────────────────────────────────────────

interface SingleResponse<T> {
  data: T;
}

// ─── User param / body types ──────────────────────────────────────────────────

interface UserListParams {
  page?: number;
  page_size?: number;
  role?: UserRole;
  status?: string;
  search?: string;
}

interface CreateUserBody {
  email: string;
  display_name: string;
  role: UserRole;
  password: string;
  location_id?: string;
}

interface UpdateUserPayload {
  id: string;
  body: Partial<Omit<CreateUserBody, 'password'>>;
}

// ─── Audit param types ────────────────────────────────────────────────────────

interface AuditLogParams {
  event_type?: string;
  page?: number;
  page_size?: number;
}

interface SecurityEventParams {
  page?: number;
  page_size?: number;
}

// ─── Backup param types ───────────────────────────────────────────────────────

interface BackupListParams {
  page?: number;
  page_size?: number;
}

// ─── Biometric types ──────────────────────────────────────────────────────────

interface BiometricRecord {
  id: string;
  user_id: string;
  template_ref: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

interface EncryptionKey {
  id: string;
  purpose: string;
  created_at: string;
  is_active: boolean;
}

interface RegisterBiometricBody {
  user_id: string;
  template_ref: string;
}

// ─── Retention param types ────────────────────────────────────────────────────

interface UpdateRetentionPolicyPayload {
  entity_type: string;
  body: Pick<RetentionPolicy, 'retention_days'>;
}

// ─── User hooks ───────────────────────────────────────────────────────────────

export function useUserList(params: UserListParams = {}) {
  return useQuery({
    queryKey: ['users', params],
    queryFn: () =>
      apiClient.get<PaginatedResponse<User>>('/admin/users', {
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
        role: params.role,
        status: params.status,
        search: params.search,
      }),
  });
}

export function useUser(id: string | undefined) {
  return useQuery({
    queryKey: ['users', id],
    queryFn: async () => {
      const response = await apiClient.get<SingleResponse<User>>(`/admin/users/${id}`);
      return response.data;
    },
    enabled: Boolean(id),
  });
}

export function useCreateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateUserBody) =>
      apiClient.post<SingleResponse<User>>('/admin/users', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['users'] });
    },
  });
}

export function useUpdateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: UpdateUserPayload) =>
      apiClient.put<SingleResponse<User>>(`/admin/users/${id}`, body),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['users'] });
      qc.invalidateQueries({ queryKey: ['users', id] });
    },
  });
}

export function useDeactivateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiClient.post<SingleResponse<User>>(`/admin/users/${id}/deactivate`, {}),
    onSuccess: (_, id) => {
      qc.invalidateQueries({ queryKey: ['users'] });
      qc.invalidateQueries({ queryKey: ['users', id] });
    },
  });
}

// ─── Audit hooks ──────────────────────────────────────────────────────────────

export function useAuditLog(params: AuditLogParams = {}) {
  return useQuery({
    queryKey: ['audit', params],
    queryFn: () =>
      apiClient.get<PaginatedResponse<AuditEvent>>('/admin/audit-log', {
        event_type: params.event_type,
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
      }),
  });
}

export function useSecurityEvents(params: SecurityEventParams = {}) {
  return useQuery({
    queryKey: ['audit', 'security', params],
    queryFn: () =>
      apiClient.get<PaginatedResponse<AuditEvent>>('/admin/audit-log/security', {
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
      }),
  });
}

// ─── Backup hooks ─────────────────────────────────────────────────────────────

export function useBackupList(params: BackupListParams = {}) {
  return useQuery({
    queryKey: ['backups', params],
    queryFn: () =>
      apiClient.get<PaginatedResponse<BackupRun>>('/admin/backups', {
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
      }),
  });
}

export function useTriggerBackup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () =>
      apiClient.post<SingleResponse<BackupRun>>('/admin/backups', {}),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['backups'] });
    },
  });
}

// ─── Biometric hooks ──────────────────────────────────────────────────────────

export function useBiometric(userId: string | undefined) {
  return useQuery({
    queryKey: ['biometrics', userId],
    queryFn: async () => {
      const response = await apiClient.get<SingleResponse<BiometricRecord>>(
        `/admin/biometrics/${userId}`,
      );
      return response.data;
    },
    enabled: Boolean(userId),
  });
}

export function useRegisterBiometric() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: RegisterBiometricBody) =>
      apiClient.post<SingleResponse<BiometricRecord>>('/admin/biometrics', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['biometrics'] });
    },
  });
}

export function useRevokeBiometric() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (userId: string) =>
      apiClient.post<SingleResponse<BiometricRecord>>(
        `/admin/biometrics/${userId}/revoke`,
        {},
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['biometrics'] });
    },
  });
}

export function useRotateKey() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () =>
      apiClient.post<SingleResponse<EncryptionKey>>('/admin/biometrics/rotate-key', {}),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['encryption-keys'] });
    },
  });
}

export function useEncryptionKeys() {
  return useQuery({
    queryKey: ['encryption-keys'],
    queryFn: async () => {
      const response = await apiClient.get<ApiEnvelope<EncryptionKey[]>>('/admin/biometrics/keys');
      return response.data;
    },
  });
}

// ─── Retention hooks ──────────────────────────────────────────────────────────

export function useRetentionPolicies() {
  return useQuery({
    queryKey: ['retention-policies'],
    queryFn: async () => {
      const response = await apiClient.get<ApiEnvelope<RetentionPolicy[]>>('/admin/retention-policies');
      return response.data;
    },
  });
}

export function useUpdateRetentionPolicy() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ entity_type, body }: UpdateRetentionPolicyPayload) =>
      apiClient.put<SingleResponse<RetentionPolicy>>(
        `/admin/retention-policies/${entity_type}`,
        body,
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['retention-policies'] });
    },
  });
}
