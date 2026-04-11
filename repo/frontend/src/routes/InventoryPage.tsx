import { useState, useCallback } from 'react';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import AddIcon from '@mui/icons-material/Add';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import Alert from '@mui/material/Alert';
import { OfflineDataNotice } from '@/components/OfflineDataNotice';
import { PageContainer } from '@/components/PageContainer';
import { DataTable, type Column } from '@/components/DataTable';
import { RequireRole } from '@/lib/auth';
import { OFFLINE_MUTATION_MESSAGE, useOfflineStatus } from '@/lib/offline';
import {
  useInventorySnapshots,
  useInventoryAdjustments,
  useCreateAdjustment,
} from '@/lib/hooks/useInventory';
import { useNotify } from '@/lib/notifications';
import { inventoryAdjustmentSchema, type InventoryAdjustmentFormData } from '@/lib/validation';
import type { InventorySnapshot, InventoryAdjustment } from '@/lib/types';

const SNAPSHOT_COLUMNS: Column<InventorySnapshot>[] = [
  { key: 'item_id', label: 'Item ID', render: row => row.item_id.slice(0, 8) + '…' },
  { key: 'location_id', label: 'Location', render: row => row.location_id ? row.location_id.slice(0, 8) + '…' : '—' },
  { key: 'quantity_on_hand', label: 'On Hand', align: 'right' },
  {
    key: 'snapshot_date',
    label: 'Date',
    render: row => new Date(row.snapshot_date).toLocaleDateString(),
  },
];

const ADJUSTMENT_COLUMNS: Column<InventoryAdjustment>[] = [
  { key: 'item_id', label: 'Item ID', render: row => row.item_id.slice(0, 8) + '…' },
  {
    key: 'quantity_change',
    label: 'Change',
    align: 'right',
    render: row => (
      <Typography
        variant="body2"
        component="span"
        color={row.quantity_change >= 0 ? 'success.main' : 'error.main'}
      >
        {row.quantity_change >= 0 ? '+' : ''}{row.quantity_change}
      </Typography>
    ),
  },
  { key: 'reason', label: 'Reason' },
  {
    key: 'created_at',
    label: 'Date',
    render: row => new Date(row.created_at).toLocaleDateString(),
  },
];

function CreateAdjustmentDialog({ open, onClose, isOffline }: { open: boolean; onClose: () => void; isOffline: boolean }) {
  const notify = useNotify();
  const createMutation = useCreateAdjustment();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<InventoryAdjustmentFormData>({
    resolver: zodResolver(inventoryAdjustmentSchema),
  });

  const onSubmit = async (data: InventoryAdjustmentFormData) => {
    try {
      await createMutation.mutateAsync(data);
      notify.success('Adjustment created.');
      reset();
      onClose();
    } catch {
      notify.error('Failed to create adjustment.');
    }
  };

  const handleClose = () => { reset(); onClose(); };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="xs" fullWidth>
      <DialogTitle>Create Adjustment</DialogTitle>
      <Box component="form" onSubmit={handleSubmit(onSubmit)} noValidate>
        <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
          {isOffline && (
            <Alert severity="warning">{OFFLINE_MUTATION_MESSAGE}</Alert>
          )}
          <TextField
            {...register('item_id')}
            label="Item ID (UUID)"
            fullWidth
            size="small"
            required
            error={Boolean(errors.item_id)}
            helperText={errors.item_id?.message}
          />
          <TextField
            {...register('quantity_change', { valueAsNumber: true })}
            label="Quantity Change"
            fullWidth
            size="small"
            type="number"
            required
            inputProps={{ step: 1 }}
            error={Boolean(errors.quantity_change)}
            helperText={errors.quantity_change?.message ?? 'Use negative numbers to decrease'}
          />
          <TextField
            {...register('reason')}
            label="Reason"
            fullWidth
            size="small"
            required
            error={Boolean(errors.reason)}
            helperText={errors.reason?.message}
          />
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={handleClose} disabled={isSubmitting}>Cancel</Button>
          <Button
            type="submit"
            variant="contained"
            disabled={isSubmitting || isOffline}
            startIcon={isSubmitting ? <CircularProgress size={16} color="inherit" /> : undefined}
          >
            Create
          </Button>
        </DialogActions>
      </Box>
    </Dialog>
  );
}

export default function InventoryPage() {
  const { isOffline } = useOfflineStatus();
  const [tab, setTab] = useState(0);
  const [adjPage, setAdjPage] = useState(0);
  const [pageSize] = useState(20);
  const [adjustOpen, setAdjustOpen] = useState(false);

  const snapshots = useInventorySnapshots({});
  const adjustments = useInventoryAdjustments({ page: adjPage + 1, page_size: pageSize });

  const handleTabChange = useCallback((_: React.SyntheticEvent, v: number) => setTab(v), []);

  // snapshots response: { data: InventorySnapshot[] } (not paginated)
  const snapshotRows = (snapshots.data as { data: InventorySnapshot[] } | undefined)?.data ?? [];
  const adjRows = adjustments.data?.data ?? [];

  return (
    <PageContainer
      title="Inventory"
      breadcrumbs={[{ label: 'Inventory' }]}
      actions={
        <RequireRole roles={['administrator', 'operations_manager']}>
          <Button
            variant="contained"
            size="small"
            startIcon={<AddIcon />}
            onClick={() => setAdjustOpen(true)}
            disabled={isOffline}
          >
            Create Adjustment
          </Button>
        </RequireRole>
      }
    >
      <Tabs value={tab} onChange={handleTabChange} sx={{ mb: 2 }}>
        <Tab label="Snapshots" />
        <Tab label="Adjustments" />
      </Tabs>

      <OfflineDataNotice
        hasData={tab === 0 ? snapshotRows.length > 0 : adjRows.length > 0}
        dataUpdatedAt={tab === 0 ? snapshots.dataUpdatedAt : adjustments.dataUpdatedAt}
      />

      {tab === 0 && (
        <DataTable
          columns={SNAPSHOT_COLUMNS}
          rows={snapshotRows}
          rowKey={row => row.id}
          loading={snapshots.isLoading}
          error={snapshots.error ? 'Failed to load snapshots.' : null}
          page={0}
          pageSize={20}
          totalCount={snapshotRows.length}
          onPageChange={() => { }}
          onPageSizeChange={() => { }}
          emptyTitle="No snapshots available"
          emptyDescription="Inventory snapshots will appear here once generated."
        />
      )}

      {tab === 1 && (
        <DataTable
          columns={ADJUSTMENT_COLUMNS}
          rows={adjRows}
          rowKey={row => row.id}
          loading={adjustments.isLoading}
          error={adjustments.error ? 'Failed to load adjustments.' : null}
          page={adjPage}
          pageSize={pageSize}
          totalCount={adjustments.data?.pagination.total_count ?? 0}
          onPageChange={setAdjPage}
          onPageSizeChange={() => { }}
          emptyTitle="No adjustments found"
          emptyDescription="Inventory adjustments will appear here once created."
        />
      )}

      <CreateAdjustmentDialog open={adjustOpen} onClose={() => setAdjustOpen(false)} isOffline={isOffline} />
    </PageContainer>
  );
}
