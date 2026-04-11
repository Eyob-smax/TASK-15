const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';

export interface ApiError {
  code: string;
  message: string;
  details?: { field?: string; message: string }[];
  status: number;
}

function createNetworkError(): ApiError {
  const offline = typeof navigator !== 'undefined' && !navigator.onLine;
  return {
    code: offline ? 'NETWORK_OFFLINE' : 'NETWORK_ERROR',
    message: offline
      ? 'You are offline. Cached data remains available until you reconnect.'
      : 'Unable to reach the FitCommerce server.',
    status: 0,
  };
}

export function isOfflineApiError(error: unknown): error is ApiError {
  if (!error || typeof error !== 'object') {
    return false;
  }
  const candidate = error as Partial<ApiError>;
  return candidate.status === 0 && (candidate.code === 'NETWORK_OFFLINE' || candidate.code === 'NETWORK_ERROR');
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    page_size: number;
    total_count: number;
    total_pages: number;
  };
}

export interface ApiEnvelope<T> {
  data: T;
}

export interface ApiMessage {
  message: string;
}

export interface PaginationParams {
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

async function handleApiError(response: Response): Promise<never> {
  let errorData: { error?: { code?: string; message?: string; details?: { field?: string; message: string }[] } } = {};
  try {
    errorData = await response.json();
  } catch {
    // Response body may not be JSON
  }

  const apiError: ApiError = {
    code: errorData.error?.code ?? 'UNKNOWN_ERROR',
    message: errorData.error?.message ?? `HTTP ${response.status}: ${response.statusText}`,
    details: errorData.error?.details,
    status: response.status,
  };

  if (response.status === 401) {
    window.dispatchEvent(new CustomEvent('auth:session-expired'));
  }

  throw apiError;
}

async function fetchWithHandling(input: RequestInfo | URL, init: RequestInit): Promise<Response> {
  try {
    return await fetch(input, init);
  } catch {
    throw createNetworkError();
  }
}

function buildQueryString(params?: Record<string, string | number | boolean | undefined | null>): string {
  if (!params) return '';
  const searchParams = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== null && value !== '') {
      searchParams.set(key, String(value));
    }
  }
  const qs = searchParams.toString();
  return qs ? `?${qs}` : '';
}

export const apiClient = {
  async get<T>(
    path: string,
    params?: Record<string, string | number | boolean | undefined | null>,
  ): Promise<T> {
    const url = `${BASE_URL}${path}${buildQueryString(params)}`;
    const response = await fetchWithHandling(url, {
      method: 'GET',
      headers: {
        'Accept': 'application/json',
      },
      credentials: 'include',
    });

    if (!response.ok) {
      await handleApiError(response);
    }

    return response.json() as Promise<T>;
  },

  async post<T>(path: string, body: unknown): Promise<T> {
    const url = `${BASE_URL}${path}`;
    const response = await fetchWithHandling(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      await handleApiError(response);
    }

    if (response.status === 204) {
      return undefined as T;
    }

    return response.json() as Promise<T>;
  },

  async put<T>(path: string, body: unknown): Promise<T> {
    const url = `${BASE_URL}${path}`;
    const response = await fetchWithHandling(url, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      await handleApiError(response);
    }

    if (response.status === 204) {
      return undefined as T;
    }

    return response.json() as Promise<T>;
  },

  async delete<T>(path: string): Promise<T> {
    const url = `${BASE_URL}${path}`;
    const response = await fetchWithHandling(url, {
      method: 'DELETE',
      headers: {
        'Accept': 'application/json',
      },
      credentials: 'include',
    });

    if (!response.ok) {
      await handleApiError(response);
    }

    if (response.status === 204) {
      return undefined as T;
    }

    return response.json() as Promise<T>;
  },
};

function extractFilename(contentDisposition: string | null): string | undefined {
  if (!contentDisposition) {
    return undefined;
  }

  const utf8Match = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i);
  if (utf8Match?.[1]) {
    return decodeURIComponent(utf8Match[1]);
  }

  const plainMatch = contentDisposition.match(/filename="?([^"]+)"?/i);
  return plainMatch?.[1];
}

export async function downloadFile(url: string, filename?: string): Promise<string> {
  const fullUrl = `${BASE_URL}${url}`;
  const response = await fetchWithHandling(fullUrl, {
    method: 'GET',
    credentials: 'include',
  });

  if (!response.ok) {
    await handleApiError(response);
  }

  const blob = await response.blob();
  const resolvedFilename = filename ?? extractFilename(response.headers.get('Content-Disposition')) ?? 'download';
  const objectUrl = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = objectUrl;
  anchor.download = resolvedFilename;
  document.body.appendChild(anchor);
  anchor.click();
  document.body.removeChild(anchor);
  URL.revokeObjectURL(objectUrl);

  return resolvedFilename;
}
