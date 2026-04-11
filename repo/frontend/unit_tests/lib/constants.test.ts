import { describe, it, expect } from 'vitest';
import {
  USER_ROLES,
  ROLE_PERMISSIONS,
  DEFAULT_REFUNDABLE_DEPOSIT,
  AUTO_CLOSE_TIMEOUT_MINUTES,
  RETENTION_FINANCIAL_YEARS,
  RETENTION_ACCESS_LOG_YEARS,
} from '@/lib/constants';

// ─── USER_ROLES ────────────────────────────────────────────────────────────────

describe('USER_ROLES', () => {
  it('has exactly 5 entries', () => {
    expect(USER_ROLES).toHaveLength(5);
  });

  it('contains all expected roles', () => {
    expect(USER_ROLES).toContain('administrator');
    expect(USER_ROLES).toContain('operations_manager');
    expect(USER_ROLES).toContain('procurement_specialist');
    expect(USER_ROLES).toContain('coach');
    expect(USER_ROLES).toContain('member');
  });
});

// ─── ROLE_PERMISSIONS ──────────────────────────────────────────────────────────

describe('ROLE_PERMISSIONS', () => {
  it('covers all 5 roles', () => {
    const roles = Object.keys(ROLE_PERMISSIONS);
    expect(roles).toHaveLength(5);
    expect(roles).toContain('administrator');
    expect(roles).toContain('operations_manager');
    expect(roles).toContain('procurement_specialist');
    expect(roles).toContain('coach');
    expect(roles).toContain('member');
  });

  it('administrator has the most permissions', () => {
    const adminPerms = ROLE_PERMISSIONS['administrator'];
    for (const [role, perms] of Object.entries(ROLE_PERMISSIONS)) {
      if (role !== 'administrator') {
        expect(adminPerms.length).toBeGreaterThanOrEqual(perms.length);
      }
    }
  });

  it('administrator can access all core modules', () => {
    const adminPerms = ROLE_PERMISSIONS['administrator'];
    expect(adminPerms).toContain('dashboard');
    expect(adminPerms).toContain('catalog');
    expect(adminPerms).toContain('inventory');
    expect(adminPerms).toContain('orders');
    expect(adminPerms).toContain('procurement');
    expect(adminPerms).toContain('reports');
    expect(adminPerms).toContain('admin');
    expect(adminPerms).toContain('audit');
    expect(adminPerms).toContain('backups');
  });

  it('member has the fewest permissions', () => {
    const memberPerms = ROLE_PERMISSIONS['member'];
    for (const [role, perms] of Object.entries(ROLE_PERMISSIONS)) {
      if (role !== 'member') {
        expect(memberPerms.length).toBeLessThanOrEqual(perms.length);
      }
    }
  });

  it('member permissions include catalog, group_buys, orders', () => {
    const memberPerms = ROLE_PERMISSIONS['member'];
    expect(memberPerms).toContain('catalog');
    expect(memberPerms).toContain('group_buys');
    expect(memberPerms).toContain('orders');
  });

  it('member does not have admin or procurement access', () => {
    const memberPerms = ROLE_PERMISSIONS['member'];
    expect(memberPerms).not.toContain('admin');
    expect(memberPerms).not.toContain('procurement');
    expect(memberPerms).not.toContain('dashboard');
  });
});

// ─── Constants ─────────────────────────────────────────────────────────────────

describe('business constants', () => {
  it('DEFAULT_REFUNDABLE_DEPOSIT is 50', () => {
    expect(DEFAULT_REFUNDABLE_DEPOSIT).toBe(50);
  });

  it('AUTO_CLOSE_TIMEOUT_MINUTES is 30', () => {
    expect(AUTO_CLOSE_TIMEOUT_MINUTES).toBe(30);
  });

  it('RETENTION_FINANCIAL_YEARS is 7', () => {
    expect(RETENTION_FINANCIAL_YEARS).toBe(7);
  });

  it('RETENTION_ACCESS_LOG_YEARS is 2', () => {
    expect(RETENTION_ACCESS_LOG_YEARS).toBe(2);
  });
});
