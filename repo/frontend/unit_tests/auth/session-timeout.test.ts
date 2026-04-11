import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { apiClient } from '@/lib/api-client';

// We test the api-client's 401 → session-expired event dispatch in isolation.

describe('API client 401 dispatches auth:session-expired', () => {
  let dispatchedEvents: string[] = [];
  let handler: () => void;
  let originalFetch: typeof fetch;

  beforeEach(() => {
    dispatchedEvents = [];
    handler = () => {
      dispatchedEvents.push('auth:session-expired');
    };
    window.addEventListener('auth:session-expired', handler);
    originalFetch = globalThis.fetch;
  });

  afterEach(() => {
    window.removeEventListener('auth:session-expired', handler);
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it('fires auth:session-expired event on 401 response', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      json: async () => ({ error: { code: 'UNAUTHORIZED', message: 'session expired' } }),
    } as Response);

    try {
      await apiClient.get('/some/protected/endpoint');
    } catch {
      // Expected — we only care about the side effect.
    }

    // Allow microtask queue to flush.
    await new Promise(resolve => setTimeout(resolve, 0));

    expect(dispatchedEvents).toContain('auth:session-expired');
  });

  it('does NOT fire auth:session-expired on non-401 errors', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 422,
      statusText: 'Unprocessable Entity',
      json: async () => ({ error: { code: 'VALIDATION_ERROR', message: 'invalid' } }),
    } as Response);

    try {
      await apiClient.post('/items', {});
    } catch {
      // Expected.
    }

    await new Promise(resolve => setTimeout(resolve, 0));

    expect(dispatchedEvents).not.toContain('auth:session-expired');
  });

  it('does NOT fire auth:session-expired on successful responses', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ data: { id: '123' } }),
    } as Response);

    await apiClient.get('/items/123');

    await new Promise(resolve => setTimeout(resolve, 0));

    expect(dispatchedEvents).not.toContain('auth:session-expired');
  });

  it('does NOT fire auth:session-expired on network failures', async () => {
    globalThis.fetch = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'));

    try {
      await apiClient.get('/some/protected/endpoint');
    } catch {
      // Expected.
    }

    await new Promise(resolve => setTimeout(resolve, 0));

    expect(dispatchedEvents).not.toContain('auth:session-expired');
  });
});
