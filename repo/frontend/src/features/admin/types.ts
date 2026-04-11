import type {
  User,
  UserRole,
  UserStatus,
  AuditEvent,
  BackupRun,
  BackupStatus,
  Location,
  RetentionPolicy,
} from '@/lib/types';

export type {
  User,
  UserRole,
  UserStatus,
  AuditEvent,
  BackupRun,
  BackupStatus,
  Location,
  RetentionPolicy,
};

export interface UserFilters {
  search?: string;
  role?: UserRole;
  status?: UserStatus;
  location_id?: string;
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface UserFormData {
  email: string;
  password: string;
  display_name: string;
  role: UserRole;
  location_id: string | null;
}

export interface AuditFilters {
  user_id?: string;
  action?: string;
  resource_type?: string;
  resource_id?: string;
  date_from?: string;
  date_to?: string;
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface BackupFilters {
  status?: BackupStatus;
  date_from?: string;
  date_to?: string;
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface BiometricKeyInfo {
  id: string;
  user_id: string;
  key_name: string;
  created_at: string;
  last_used_at: string | null;
  expires_at: string;
  is_active: boolean;
}
