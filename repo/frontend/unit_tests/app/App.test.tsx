import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/react';

// Stub out Providers and RouterProvider so App can render without hitting
// IndexedDB or the real router. We only want to prove that App mounts.
vi.mock('@/app/providers', () => ({
  default: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="providers-wrapper">{children}</div>
  ),
}));

vi.mock('@/app/routes', () => ({
  router: { routes: [] },
}));

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return {
    ...actual,
    RouterProvider: () => <div data-testid="router-provider" />,
  };
});

import App from '@/app/App';

describe('App', () => {
  it('renders Providers wrapping the RouterProvider', () => {
    const { getByTestId } = render(<App />);
    expect(getByTestId('providers-wrapper')).toBeInTheDocument();
    expect(getByTestId('router-provider')).toBeInTheDocument();
  });
});
