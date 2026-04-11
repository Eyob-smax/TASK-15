import { useState, useCallback, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import LinearProgress from '@mui/material/LinearProgress';
import Typography from '@mui/material/Typography';
import AddIcon from '@mui/icons-material/Add';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import CircularProgress from '@mui/material/CircularProgress';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import Alert from '@mui/material/Alert';
import { OfflineDataNotice } from '@/components/OfflineDataNotice';
import { PageContainer } from '@/components/PageContainer';
import { DataTable, type Column } from '@/components/DataTable';
import { FilterBar, type FilterField } from '@/components/FilterBar';
import { StatusChip } from '@/components/StatusChip';
import { useAuth } from '@/lib/auth';
import { useOfflineStatus } from '@/lib/offline';
import { useCampaignList, useCreateCampaign } from '@/lib/hooks/useCampaigns';
import { useNotify } from '@/lib/notifications';
import { createCampaignSchema, type CreateCampaignFormData } from '@/lib/validation';
import type { GroupBuyCampaign } from '@/lib/types';

const FILTER_FIELDS: FilterField[] = [
  {
    key: 'status',
    label: 'Status',
    type: 'select',
    options: [
      { value: 'active', label: 'Active' },
      { value: 'succeeded', label: 'Succeeded' },
      { value: 'failed', label: 'Failed' },
      { value: 'cancelled', label: 'Cancelled' },
    ],
  },
];

function ProgressCell({ campaign }: { campaign: GroupBuyCampaign }) {
  const pct = campaign.min_quantity > 0
    ? Math.min((campaign.current_committed_qty / campaign.min_quantity) * 100, 100)
    : 0;
  return (
    <Box sx={{ minWidth: 120 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.25 }}>
        <Typography variant="caption">{campaign.current_committed_qty}/{campaign.min_quantity}</Typography>
        <Typography variant="caption">{Math.round(pct)}%</Typography>
      </Box>
      <LinearProgress
        variant="determinate"
        value={pct}
        color={pct >= 100 ? 'success' : 'primary'}
        sx={{ height: 6, borderRadius: 3 }}
      />
    </Box>
  );
}

function CreateCampaignDialog({
  open,
  onClose,
  defaultItemID,
  itemReadOnly = false,
  title = 'Create Campaign',
  submitLabel = 'Create',
  isOffline = false,
}: {
  open: boolean;
  onClose: () => void;
  defaultItemID?: string;
  itemReadOnly?: boolean;
  title?: string;
  submitLabel?: string;
  isOffline?: boolean;
}) {
  const notify = useNotify();
  const createMutation = useCreateCampaign();

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm<CreateCampaignFormData>({
    resolver: zodResolver(createCampaignSchema),
  });

  useEffect(() => {
    if (!open) {
      return;
    }
    setValue('item_id', defaultItemID ?? '', {
      shouldDirty: false,
      shouldTouch: false,
      shouldValidate: false,
    });
  }, [open, defaultItemID, setValue]);

  const onSubmit = async (data: CreateCampaignFormData) => {
    try {
      await createMutation.mutateAsync({
        item_id: data.item_id,
        min_quantity: data.min_quantity,
        cutoff_time: data.cutoff_time,
      });
      if (isOffline) {
        notify.info('Action queued — will sync when you reconnect.');
      } else {
        notify.success(submitLabel === 'Start' ? 'Group buy started successfully.' : 'Campaign created successfully.');
      }
      reset();
      onClose();
    } catch {
      notify.error(submitLabel === 'Start' ? 'Failed to start group buy.' : 'Failed to create campaign.');
    }
  };

  const handleClose = () => {
    reset();
    onClose();
  };

  const itemIDHelperText = errors.item_id?.message
    ?? (itemReadOnly ? 'This campaign is locked to the selected catalog item.' : undefined);

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>{title}</DialogTitle>
      <Box component="form" onSubmit={handleSubmit(onSubmit)} noValidate>
        <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
          {isOffline && (
            <Alert severity="info">Offline — this action will be queued and applied when you reconnect.</Alert>
          )}
          <TextField
            {...register('item_id')}
            label="Item ID (UUID)"
            fullWidth
            size="small"
            required
            disabled={itemReadOnly}
            error={Boolean(errors.item_id)}
            helperText={itemIDHelperText}
          />
          <TextField
            {...register('min_quantity', { valueAsNumber: true })}
            label="Minimum Quantity"
            fullWidth
            size="small"
            type="number"
            required
            inputProps={{ min: 1, step: 1 }}
            error={Boolean(errors.min_quantity)}
            helperText={errors.min_quantity?.message}
          />
          <TextField
            {...register('cutoff_time')}
            label="Cutoff Time"
            fullWidth
            size="small"
            type="datetime-local"
            required
            InputLabelProps={{ shrink: true }}
            error={Boolean(errors.cutoff_time)}
            helperText={errors.cutoff_time?.message}
          />
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={handleClose} disabled={isSubmitting}>Cancel</Button>
          <Button
            type="submit"
            variant="contained"
            disabled={isSubmitting}
            startIcon={isSubmitting ? <CircularProgress size={16} color="inherit" /> : undefined}
          >
            {submitLabel}
          </Button>
        </DialogActions>
      </Box>
    </Dialog>
  );
}

export default function GroupBuysPage() {
  const { user } = useAuth();
  const { isOffline } = useOfflineStatus();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [filters, setFilters] = useState<Record<string, string>>({});
  const [createOpen, setCreateOpen] = useState(false);

  const defaultItemID = searchParams.get('item_id') ?? undefined;
  const canCreateCampaign = user?.role === 'administrator' || user?.role === 'operations_manager';
  const memberStartingFromItem = user?.role === 'member' && Boolean(defaultItemID);

  useEffect(() => {
    if (defaultItemID) {
      setCreateOpen(true);
    }
  }, [defaultItemID]);

  const { data, isLoading, error, dataUpdatedAt } = useCampaignList({
    page: page + 1,
    page_size: pageSize,
    status: filters.status || undefined,
  });
  const emptyDescription = user?.role === 'member'
    ? 'Open a published catalog item to start a group buy, or join an existing campaign once it appears here.'
    : 'Create a group buy campaign to get started.';

  const handleFiltersChange = useCallback((values: Record<string, string>) => {
    setFilters(values);
    setPage(0);
  }, []);

  const handleCreateDialogClose = useCallback(() => {
    setCreateOpen(false);
    if (searchParams.has('item_id')) {
      const next = new URLSearchParams(searchParams);
      next.delete('item_id');
      setSearchParams(next, { replace: true });
    }
  }, [searchParams, setSearchParams]);

  const columns: Column<GroupBuyCampaign>[] = [
    {
      key: 'item_id',
      label: 'Item',
      width: '28%',
      render: row => (
        <Typography
          variant="body2"
          component="span"
          sx={{ cursor: 'pointer', fontWeight: 600, '&:hover': { textDecoration: 'underline' } }}
          onClick={() => navigate(`/group-buys/${row.id}`)}
        >
          {`Item ${row.item_id.slice(0, 8)}`}
        </Typography>
      ),
    },
    {
      key: 'status',
      label: 'Status',
      render: row => <StatusChip status={row.status} />,
    },
    {
      key: 'progress',
      label: 'Progress',
      width: '180px',
      render: row => <ProgressCell campaign={row} />,
    },
    {
      key: 'cutoff_time',
      label: 'Cutoff',
      render: row => new Date(row.cutoff_time).toLocaleString(),
    },
  ];

  const rows = data?.data ?? [];
  const totalCount = data?.pagination.total_count ?? 0;

  return (
    <PageContainer
      title="Group Buy Campaigns"
      breadcrumbs={[{ label: 'Group Buys' }]}
      actions={
        canCreateCampaign ? (
          <Button
            variant="contained"
            size="small"
            startIcon={<AddIcon />}
            onClick={() => setCreateOpen(true)}
          >
            Create Campaign
          </Button>
        ) : null
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
        error={error ? 'Failed to load campaigns.' : null}
        page={page}
        pageSize={pageSize}
        totalCount={totalCount}
        onPageChange={setPage}
        onPageSizeChange={ps => { setPageSize(ps); setPage(0); }}
        emptyTitle="No campaigns found"
        emptyDescription={emptyDescription}
      />

      <CreateCampaignDialog
        open={createOpen}
        onClose={handleCreateDialogClose}
        defaultItemID={defaultItemID}
        itemReadOnly={memberStartingFromItem}
        title={memberStartingFromItem ? 'Start Group Buy' : 'Create Campaign'}
        submitLabel={memberStartingFromItem ? 'Start' : 'Create'}
        isOffline={isOffline}
      />
    </PageContainer>
  );
}
