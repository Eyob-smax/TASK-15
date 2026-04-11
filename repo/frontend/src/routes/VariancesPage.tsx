import { useState, useCallback } from 'react';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import { PageContainer } from '@/components/PageContainer';
import { DataTable, type Column } from '@/components/DataTable';
import { FilterBar, type FilterField } from '@/components/FilterBar';
import { StatusChip } from '@/components/StatusChip';
import { RequireRole } from '@/lib/auth';
import { useVarianceList, useResolveVariance } from '@/lib/hooks/useProcurement';
import type { VarianceRecord } from '@/lib/types';

const FILTER_FIELDS: FilterField[] = [
  {
    key: 'status',
    label: 'Status',
    type: 'select',
    options: [
      { value: 'open', label: 'Open' },
      { value: 'resolved', label: 'Resolved' },
      { value: 'escalated', label: 'Escalated' },
    ],
  },
];

function shortId(id: string) {
  return id.slice(0, 8) + '\u2026';
}

function isOverdue(v: VarianceRecord): boolean {
  return v.status === 'open' && v.is_overdue;
}

function canResolve(v: VarianceRecord): boolean {
  return v.status === 'open' || v.status === 'escalated';
}

export default function VariancesPage() {
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [filters, setFilters] = useState<Record<string, string>>({});

  const [resolveTarget, setResolveTarget] = useState<VarianceRecord | null>(null);
  const [resolutionAction, setResolutionAction] = useState<'adjustment' | 'return'>('adjustment');
  const [resolutionNotes, setResolutionNotes] = useState('');
  const [quantityChange, setQuantityChange] = useState('');
  const [resolveError, setResolveError] = useState<string | null>(null);

  const { data, isLoading, error } = useVarianceList({
    page: page + 1,
    page_size: pageSize,
    status: filters.status || undefined,
  });

  const resolveMutation = useResolveVariance();

  const handleFiltersChange = useCallback((values: Record<string, string>) => {
    setFilters(values);
    setPage(0);
  }, []);

  const openResolve = useCallback((v: VarianceRecord) => {
    setResolveTarget(v);
    setResolutionAction('adjustment');
    setResolutionNotes('');
    setQuantityChange('');
    setResolveError(null);
  }, []);

  const handleResolveSubmit = async () => {
    if (!resolveTarget) return;
    try {
      await resolveMutation.mutateAsync({
        id: resolveTarget.id,
        action: resolutionAction,
        resolution_notes: resolutionNotes.trim(),
        quantity_change: resolutionAction === 'adjustment' ? Number(quantityChange) : undefined,
      });
      setResolveTarget(null);
    } catch (e: unknown) {
      setResolveError(e instanceof Error ? e.message : 'Failed to resolve variance');
    }
  };

  const columns: Column<VarianceRecord>[] = [
    {
      key: 'id',
      label: 'ID',
      render: (v) => (
        <Typography variant="body2" fontFamily="monospace">
          {shortId(v.id)}
        </Typography>
      ),
    },
    {
      key: 'po_line_id',
      label: 'PO Line',
      render: (v) => (
        <Typography variant="body2" fontFamily="monospace">
          {shortId(v.po_line_id)}
        </Typography>
      ),
    },
    {
      key: 'type',
      label: 'Type',
      render: (v) => <StatusChip status={v.type} />,
    },
    {
      key: 'expected_value',
      label: 'Expected',
      render: (v) => v.expected_value.toFixed(2),
    },
    {
      key: 'actual_value',
      label: 'Actual',
      render: (v) => v.actual_value.toFixed(2),
    },
    {
      key: 'difference_amount',
      label: 'Variance',
      render: (v) => v.difference_amount.toFixed(2),
    },
    {
      key: 'status',
      label: 'Status',
      render: (v) => (
        <Box sx={{ display: 'flex', gap: 1, alignItems: 'center' }}>
          <StatusChip status={v.status} />
          {isOverdue(v) && (
            <Chip label="Overdue" color="error" size="small" />
          )}
        </Box>
      ),
    },
    {
      key: 'actions',
      label: '',
      render: (v) =>
        canResolve(v) ? (
          <RequireRole roles={['administrator', 'operations_manager', 'procurement_specialist']}>
            <Button size="small" variant="outlined" onClick={() => openResolve(v)}>
              Resolve
            </Button>
          </RequireRole>
        ) : null,
    },
  ];

  return (
    <PageContainer
      title="Variances"
      breadcrumbs={[
        { label: 'Procurement', to: '/procurement' },
        { label: 'Variances' },
      ]}
    >
      <FilterBar fields={FILTER_FIELDS} onChange={handleFiltersChange} />

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          Failed to load variances
        </Alert>
      )}

      <DataTable
        columns={columns}
        rows={data?.data ?? []}
        rowKey={(row) => row.id}
        loading={isLoading}
        page={page}
        pageSize={pageSize}
        totalCount={data?.pagination?.total_count ?? 0}
        onPageChange={setPage}
        onPageSizeChange={(s) => { setPageSize(s); setPage(0); }}
        emptyTitle="No variances found"
        emptyDescription="Received procurement variances will appear here for review."
      />

      {/* Resolve dialog */}
      <Dialog open={Boolean(resolveTarget)} onClose={() => setResolveTarget(null)} maxWidth="sm" fullWidth>
        <DialogTitle>Resolve Variance</DialogTitle>
        <DialogContent>
          {resolveError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {resolveError}
            </Alert>
          )}
          <TextField
            select
            label="Resolution Action"
            value={resolutionAction}
            onChange={(e) => setResolutionAction(e.target.value as 'adjustment' | 'return')}
            fullWidth
            sx={{ mb: 2, mt: 1 }}
          >
            <MenuItem value="adjustment">Adjustment</MenuItem>
            <MenuItem value="return">Return</MenuItem>
          </TextField>
          {resolutionAction === 'adjustment' && (
            <TextField
              label="Quantity Change"
              value={quantityChange}
              onChange={(e) => setQuantityChange(e.target.value)}
              type="number"
              fullWidth
              sx={{ mb: 2 }}
            />
          )}
          <TextField
            label="Resolution Notes"
            value={resolutionNotes}
            onChange={(e) => setResolutionNotes(e.target.value)}
            fullWidth
            multiline
            rows={4}
            helperText="Explain the procurement adjustment or return decision."
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setResolveTarget(null)}>Cancel</Button>
          <Button
            variant="contained"
            onClick={handleResolveSubmit}
            disabled={
              resolveMutation.isPending ||
              !resolutionNotes.trim() ||
              (resolutionAction === 'adjustment' && quantityChange.trim() === '')
            }
          >
            Resolve
          </Button>
        </DialogActions>
      </Dialog>
    </PageContainer>
  );
}
