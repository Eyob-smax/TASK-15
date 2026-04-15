import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import Providers from '@/app/providers';

// Mock the offline-cache module so we don't need IndexedDB in jsdom.
vi.mock('@/lib/offline-cache', () => ({
  loadPersistedQuerySnapshot: vi.fn().mockResolvedValue(null),
  persistOfflineQueryCache: vi.fn().mockResolvedValue(null),
  hydrateOfflineQueryCache: vi.fn(),
}));

vi.mock('@/lib/offline-mutations', () => ({
  replayOfflineMutations: vi.fn().mockResolvedValue(undefined),
}));

// Keep AuthProvider from hitting the real API.
vi.mock('@/lib/auth', async () => {
  const actual = await vi.importActual<typeof import('@/lib/auth')>('@/lib/auth');
  return {
    ...actual,
    AuthProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  };
});

describe('Providers', () => {
  beforeEach(() => vi.clearAllMocks());

  it('renders children once hydrated', async () => {
    render(
      <Providers>
        <div data-testid="app-body">Hello</div>
      </Providers>,
    );

    // After the async hydrate effect completes, children appear.
    await waitFor(() =>
      expect(screen.getByTestId('app-body')).toBeInTheDocument(),
    );
  });
});
