import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { apiClient, isOfflineApiError, downloadFile, type ApiError } from '@/lib/api-client';

// All api-client methods use fetch under the hood. We replace globalThis.fetch
// with a vi.fn so we can assert calls and stub responses per-test. We also
// stub navigator.onLine to exercise the offline code path in createNetworkError.

function makeResponse(init: {
  status?: number;
  ok?: boolean;
  body?: unknown;
  statusText?: string;
  contentDisposition?: string;
  isBlob?: boolean;
}): Response {
  const status = init.status ?? 200;
  const ok = init.ok ?? (status >= 200 && status < 300);
  const headers: Record<string, string> = {};
  if (init.contentDisposition) {
    headers['Content-Disposition'] = init.contentDisposition;
  }
  return {
    status,
    ok,
    statusText: init.statusText ?? 'OK',
    headers: {
      get: (name: string) => headers[name] ?? null,
    },
    json: async () => init.body,
    blob: async () => new Blob(['x'], { type: 'text/plain' }),
  } as unknown as Response;
}

const originalFetch = globalThis.fetch;

describe('isOfflineApiError', () => {
  it('returns true for NETWORK_OFFLINE with status 0', () => {
    expect(
      isOfflineApiError({ code: 'NETWORK_OFFLINE', message: 'off', status: 0 }),
    ).toBe(true);
  });

  it('returns true for NETWORK_ERROR with status 0', () => {
    expect(
      isOfflineApiError({ code: 'NETWORK_ERROR', message: 'x', status: 0 }),
    ).toBe(true);
  });

  it('returns false for non-network status', () => {
    expect(
      isOfflineApiError({ code: 'NETWORK_ERROR', message: 'x', status: 500 }),
    ).toBe(false);
  });

  it('returns false for null / primitive', () => {
    expect(isOfflineApiError(null)).toBe(false);
    expect(isOfflineApiError('boom')).toBe(false);
    expect(isOfflineApiError(123)).toBe(false);
  });
});

describe('apiClient.get', () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    globalThis.fetch = fetchMock as unknown as typeof fetch;
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('builds query string and returns parsed JSON', async () => {
    fetchMock.mockResolvedValue(makeResponse({ body: { ok: true } }));
    const out = await apiClient.get<{ ok: boolean }>('/items', {
      page: 1,
      search: 'bike',
      dropMe: undefined,
      alsoDropMe: null,
      emptyString: '',
    });
    expect(out).toEqual({ ok: true });

    const [url, init] = fetchMock.mock.calls[0];
    expect(String(url)).toContain('/items?');
    expect(String(url)).toContain('page=1');
    expect(String(url)).toContain('search=bike');
    expect(String(url)).not.toContain('dropMe');
    expect(String(url)).not.toContain('alsoDropMe');
    expect(String(url)).not.toContain('emptyString');
    expect(init.method).toBe('GET');
    expect(init.credentials).toBe('include');
  });

  it('omits query string when no params', async () => {
    fetchMock.mockResolvedValue(makeResponse({ body: { data: [] } }));
    await apiClient.get('/items');
    const [url] = fetchMock.mock.calls[0];
    expect(String(url)).not.toContain('?');
  });

  it('throws a structured ApiError with server-provided code on failure', async () => {
    fetchMock.mockResolvedValue(
      makeResponse({
        status: 422,
        ok: false,
        statusText: 'Unprocessable',
        body: {
          error: {
            code: 'VALIDATION_ERROR',
            message: 'invalid',
            details: [{ field: 'name', message: 'required' }],
          },
        },
      }),
    );
    try {
      await apiClient.get('/items');
      throw new Error('should not reach');
    } catch (e) {
      const err = e as ApiError;
      expect(err.code).toBe('VALIDATION_ERROR');
      expect(err.message).toBe('invalid');
      expect(err.status).toBe(422);
      expect(err.details?.[0]?.field).toBe('name');
    }
  });

  it('falls back to HTTP status text when body is not JSON', async () => {
    fetchMock.mockResolvedValue({
      status: 500,
      ok: false,
      statusText: 'Server Error',
      headers: { get: () => null },
      json: async () => {
        throw new Error('not json');
      },
    } as unknown as Response);
    try {
      await apiClient.get('/boom');
      throw new Error('should not reach');
    } catch (e) {
      const err = e as ApiError;
      expect(err.code).toBe('UNKNOWN_ERROR');
      expect(err.status).toBe(500);
      expect(err.message).toContain('500');
    }
  });

  it('dispatches auth:session-expired on 401', async () => {
    const listener = vi.fn();
    window.addEventListener('auth:session-expired', listener);
    fetchMock.mockResolvedValue(
      makeResponse({ status: 401, ok: false, body: { error: { code: 'UNAUTH', message: 'no' } } }),
    );
    await expect(apiClient.get('/items')).rejects.toMatchObject({ status: 401 });
    expect(listener).toHaveBeenCalled();
    window.removeEventListener('auth:session-expired', listener);
  });

  it('maps fetch rejection to NETWORK_ERROR when online', async () => {
    Object.defineProperty(window.navigator, 'onLine', { configurable: true, value: true });
    fetchMock.mockRejectedValue(new Error('network blown up'));
    try {
      await apiClient.get('/items');
      throw new Error('should not reach');
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(0);
      expect(err.code).toBe('NETWORK_ERROR');
    }
  });

  it('maps fetch rejection to NETWORK_OFFLINE when offline', async () => {
    Object.defineProperty(window.navigator, 'onLine', { configurable: true, value: false });
    fetchMock.mockRejectedValue(new Error('no net'));
    try {
      await apiClient.get('/items');
      throw new Error('should not reach');
    } catch (e) {
      const err = e as ApiError;
      expect(err.code).toBe('NETWORK_OFFLINE');
    }
    Object.defineProperty(window.navigator, 'onLine', { configurable: true, value: true });
  });
});

describe('apiClient.post / put / delete', () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    globalThis.fetch = fetchMock as unknown as typeof fetch;
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('post sends JSON body and returns parsed response', async () => {
    fetchMock.mockResolvedValue(makeResponse({ status: 201, body: { id: 'x' } }));
    const out = await apiClient.post<{ id: string }>('/items', { name: 'Bike' });
    expect(out).toEqual({ id: 'x' });
    const [, init] = fetchMock.mock.calls[0];
    expect(init.method).toBe('POST');
    expect(init.body).toBe(JSON.stringify({ name: 'Bike' }));
    expect((init.headers as Record<string, string>)['Content-Type']).toBe('application/json');
  });

  it('post returns undefined for 204', async () => {
    fetchMock.mockResolvedValue(makeResponse({ status: 204, body: undefined }));
    const out = await apiClient.post('/items', {});
    expect(out).toBeUndefined();
  });

  it('put sends JSON body', async () => {
    fetchMock.mockResolvedValue(makeResponse({ body: { id: 'x' } }));
    await apiClient.put('/items/x', { name: 'y' });
    const [url, init] = fetchMock.mock.calls[0];
    expect(String(url)).toContain('/items/x');
    expect(init.method).toBe('PUT');
    expect(init.body).toBe(JSON.stringify({ name: 'y' }));
  });

  it('put returns undefined for 204', async () => {
    fetchMock.mockResolvedValue(makeResponse({ status: 204, body: undefined }));
    const out = await apiClient.put('/items/x', {});
    expect(out).toBeUndefined();
  });

  it('delete uses DELETE method and returns undefined on 204', async () => {
    fetchMock.mockResolvedValue(makeResponse({ status: 204, body: undefined }));
    const out = await apiClient.delete('/items/x');
    expect(out).toBeUndefined();
    const [, init] = fetchMock.mock.calls[0];
    expect(init.method).toBe('DELETE');
  });

  it('delete returns JSON on 200', async () => {
    fetchMock.mockResolvedValue(makeResponse({ body: { ok: true } }));
    const out = await apiClient.delete<{ ok: boolean }>('/items/x');
    expect(out).toEqual({ ok: true });
  });
});

describe('downloadFile', () => {
  let fetchMock: ReturnType<typeof vi.fn>;
  const originalCreateObjectURL = URL.createObjectURL;
  const originalRevokeObjectURL = URL.revokeObjectURL;

  beforeEach(() => {
    fetchMock = vi.fn();
    globalThis.fetch = fetchMock as unknown as typeof fetch;
    URL.createObjectURL = vi.fn(() => 'blob:mock');
    URL.revokeObjectURL = vi.fn();
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
    URL.createObjectURL = originalCreateObjectURL;
    URL.revokeObjectURL = originalRevokeObjectURL;
  });

  it('uses the provided filename when given', async () => {
    fetchMock.mockResolvedValue(makeResponse({ body: undefined, isBlob: true }));
    const out = await downloadFile('/exports/1/download', 'myfile.csv');
    expect(out).toBe('myfile.csv');
    expect(URL.createObjectURL).toHaveBeenCalled();
    expect(URL.revokeObjectURL).toHaveBeenCalled();
  });

  it('falls back to plain filename= header when present', async () => {
    fetchMock.mockResolvedValue(
      makeResponse({
        body: undefined,
        isBlob: true,
        contentDisposition: 'attachment; filename="report.csv"',
      }),
    );
    const out = await downloadFile('/exports/1/download');
    expect(out).toBe('report.csv');
  });

  it('decodes UTF-8 filename* parameter', async () => {
    fetchMock.mockResolvedValue(
      makeResponse({
        body: undefined,
        isBlob: true,
        contentDisposition: "attachment; filename*=UTF-8''rep%C3%B3rt.csv",
      }),
    );
    const out = await downloadFile('/exports/1/download');
    expect(out).toBe('rep\u00F3rt.csv');
  });

  it('uses "download" when no disposition or filename provided', async () => {
    fetchMock.mockResolvedValue(makeResponse({ body: undefined, isBlob: true }));
    const out = await downloadFile('/exports/1/download');
    expect(out).toBe('download');
  });

  it('throws ApiError on non-ok response', async () => {
    fetchMock.mockResolvedValue(
      makeResponse({ status: 500, ok: false, body: { error: { code: 'BOOM', message: 'no' } } }),
    );
    await expect(downloadFile('/exports/1/download')).rejects.toMatchObject({
      code: 'BOOM',
      status: 500,
    });
  });
});
