import { describe, it, expect } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import { NotificationsProvider, useNotify } from '@/lib/notifications';

function Trigger({ severity, message }: { severity: 'success' | 'error' | 'warning' | 'info'; message: string }) {
  const notify = useNotify();
  return (
    <button type="button" onClick={() => notify[severity](message)}>
      trigger-{severity}
    </button>
  );
}

function renderWithProvider(ui: React.ReactNode) {
  return render(<NotificationsProvider>{ui}</NotificationsProvider>);
}

describe('NotificationsProvider', () => {
  it('displays a success notification when triggered', async () => {
    renderWithProvider(<Trigger severity="success" message="saved!" />);
    const btn = screen.getByText('trigger-success');
    await act(async () => {
      btn.click();
    });
    expect(await screen.findByText('saved!')).toBeInTheDocument();
  });

  it('displays error notification', async () => {
    renderWithProvider(<Trigger severity="error" message="oops" />);
    await act(async () => {
      screen.getByText('trigger-error').click();
    });
    expect(await screen.findByText('oops')).toBeInTheDocument();
  });

  it('supports warning and info variants', async () => {
    renderWithProvider(
      <>
        <Trigger severity="warning" message="careful" />
        <Trigger severity="info" message="heads-up" />
      </>,
    );
    await act(async () => {
      screen.getByText('trigger-warning').click();
    });
    expect(await screen.findByText('careful')).toBeInTheDocument();
  });
});

describe('useNotify outside provider', () => {
  it('throws helpful error', () => {
    function NoContext() {
      useNotify();
      return null;
    }
    // React will log the error; we just want to check the throw.
    expect(() => render(<NoContext />)).toThrow(/useNotify must be used within NotificationsProvider/);
  });
});
