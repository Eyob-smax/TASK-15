import { describe, it, expect, vi } from 'vitest';

// The routes module calls createBrowserRouter at import time. Mock out the
// offline-cache/mutations modules that Providers depends on (so importing
// App.tsx downstream does not attempt real IndexedDB access). The module
// exports a Router object — we only smoke-test that it exports the expected
// shape and route table.

vi.mock('@/lib/offline-cache', () => ({
  loadPersistedQuerySnapshot: vi.fn().mockResolvedValue(null),
  persistOfflineQueryCache: vi.fn().mockResolvedValue(null),
  hydrateOfflineQueryCache: vi.fn(),
}));

vi.mock('@/lib/offline-mutations', () => ({
  replayOfflineMutations: vi.fn().mockResolvedValue(undefined),
}));

describe('routes config', () => {
  it('exports a configured browser router', async () => {
    const mod = await import('@/app/routes');
    expect(mod.router).toBeDefined();
    // createBrowserRouter produces an object with a `routes` field.
    const routerObj = mod.router as unknown as { routes?: Array<{ path?: string }> };
    expect(Array.isArray(routerObj.routes)).toBe(true);
    expect((routerObj.routes ?? []).length).toBeGreaterThan(0);
  });

  it('includes a login route at the top level', async () => {
    const mod = await import('@/app/routes');
    const routerObj = mod.router as unknown as { routes: Array<{ path?: string }> };
    const paths = routerObj.routes.map((r) => r.path).filter(Boolean);
    expect(paths).toContain('/login');
  });
});
