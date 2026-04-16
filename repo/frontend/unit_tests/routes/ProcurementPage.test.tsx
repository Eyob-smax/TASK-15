import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ProcurementPage from '@/routes/ProcurementPage';

describe('ProcurementPage', () => {
  it('renders the page heading', () => {
    render(<ProcurementPage />);
    expect(screen.getByRole('heading', { name: /procurement/i })).toBeInTheDocument();
  });
});
