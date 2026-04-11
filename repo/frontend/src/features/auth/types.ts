import type { User, Session } from '@/lib/types';

export type { User, Session };

export interface LoginFormData {
  email: string;
  password: string;
}

export interface SessionInfo {
  user: User;
  session: Session;
  permissions: string[];
}

export interface LoginResponse {
  user: User;
  session: Session;
}

export interface PasswordChangeRequest {
  current_password: string;
  new_password: string;
}
