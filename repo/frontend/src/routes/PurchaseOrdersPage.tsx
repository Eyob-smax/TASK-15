import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import Button from '@mui/material/Button';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import InputLabel from '@mui/material/InputLabel';
import FormControl from '@mui/material/FormControl';
import FormHelperText from '@mui/material/FormHelperText';
import IconButton from '@mui/material/IconButton';
import Divider from '@mui/material/Divider';
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import Alert from '@mui/material/Alert';
import { OfflineDataNotice } from '@/components/OfflineDataNotice';
import { PageContainer } from '@/components/PageContainer';
import { DataTable, type Column } from '@/components/DataTable';
import { FilterBar, type FilterField } from '@/components/FilterBar';
import { StatusChip } from '@/components/StatusChip';
import { RequireRole } from '@/lib/auth';
import { OFFLINE_MUTATION_MESSAGE, useOfflineStatus } from '@/lib/offline';
import { usePOList, useCreatePO, useSupplierList } from '@/lib/hooks/useProcurement';
import type { PurchaseOrder } from '@/lib/types';

const FILTER_FIELDS: FilterField[] = [
  {
    key: 'status',
    label: 'Status',
    type: 'select',
    options: [
      { value: 'created', label: 'Created' },
      { value: 'approved', label: 'Approved' },
      { value: 'received', label: 'Received' },
      { value: 'returned', label: 'Returned' },
      { value: 'voided', label: 'Voided' },
    ],
  },
];

interface POLine {
  item_id: string;
  ordered_quantity: string;
  ordered_unit_price: string;
}

const EMPTY_LINE: POLine = { item_id: '', ordered_quantity: '', ordered_unit_price: '' };

interface CreatePODialogProps {
  open: boolean;
  onClose: () => void;
  isOffline: boolean;
}

function CreatePODialog({ open, onClose, isOffline }: CreatePODialogProps) {
  const [supplierId, setSupplierId] = useState('');
  const [lines, setLines] = useState<POLine[]>([{ ...EMPTY_LINE }]);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const { data: suppliersData } = useSupplierList({ page_size: 100 });
  const suppliers = suppliersData?.data ?? [];
  const createPO = useCreatePO();

  const handleClose = () => {
    setSupplierId('');
    setLines([{ ...EMPTY_LINE }]);
    setErrors({});
    onClose();
  };

  const addLine = () => setLines(prev => [...prev, { ...EMPTY_LINE }]);

  const removeLine = (index: number) =>
    setLines(prev => prev.filter((_, i) => i !== index));

  const updateLine = (index: number, field: keyof POLine, value: string) =>
    setLines(prev => prev.map((l, i) => (i === index ? { ...l, [field]: value } : l)));

  const validate = () => {
    const errs: Record<string, string> = {};
    if (!supplierId) errs.supplier_id = 'Supplier is required';
    lines.forEach((l, i) => {
      if (!l.item_id.trim()) errs[`line_${i}_item_id`] = 'Item ID required';
      const qty = parseInt(l.ordered_quantity, 10);
      if (!l.ordered_quantity || isNaN(qty) || qty <= 0)
        errs[`line_${i}_qty`] = 'Quantity must be > 0';
      const price = parseFloat(l.ordered_unit_price);
      if (!l.ordered_unit_price || isNaN(price) || price < 0)
        errs[`line_${i}_price`] = 'Price must be ≥ 0';
    });
    return errs;
  };

  const handleSubmit = () => {
    if (isOffline) {
      return;
    }
    const errs = validate();
    if (Object.keys(errs).length > 0) {
      setErrors(errs);
      return;
    }
    createPO.mutate(
      {
        supplier_id: supplierId,
        lines: lines.map(l => ({
          item_id: l.item_id.trim(),
          ordered_quantity: parseInt(l.ordered_quantity, 10),
          ordered_unit_price: parseFloat(l.ordered_unit_price),
        })),
      },
      { onSuccess: handleClose },
    );
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <DialogTitle>New Purchase Order</DialogTitle>
      <DialogContent dividers>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: 1 }}>
          {isOffline && (
            <Alert severity="warning">{OFFLINE_MUTATION_MESSAGE}</Alert>
          )}
          <FormControl fullWidth size="small" error={Boolean(errors.supplier_id)}>
            <InputLabel id="po-supplier-label">Supplier</InputLabel>
            <Select
              labelId="po-supplier-label"
              label="Supplier"
              value={supplierId}
              onChange={e => setSupplierId(e.target.value)}
            >
              {suppliers.map(s => (
                <MenuItem key={s.id} value={s.id}>
                  {s.name}
                </MenuItem>
              ))}
            </Select>
            {errors.supplier_id && <FormHelperText>{errors.supplier_id}</FormHelperText>}
          </FormControl>

          <Divider />

          <Typography variant="subtitle2">Line Items</Typography>

          {lines.map((line, i) => (
            <Box key={i} sx={{ display: 'flex', gap: 1, alignItems: 'flex-start' }}>
              <TextField
                label="Item ID (UUID)"
                size="small"
                value={line.item_id}
                onChange={e => updateLine(i, 'item_id', e.target.value)}
                error={Boolean(errors[`line_${i}_item_id`])}
                helperText={errors[`line_${i}_item_id`]}
                sx={{ flex: 2 }}
                placeholder="e.g. 550e8400-…"
              />
              <TextField
                label="Quantity"
                size="small"
                type="number"
                value={line.ordered_quantity}
                onChange={e => updateLine(i, 'ordered_quantity', e.target.value)}
                error={Boolean(errors[`line_${i}_qty`])}
                helperText={errors[`line_${i}_qty`]}
                sx={{ flex: 1 }}
                inputProps={{ min: 1 }}
              />
              <TextField
                label="Unit Price"
                size="small"
                type="number"
                value={line.ordered_unit_price}
                onChange={e => updateLine(i, 'ordered_unit_price', e.target.value)}
                error={Boolean(errors[`line_${i}_price`])}
                helperText={errors[`line_${i}_price`]}
                sx={{ flex: 1 }}
                inputProps={{ min: 0, step: 0.01 }}
              />
              <IconButton
                size="small"
                onClick={() => removeLine(i)}
                disabled={lines.length === 1}
                sx={{ mt: 0.5 }}
              >
                <DeleteIcon fontSize="small" />
              </IconButton>
            </Box>
          ))}

          <Button
            size="small"
            startIcon={<AddIcon />}
            onClick={addLine}
            disabled={isOffline}
            sx={{ alignSelf: 'flex-start' }}
          >
            Add Line
          </Button>

          {createPO.isError && (
            <Typography color="error" variant="body2">
              Failed to create purchase order. Please try again.
            </Typography>
          )}
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose} disabled={createPO.isPending}>
          Cancel
        </Button>
        <Button
          variant="contained"
          onClick={handleSubmit}
          disabled={createPO.isPending || isOffline}
        >
          {createPO.isPending ? 'Creating…' : 'Create PO'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function shortId(id: string) {
  return id.slice(0, 8) + '…';
}

export default function PurchaseOrdersPage() {
  const navigate = useNavigate();
  const { isOffline } = useOfflineStatus();
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [filters, setFilters] = useState<Record<string, string>>({});
  const [createOpen, setCreateOpen] = useState(false);

  const { data, isLoading, error, dataUpdatedAt } = usePOList({
    page: page + 1,
    page_size: pageSize,
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

  const columns: Column<PurchaseOrder>[] = [
    {
      key: 'id',
      label: 'PO ID',
      render: row => (
        <Typography
          variant="body2"
          component="span"
          sx={{
            cursor: 'pointer',
            fontFamily: 'monospace',
            '&:hover': { textDecoration: 'underline' },
          }}
          onClick={() => navigate(`/procurement/purchase-orders/${row.id}`)}
        >
          {shortId(row.id)}
        </Typography>
      ),
    },
    {
      key: 'supplier_id',
      label: 'Supplier ID',
      render: row => (
        <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
          {shortId(row.supplier_id)}
        </Typography>
      ),
    },
    {
      key: 'status',
      label: 'Status',
      render: row => <StatusChip status={row.status} />,
    },
    {
      key: 'total_amount',
      label: 'Total Amount',
      align: 'right',
      render: row => `$${row.total_amount.toFixed(2)}`,
    },
    {
      key: 'created_at',
      label: 'Created',
      render: row => new Date(row.created_at).toLocaleDateString(),
    },
  ];

  const rows = data?.data ?? [];
  const totalCount = data?.pagination.total_count ?? 0;

  return (
    <PageContainer
      title="Purchase Orders"
      breadcrumbs={[{ label: 'Procurement', to: '/procurement' }, { label: 'Purchase Orders' }]}
      actions={
        <RequireRole roles={['administrator', 'operations_manager', 'procurement_specialist']}>
          <Button
            variant="contained"
            size="small"
            startIcon={<AddIcon />}
            onClick={() => setCreateOpen(true)}
            disabled={isOffline}
          >
            New PO
          </Button>
        </RequireRole>
      }
    >
      <Box sx={{ mb: 2 }}>
        <FilterBar fields={FILTER_FIELDS} onChange={handleFiltersChange} />
      </Box>

      <OfflineDataNotice hasData={rows.length > 0} dataUpdatedAt={dataUpdatedAt} />

      <DataTable
        columns={columns}
        rows={rows}
        rowKey={row => row.id}
        loading={isLoading}
        error={error ? 'Failed to load purchase orders. Please try again.' : null}
        page={page}
        pageSize={pageSize}
        totalCount={totalCount}
        onPageChange={handlePageChange}
        onPageSizeChange={handlePageSizeChange}
        emptyTitle="No purchase orders found"
        emptyDescription="Purchase orders will appear here once created."
      />

      <CreatePODialog open={createOpen} onClose={() => setCreateOpen(false)} isOffline={isOffline} />
    </PageContainer>
  );
}
