import { describe, it, expect } from 'vitest';
import {
  formatCurrency,
  formatDate,
  formatDateTime,
  formatExportFilename,
  maskField,
  formatPercentage,
  formatQuantity,
} from '@/lib/format';

describe('formatCurrency', () => {
  it('formats positive amount as USD', () => {
    expect(formatCurrency(1234.5)).toBe('$1,234.50');
  });

  it('formats zero', () => {
    expect(formatCurrency(0)).toBe('$0.00');
  });

  it('formats negative', () => {
    expect(formatCurrency(-42.1)).toBe('-$42.10');
  });

  it('pads to two decimals', () => {
    expect(formatCurrency(7)).toBe('$7.00');
  });
});

describe('formatDate', () => {
  it('formats ISO date', () => {
    expect(formatDate('2026-04-15T00:00:00Z')).toMatch(/Apr (14|15), 2026/);
  });
});

describe('formatDateTime', () => {
  it('formats ISO datetime including time', () => {
    const out = formatDateTime('2026-04-15T14:30:00Z');
    expect(out).toMatch(/2026/);
    expect(out).toMatch(/(AM|PM)/);
  });
});

describe('formatExportFilename', () => {
  it('uses provided date and sanitizes report type', () => {
    const d = new Date('2026-04-15T10:20:30Z');
    const out = formatExportFilename('Revenue Report', 'csv', d);
    expect(out.endsWith('.csv')).toBe(true);
    expect(out).toContain('revenue_report');
    expect(out).toMatch(/\d{8}_\d{6}/);
  });

  it('defaults to now when no date provided', () => {
    const out = formatExportFilename('inventory', 'pdf');
    expect(out).toMatch(/^inventory_\d{8}_\d{6}\.pdf$/);
  });

  it('replaces special characters in the report type', () => {
    const d = new Date('2026-04-15T10:20:30Z');
    const out = formatExportFilename('Sales/Tax!Summary', 'csv', d);
    expect(out.startsWith('sales_tax_summary_')).toBe(true);
  });
});

describe('maskField', () => {
  it('returns short values unchanged', () => {
    expect(maskField('abc', 4)).toBe('abc');
  });

  it('masks emails showing first char and domain', () => {
    expect(maskField('admin@example.com')).toBe('a***@example.com');
  });

  it('masks single-char-local emails with the full local retained before the mask', () => {
    // local length is 1 → branch returns `${local}***@${domain}`
    expect(maskField('x@example.com')).toBe('x***@example.com');
  });

  it('masks non-email with asterisks and visible suffix', () => {
    const masked = maskField('1234567890', 4);
    expect(masked.endsWith('7890')).toBe(true);
    expect(masked.length).toBeGreaterThanOrEqual(8);
  });

  it('masks long strings with enough asterisks', () => {
    const masked = maskField('CREDIT-CARD-NUMBER-4321', 4);
    expect(masked.endsWith('4321')).toBe(true);
    expect(masked.startsWith('*')).toBe(true);
  });
});

describe('formatPercentage', () => {
  it('formats with one decimal', () => {
    expect(formatPercentage(12.345)).toBe('12.3%');
  });
  it('handles zero', () => {
    expect(formatPercentage(0)).toBe('0.0%');
  });
});

describe('formatQuantity', () => {
  it('formats integers with thousands separator', () => {
    expect(formatQuantity(1234567)).toBe('1,234,567');
  });
  it('rounds to integer', () => {
    expect(formatQuantity(42.7)).toBe('43');
  });
});
