import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { FilterBar, type FilterField } from '@/components/FilterBar';

const TEXT_FIELDS: FilterField[] = [
  { key: 'search', label: 'Search', type: 'text', placeholder: 'Search items…' },
];

const SELECT_FIELDS: FilterField[] = [
  {
    key: 'status',
    label: 'Status',
    type: 'select',
    options: [
      { value: 'active', label: 'Active' },
      { value: 'inactive', label: 'Inactive' },
    ],
  },
];

const MIXED_FIELDS: FilterField[] = [...TEXT_FIELDS, ...SELECT_FIELDS];

describe('FilterBar', () => {
  it('renders text filter field', () => {
    render(<FilterBar fields={TEXT_FIELDS} onChange={vi.fn()} />);
    expect(screen.getByRole('textbox', { name: /search/i })).toBeInTheDocument();
  });

  it('renders select filter field', () => {
    render(<FilterBar fields={SELECT_FIELDS} onChange={vi.fn()} />);
    expect(screen.getByRole('combobox', { name: /status/i })).toBeInTheDocument();
  });

  it('calls onChange when text input changes', () => {
    const onChange = vi.fn();
    render(<FilterBar fields={TEXT_FIELDS} onChange={onChange} />);
    const input = screen.getByRole('textbox', { name: /search/i });
    fireEvent.change(input, { target: { value: 'bike' } });
    expect(onChange).toHaveBeenCalledWith({ search: 'bike' });
  });

  it('does not show clear button when no filters are set', () => {
    render(<FilterBar fields={TEXT_FIELDS} onChange={vi.fn()} />);
    expect(screen.queryByText('Clear')).not.toBeInTheDocument();
  });

  it('shows clear button when a filter has a value', () => {
    render(
      <FilterBar
        fields={TEXT_FIELDS}
        onChange={vi.fn()}
        initialValues={{ search: 'test' }}
      />,
    );
    expect(screen.getByText('Clear')).toBeInTheDocument();
  });

  it('clears all filters and calls onChange when clear is clicked', () => {
    const onChange = vi.fn();
    render(
      <FilterBar
        fields={MIXED_FIELDS}
        onChange={onChange}
        initialValues={{ search: 'bike', status: 'active' }}
      />,
    );
    fireEvent.click(screen.getByText('Clear'));
    expect(onChange).toHaveBeenLastCalledWith({});
  });

  it('renders date filter as date input', () => {
    const dateFields: FilterField[] = [
      { key: 'from', label: 'From', type: 'date' },
    ];
    render(<FilterBar fields={dateFields} onChange={vi.fn()} />);
    const input = screen.getByLabelText(/from/i);
    expect(input).toHaveAttribute('type', 'date');
  });
});
