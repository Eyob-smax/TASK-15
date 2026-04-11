import { useState, useCallback, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Divider from '@mui/material/Divider';
import IconButton from '@mui/material/IconButton';
import MenuItem from '@mui/material/MenuItem';
import Paper from '@mui/material/Paper';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import { OfflineDataNotice } from '@/components/OfflineDataNotice';
import { PageContainer } from '@/components/PageContainer';
import { DataTable, type Column } from '@/components/DataTable';
import { FilterBar, type FilterField } from '@/components/FilterBar';
import { StatusChip } from '@/components/StatusChip';
import { RequireRole, useAuth } from '@/lib/auth';
import { OFFLINE_MUTATION_MESSAGE, useOfflineStatus } from '@/lib/offline';
import { EMPTY_CATALOG_WINDOW } from '@/lib/catalog-form';
import { useBatchEdit, useItemList, type BatchEditRow as BatchEditRequestRow } from '@/lib/hooks/useItems';
import { CONDITION_LABELS, BILLING_MODEL_LABELS } from '@/lib/constants';
import { useNotify } from '@/lib/notifications';
import type { BatchEditResponse, Item } from '@/lib/types';

const FILTER_FIELDS: FilterField[] = [
  { key: 'category', label: 'Category', type: 'text' },
  { key: 'brand', label: 'Brand', type: 'text' },
  {
    key: 'condition',
    label: 'Condition',
    type: 'select',
    options: [
      { value: 'new', label: 'New' },
      { value: 'open_box', label: 'Open Box' },
      { value: 'used', label: 'Used' },
    ],
  },
  {
    key: 'status',
    label: 'Status',
    type: 'select',
    options: [
      { value: 'draft', label: 'Draft' },
      { value: 'published', label: 'Published' },
      { value: 'unpublished', label: 'Unpublished' },
    ],
  },
];

const COLUMNS: Column<Item>[] = [
  { key: 'name', label: 'Name', width: '25%' },
  { key: 'category', label: 'Category' },
  { key: 'brand', label: 'Brand' },
  {
    key: 'condition',
    label: 'Condition',
    render: row => CONDITION_LABELS[row.condition] ?? row.condition,
  },
  {
    key: 'billing_model',
    label: 'Billing',
    render: row => BILLING_MODEL_LABELS[row.billing_model] ?? row.billing_model,
  },
  {
    key: 'unit_price',
    label: 'Price',
    align: 'right',
    render: row => `$${row.unit_price.toFixed(2)}`,
  },
  {
    key: 'refundable_deposit',
    label: 'Deposit',
    align: 'right',
    render: row => `$${row.refundable_deposit.toFixed(2)}`,
  },
  {
    key: 'status',
    label: 'Status',
    render: row => <StatusChip status={row.status} />,
  },
];

type BatchEditableField = 'unit_price' | 'quantity' | 'refundable_deposit' | 'availability_windows';

interface BatchWindowDraft {
  start_time: string;
  end_time: string;
}

interface BatchEditDraftRow {
  id: string;
  item_id: string;
  field: BatchEditableField;
  new_value: string;
  availability_windows: BatchWindowDraft[];
  client_error: string | null;
  result: { success: boolean; message: string } | null;
}

const BATCH_FIELD_OPTIONS: Array<{ value: BatchEditableField; label: string }> = [
  { value: 'unit_price', label: 'Unit Price' },
  { value: 'quantity', label: 'Quantity' },
  { value: 'refundable_deposit', label: 'Refundable Deposit' },
  { value: 'availability_windows', label: 'Availability Windows' },
];

function createBatchRow(defaultItemID = ''): BatchEditDraftRow {
  return {
    id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
    item_id: defaultItemID,
    field: 'unit_price',
    new_value: '',
    availability_windows: [],
    client_error: null,
    result: null,
  };
}

function toBatchEditPayload(rows: BatchEditDraftRow[]): BatchEditRequestRow[] {
  return rows.map((row) =>
    row.field === 'availability_windows'
      ? {
          item_id: row.item_id,
          field: row.field,
          availability_windows: row.availability_windows.map((window) => ({
            start_time: new Date(window.start_time).toISOString(),
            end_time: new Date(window.end_time).toISOString(),
          })),
        }
      : {
          item_id: row.item_id,
          field: row.field,
          new_value: row.new_value.trim(),
        },
  );
}

function batchRowValidationMessage(row: BatchEditDraftRow): string | null {
  if (!row.item_id) {
    return 'Select an item for this row.';
  }

  if (row.field === 'availability_windows') {
    for (const [index, window] of row.availability_windows.entries()) {
      if (!window.start_time || !window.end_time) {
        return `Availability window ${index + 1} needs both start and end times.`;
      }
      if (new Date(window.end_time) <= new Date(window.start_time)) {
        return `Availability window ${index + 1} must end after it starts.`;
      }
    }
    return null;
  }

  if (!row.new_value.trim()) {
    return 'Enter a value for this batch edit row.';
  }

  if (row.field === 'quantity') {
    const quantity = Number(row.new_value);
    if (!Number.isInteger(quantity) || quantity < 0) {
      return 'Quantity must be a non-negative whole number.';
    }
  }

  if (row.field === 'unit_price' || row.field === 'refundable_deposit') {
    const amount = Number(row.new_value);
    if (Number.isNaN(amount) || amount < 0) {
      return 'Amounts must be non-negative numbers.';
    }
  }

  return null;
}

function rowSuccessMessage(field: BatchEditableField): string {
  switch (field) {
    case 'availability_windows':
      return 'Availability windows updated.';
    case 'refundable_deposit':
      return 'Refundable deposit updated.';
    case 'quantity':
      return 'Quantity updated.';
    default:
      return 'Price updated.';
  }
}

function BatchEditDialog({
  open,
  onClose,
  items,
  isOffline,
}: {
  open: boolean;
  onClose: () => void;
  items: Item[];
  isOffline: boolean;
}) {
  const notify = useNotify();
  const batchEditMutation = useBatchEdit();
  const [rows, setRows] = useState<BatchEditDraftRow[]>([]);

  useEffect(() => {
    if (!open) {
      return;
    }
    setRows((prev) => (prev.length > 0 ? prev : [createBatchRow(items[0]?.id ?? '')]));
  }, [open, items]);

  const updateRow = useCallback((rowID: string, updater: (row: BatchEditDraftRow) => BatchEditDraftRow) => {
    setRows(prev => prev.map(row => (row.id === rowID ? updater(row) : row)));
  }, []);

  const addRow = useCallback(() => {
    setRows(prev => [...prev, createBatchRow(items[0]?.id ?? '')]);
  }, [items]);

  const removeRow = useCallback((rowID: string) => {
    setRows(prev => (prev.length === 1 ? prev : prev.filter(row => row.id !== rowID)));
  }, []);

  const resetAndClose = useCallback(() => {
    setRows([]);
    onClose();
  }, [onClose]);

  const applyBatchResults = useCallback((response: BatchEditResponse, sourceRows: BatchEditDraftRow[]) => {
    setRows(sourceRows.map((row, index) => {
      const result = response.results[index];
      if (!result) {
        return {
          ...row,
          result: { success: false, message: 'No response returned for this row.' },
        };
      }

      return {
        ...row,
        result: {
          success: result.success,
          message: result.success ? rowSuccessMessage(row.field) : (result.failure_reason ?? 'This row failed to update.'),
        },
      };
    }));
  }, []);

  const handleSubmit = async () => {
    const validatedRows = rows.map((row) => ({
      ...row,
      client_error: batchRowValidationMessage(row),
      result: null,
    }));
    setRows(validatedRows);

    if (validatedRows.some(row => row.client_error)) {
      notify.error('Fix the highlighted batch edit rows before submitting.');
      return;
    }

    try {
      const response = await batchEditMutation.mutateAsync(toBatchEditPayload(validatedRows));
      applyBatchResults(response, validatedRows);

      if (response.failure_count === 0) {
        notify.success(`Batch edit applied to ${response.success_count} row${response.success_count === 1 ? '' : 's'}.`);
        resetAndClose();
        return;
      }

      notify.warning(`${response.success_count} row${response.success_count === 1 ? '' : 's'} succeeded and ${response.failure_count} failed.`);
    } catch {
      notify.error('Failed to apply batch edits.');
    }
  };

  return (
    <Dialog open={open} onClose={resetAndClose} maxWidth="md" fullWidth>
      <DialogTitle>Batch Edit Catalog</DialogTitle>
      <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: 2 }}>
        <Typography variant="body2" color="text.secondary">
          Update prices, quantities, deposits, or replace availability windows in one pass. Each row is validated independently and reports its own outcome.
        </Typography>

        {isOffline && (
          <Alert severity="warning">{OFFLINE_MUTATION_MESSAGE}</Alert>
        )}

        {items.length === 0 && (
          <Alert severity="info">Load at least one catalog item in the current page view before running a batch edit.</Alert>
        )}

        {rows.map((row, rowIndex) => (
          <Paper key={row.id} variant="outlined" sx={{ p: 2 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
              <Typography variant="subtitle2" fontWeight={600}>
                Row {rowIndex + 1}
              </Typography>
              <IconButton
                aria-label={`Remove batch edit row ${rowIndex + 1}`}
                onClick={() => removeRow(row.id)}
                disabled={rows.length === 1}
              >
                <DeleteIcon />
              </IconButton>
            </Box>

            <Box sx={{ display: 'grid', gap: 2, gridTemplateColumns: { xs: '1fr', md: '1.2fr 1fr 1fr' } }}>
              <TextField
                select
                label="Item"
                value={row.item_id}
                onChange={(event) => updateRow(row.id, (current) => ({
                  ...current,
                  item_id: event.target.value,
                  client_error: null,
                  result: null,
                }))}
                size="small"
              >
                {items.map((item) => (
                  <MenuItem key={item.id} value={item.id}>
                    {item.name}
                  </MenuItem>
                ))}
              </TextField>

              <TextField
                select
                label="Field"
                value={row.field}
                onChange={(event) => updateRow(row.id, (current) => ({
                  ...current,
                  field: event.target.value as BatchEditableField,
                  new_value: '',
                  availability_windows: event.target.value === 'availability_windows' ? current.availability_windows : [],
                  client_error: null,
                  result: null,
                }))}
                size="small"
              >
                {BATCH_FIELD_OPTIONS.map((option) => (
                  <MenuItem key={option.value} value={option.value}>
                    {option.label}
                  </MenuItem>
                ))}
              </TextField>

              {row.field === 'availability_windows' ? (
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <Typography variant="body2" color="text.secondary">
                    Replaces the current availability windows for the selected item.
                  </Typography>
                </Box>
              ) : (
                <TextField
                  label="New Value"
                  value={row.new_value}
                  onChange={(event) => updateRow(row.id, (current) => ({
                    ...current,
                    new_value: event.target.value,
                    client_error: null,
                    result: null,
                  }))}
                  size="small"
                  type={row.field === 'quantity' || row.field === 'unit_price' || row.field === 'refundable_deposit' ? 'number' : 'text'}
                  inputProps={row.field === 'quantity' ? { min: 0, step: 1 } : { min: 0, step: '0.01' }}
                />
              )}
            </Box>

            {row.field === 'availability_windows' && (
              <Box sx={{ mt: 2, display: 'flex', flexDirection: 'column', gap: 2 }}>
                {row.availability_windows.length === 0 && (
                  <Typography variant="body2" color="text.secondary">
                    Leave this row empty to clear the existing availability windows, or add one or more replacement windows below.
                  </Typography>
                )}

                {row.availability_windows.map((window, windowIndex) => (
                  <Box
                    key={`${row.id}-window-${windowIndex}`}
                    sx={{
                      display: 'grid',
                      gap: 2,
                      gridTemplateColumns: { xs: '1fr', md: '1fr 1fr auto' },
                      alignItems: 'center',
                    }}
                  >
                    <TextField
                      label="Start"
                      type="datetime-local"
                      value={window.start_time}
                      onChange={(event) => updateRow(row.id, (current) => ({
                        ...current,
                        availability_windows: current.availability_windows.map((currentWindow, index) => (
                          index === windowIndex ? { ...currentWindow, start_time: event.target.value } : currentWindow
                        )),
                        client_error: null,
                        result: null,
                      }))}
                      size="small"
                      InputLabelProps={{ shrink: true }}
                    />
                    <TextField
                      label="End"
                      type="datetime-local"
                      value={window.end_time}
                      onChange={(event) => updateRow(row.id, (current) => ({
                        ...current,
                        availability_windows: current.availability_windows.map((currentWindow, index) => (
                          index === windowIndex ? { ...currentWindow, end_time: event.target.value } : currentWindow
                        )),
                        client_error: null,
                        result: null,
                      }))}
                      size="small"
                      InputLabelProps={{ shrink: true }}
                    />
                    <IconButton
                      aria-label={`Remove availability window ${windowIndex + 1} from row ${rowIndex + 1}`}
                      onClick={() => updateRow(row.id, (current) => ({
                        ...current,
                        availability_windows: current.availability_windows.filter((_, index) => index !== windowIndex),
                        client_error: null,
                        result: null,
                      }))}
                    >
                      <DeleteIcon />
                    </IconButton>
                  </Box>
                ))}

                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<AddIcon />}
                  disabled={isOffline}
                  onClick={() => updateRow(row.id, (current) => ({
                    ...current,
                    availability_windows: [...current.availability_windows, { ...EMPTY_CATALOG_WINDOW }],
                    client_error: null,
                    result: null,
                  }))}
                  sx={{ alignSelf: 'flex-start' }}
                >
                  Add Availability Window
                </Button>
              </Box>
            )}

            {(row.client_error || row.result) && (
              <>
                <Divider sx={{ my: 2 }} />
                {row.client_error && <Alert severity="error">{row.client_error}</Alert>}
                {row.result && <Alert severity={row.result.success ? 'success' : 'error'}>{row.result.message}</Alert>}
              </>
            )}
          </Paper>
        ))}

        <Button variant="text" startIcon={<AddIcon />} onClick={addRow} disabled={isOffline} sx={{ alignSelf: 'flex-start' }}>
          Add Row
        </Button>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={resetAndClose} disabled={batchEditMutation.isPending}>Close</Button>
        <Button
          variant="contained"
          onClick={handleSubmit}
          disabled={batchEditMutation.isPending || items.length === 0 || isOffline}
          startIcon={<EditIcon />}
        >
          Apply Batch Edit
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default function CatalogPage() {
  const { user } = useAuth();
  const { isOffline } = useOfflineStatus();
  const navigate = useNavigate();
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [filters, setFilters] = useState<Record<string, string>>({});
  const [batchEditOpen, setBatchEditOpen] = useState(false);

  const { data, isLoading, error, dataUpdatedAt } = useItemList({
    page: page + 1,
    page_size: pageSize,
    category: filters.category || undefined,
    brand: filters.brand || undefined,
    condition: filters.condition || undefined,
    status: filters.status || undefined,
  });

  const handleFiltersChange = useCallback((values: Record<string, string>) => {
    setFilters(values);
    setPage(0);
  }, []);

  const handlePageChange = useCallback((p: number) => setPage(p), []);
  const handlePageSizeChange = useCallback((ps: number) => {
    setPageSize(ps);
    setPage(0);
  }, []);

  const rows = data?.data ?? [];
  const totalCount = data?.pagination.total_count ?? 0;
  const emptyDescription = user?.role === 'member'
    ? 'Published catalog items will appear here once they match your current filters.'
    : 'Create a new item to get started.';

  return (
    <PageContainer
      title="Catalog"
      breadcrumbs={[{ label: 'Catalog' }]}
      actions={
        <RequireRole roles={['administrator', 'operations_manager']}>
          <Box sx={{ display: 'flex', gap: 1 }}>
            <Button
              variant="outlined"
              size="small"
              startIcon={<EditIcon />}
              onClick={() => setBatchEditOpen(true)}
              disabled={isOffline}
            >
              Batch Edit
            </Button>
            <Button
              variant="contained"
              size="small"
              startIcon={<AddIcon />}
              onClick={() => navigate('/catalog/new')}
              disabled={isOffline}
            >
              New Item
            </Button>
          </Box>
        </RequireRole>
      }
    >
      <Box sx={{ mb: 2 }}>
        <FilterBar fields={FILTER_FIELDS} onChange={handleFiltersChange} />
      </Box>

      <OfflineDataNotice hasData={rows.length > 0} dataUpdatedAt={dataUpdatedAt} />

      <DataTable
        columns={COLUMNS}
        rows={rows}
        rowKey={row => row.id}
        loading={isLoading}
        error={error ? 'Failed to load items. Please try again.' : null}
        page={page}
        pageSize={pageSize}
        totalCount={totalCount}
        onPageChange={handlePageChange}
        onPageSizeChange={handlePageSizeChange}
        emptyTitle="No items found"
        emptyDescription={emptyDescription}
      />

      <BatchEditDialog
        open={batchEditOpen}
        onClose={() => setBatchEditOpen(false)}
        items={rows}
        isOffline={isOffline}
      />
    </PageContainer>
  );
}
