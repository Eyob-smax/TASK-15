import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DataTable, type Column } from '@/components/DataTable';

interface Row {
  id: string;
  name: string;
  status: string;
}

const COLUMNS: Column<Row>[] = [
  { key: 'name', label: 'Name' },
  { key: 'status', label: 'Status' },
];

const ROWS: Row[] = [
  { id: '1', name: 'Alpha', status: 'active' },
  { id: '2', name: 'Beta', status: 'inactive' },
];

const DEFAULT_PROPS = {
  columns: COLUMNS,
  rows: ROWS,
  rowKey: (row: Row) => row.id,
  page: 0,
  pageSize: 10,
  totalCount: 2,
  onPageChange: vi.fn(),
  onPageSizeChange: vi.fn(),
};

describe('DataTable', () => {
  it('renders column headers', () => {
    render(<DataTable {...DEFAULT_PROPS} />);
    expect(screen.getByText('Name')).toBeInTheDocument();
    expect(screen.getByText('Status')).toBeInTheDocument();
  });

  it('renders row data', () => {
    render(<DataTable {...DEFAULT_PROPS} />);
    expect(screen.getByText('Alpha')).toBeInTheDocument();
    expect(screen.getByText('Beta')).toBeInTheDocument();
    expect(screen.getByText('active')).toBeInTheDocument();
  });

  it('shows skeleton rows when loading', () => {
    render(<DataTable {...DEFAULT_PROPS} rows={[]} loading totalCount={0} />);
    // Skeletons are rendered as visual placeholders — table should still be present
    const table = screen.getByRole('table');
    expect(table).toBeInTheDocument();
  });

  it('shows empty state when no rows and not loading', () => {
    render(<DataTable {...DEFAULT_PROPS} rows={[]} totalCount={0} />);
    expect(screen.getByText(/no items found/i)).toBeInTheDocument();
  });

  it('shows custom empty title', () => {
    render(
      <DataTable
        {...DEFAULT_PROPS}
        rows={[]}
        totalCount={0}
        emptyTitle="No orders yet"
      />,
    );
    expect(screen.getByText('No orders yet')).toBeInTheDocument();
  });

  it('shows error alert', () => {
    render(<DataTable {...DEFAULT_PROPS} error="Failed to load data" />);
    expect(screen.getByText('Failed to load data')).toBeInTheDocument();
  });

  it('calls onPageChange when next page is clicked', () => {
    const onPageChange = vi.fn();
    render(
      <DataTable
        {...DEFAULT_PROPS}
        totalCount={30}
        onPageChange={onPageChange}
      />,
    );
    const nextButton = screen.getByTitle('Go to next page');
    fireEvent.click(nextButton);
    expect(onPageChange).toHaveBeenCalledWith(1);
  });

  it('renders with custom render function', () => {
    const columns: Column<Row>[] = [
      { key: 'name', label: 'Name' },
      {
        key: 'status',
        label: 'Status',
        render: row => <span data-testid="status-badge">{row.status.toUpperCase()}</span>,
      },
    ];
    render(<DataTable {...DEFAULT_PROPS} columns={columns} />);
    const badges = screen.getAllByTestId('status-badge');
    expect(badges).toHaveLength(2);
    expect(badges[0]).toHaveTextContent('ACTIVE');
  });
});
