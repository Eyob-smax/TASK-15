import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ErrorBoundary } from '@/components/ErrorBoundary';

// Suppress React's expected "error boundary caught" noise and
// componentDidCatch's console.error so test output stays readable.
beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

function Bomb({ message }: { message: string }) {
  throw new Error(message);
}

function Stable() {
  return <div data-testid="stable">ok</div>;
}

describe('ErrorBoundary', () => {
  it('renders children when there is no error', () => {
    render(
      <ErrorBoundary>
        <div data-testid="child">child</div>
      </ErrorBoundary>,
    );
    expect(screen.getByTestId('child')).toBeInTheDocument();
  });

  it('renders default fallback with the error message when a child throws', () => {
    render(
      <ErrorBoundary>
        <Bomb message="boom!" />
      </ErrorBoundary>,
    );
    expect(screen.getByText(/something went wrong/i)).toBeInTheDocument();
    expect(screen.getByText('boom!')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /try again/i })).toBeInTheDocument();
  });

  it('renders custom fallback prop when provided', () => {
    render(
      <ErrorBoundary fallback={<div data-testid="custom-fallback">custom</div>}>
        <Bomb message="x" />
      </ErrorBoundary>,
    );
    expect(screen.getByTestId('custom-fallback')).toBeInTheDocument();
    expect(screen.queryByText(/something went wrong/i)).not.toBeInTheDocument();
  });

  it('reset clears the error state and re-renders children', () => {
    const { rerender } = render(
      <ErrorBoundary>
        <Bomb message="x" />
      </ErrorBoundary>,
    );
    expect(screen.getByText(/something went wrong/i)).toBeInTheDocument();

    // Swap the throwing child for a stable one. ErrorBoundary state still
    // shows error UI (rerender does not reset state automatically).
    rerender(
      <ErrorBoundary>
        <Stable />
      </ErrorBoundary>,
    );
    expect(screen.getByText(/something went wrong/i)).toBeInTheDocument();

    // Click Try again: setState resets hasError; children now mount fresh.
    fireEvent.click(screen.getByRole('button', { name: /try again/i }));
    expect(screen.getByTestId('stable')).toBeInTheDocument();
  });
});
