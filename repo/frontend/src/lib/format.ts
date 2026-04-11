import dayjs from 'dayjs';

export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount);
}

export function formatDate(iso: string): string {
  return dayjs(iso).format('MMM D, YYYY');
}

export function formatDateTime(iso: string): string {
  return dayjs(iso).format('MMM D, YYYY h:mm A');
}

export function formatExportFilename(
  reportType: string,
  format: 'csv' | 'pdf',
  date?: Date,
): string {
  const d = date ?? new Date();
  const timestamp = dayjs(d).format('YYYYMMDD_HHmmss');
  const sanitized = reportType.replace(/[^a-zA-Z0-9_]/g, '_').toLowerCase();
  return `${sanitized}_${timestamp}.${format}`;
}

export function maskField(value: string, visibleChars: number = 4): string {
  if (value.length <= visibleChars) {
    return value;
  }

  // Email masking: show first char and domain
  if (value.includes('@')) {
    const [local, domain] = value.split('@');
    if (local.length <= 1) {
      return `${local}***@${domain}`;
    }
    return `${local[0]}***@${domain}`;
  }

  // Generic masking: show last N chars
  const masked = '*'.repeat(Math.max(value.length - visibleChars, 4));
  return `${masked}${value.slice(-visibleChars)}`;
}

export function formatPercentage(value: number): string {
  return `${value.toFixed(1)}%`;
}

export function formatQuantity(qty: number): string {
  return new Intl.NumberFormat('en-US', {
    maximumFractionDigits: 0,
  }).format(qty);
}
