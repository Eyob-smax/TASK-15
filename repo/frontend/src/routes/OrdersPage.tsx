import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import CircularProgress from '@mui/material/CircularProgress';
import Alert from '@mui/material/Alert';
import { OfflineDataNotice } from '@/components/OfflineDataNotice';
import { PageContainer } from '@/components/PageContainer';
import { DataTable, type Column } from '@/components/DataTable';
import { FilterBar, type FilterField } from '@/components/FilterBar';
import { StatusChip } from '@/components/StatusChip';
import { useAuth } from '@/lib/auth';
import { OFFLINE_MUTATION_MESSAGE, useOfflineStatus } from '@/lib/offline';
import { useOrderList, useMergeOrder } from '@/lib/hooks/useOrders';
import { useNotify } from '@/lib/notifications';
import type { Order } from '@/lib/types';

const FILTER_FIELDS: FilterField[] = [
  {
    key: 'status',
    label: 'Status',
    type: 'select',
    options: [
      { value: 'created', label: 'Created' },
      { value: 'paid', label: 'Paid' },
      { value: 'cancelled', label: 'Cancelled' },
      { value: 'refunded', label: 'Refunded' },
      { value: 'auto_closed', label: 'Auto-Closed' },
    ],
  },
];

function shortId(id: string) {
  return id.slice(0, 8) + '…';
}

function MergeOrdersDialog({
  open,
  selectedCount,
  loading,
  isOffline,
  onClose,
  onConfirm,
}: {
  open: boolean;
  selectedCount: number;
  loading: boolean;
  isOffline: boolean;
  onClose: () => void;
  onConfirm: () => void;
}) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="xs" fullWidth>
      <DialogTitle>Merge Orders</DialogTitle>
      <DialogContent>
        {isOffline && (
          <Alert severity="warning" sx={{ mt: 1, mb: 2 }}>{OFFLINE_MUTATION_MESSAGE}</Alert>
        )}
        <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
          Merge {selectedCount} selected orders into a single order?
        </Typography>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose} disabled={loading}>Cancel</Button>
        <Button
          variant="contained"
          onClick={onConfirm}
          disabled={loading || selectedCount < 2 || isOffline}
          startIcon={loading ? <CircularProgress size={16} color="inherit" /> : undefined}
        >
          Merge
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function isMergeEligible(order: Order): boolean {
  return order.status === 'created' || order.status === 'paid';
}

export default function OrdersPage() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const { isOffline } = useOfflineStatus();
  const notify = useNotify();
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [filters, setFilters] = useState<Record<string, string>>({});
  const [selectedOrderIDs, setSelectedOrderIDs] = useState<string[]>([]);
  const [mergeOpen, setMergeOpen] = useState(false);
  const mergeMutation = useMergeOrder();
  const isManageOrders = user?.role === 'administrator' || user?.role === 'operations_manager';

  const { data, isLoading, error, dataUpdatedAt } = useOrderList({
    page: page + 1,
    page_size: pageSize,
    status: filters.status || undefined,
  });

  const handleFiltersChange = useCallback((values: Record<string, string>) => {
    setFilters(values);
    setPage(0);
    setSelectedOrderIDs([]);
  }, []);

  const rows = data?.data ?? [];
  const totalCount = data?.pagination.total_count ?? 0;

  const selectableIDsOnPage = rows.filter(isMergeEligible).map(row => row.id);
  const allSelectableOnPageSelected = selectableIDsOnPage.length > 0 && selectableIDsOnPage.every(id => selectedOrderIDs.includes(id));

  const toggleSelection = useCallback((orderID: string) => {
    setSelectedOrderIDs(prev => (prev.includes(orderID) ? prev.filter(id => id !== orderID) : [...prev, orderID]));
  }, []);

  const toggleSelectAllOnPage = useCallback(() => {
    setSelectedOrderIDs(prev => {
      if (allSelectableOnPageSelected) {
        return prev.filter(id => !selectableIDsOnPage.includes(id));
      }
      return Array.from(new Set([...prev, ...selectableIDsOnPage]));
    });
  }, [allSelectableOnPageSelected, selectableIDsOnPage]);

  const handleMergeConfirm = async () => {
    if (selectedOrderIDs.length < 2) return;
    try {
      const merged = await mergeMutation.mutateAsync({ order_ids: selectedOrderIDs });
      notify.success('Orders merged successfully.');
      setSelectedOrderIDs([]);
      setMergeOpen(false);
      navigate(`/orders/${merged.id}`);
    } catch {
      notify.error('Failed to merge selected orders. Ensure selected orders share the same member and item.');
    }
  };

  const columns: Column<Order>[] = [
    {
      key: 'select',
      label: 'Select',
      align: 'center',
      render: row => (
        <Checkbox
          size="small"
          checked={selectedOrderIDs.includes(row.id)}
          onChange={() => toggleSelection(row.id)}
          disabled={!isManageOrders || !isMergeEligible(row)}
        />
      ),
      width: 72,
    },
    {
      key: 'id',
      label: 'Order ID',
      render: row => (
        <Typography
          variant="body2"
          component="span"
          sx={{ cursor: 'pointer', fontFamily: 'monospace', '&:hover': { textDecoration: 'underline' } }}
          onClick={() => navigate(`/orders/${row.id}`)}
        >
          {shortId(row.id)}
        </Typography>
      ),
    },
    {
      key: 'item_id',
      label: 'Item ID',
      render: row => (
        <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
          {shortId(row.item_id)}
        </Typography>
      ),
    },
    {
      key: 'quantity',
      label: 'Qty',
      align: 'right',
    },
    {
      key: 'total_amount',
      label: 'Total',
      align: 'right',
      render: row => `$${row.total_amount.toFixed(2)}`,
    },
    {
      key: 'status',
      label: 'Status',
      render: row => <StatusChip status={row.status} />,
    },
    {
      key: 'created_at',
      label: 'Created',
      render: row => new Date(row.created_at).toLocaleDateString(),
    },
  ];

  return (
    <PageContainer title="Orders" breadcrumbs={[{ label: 'Orders' }]}>
      <Box sx={{ mb: 2 }}>
        <FilterBar fields={FILTER_FIELDS} onChange={handleFiltersChange} />
      </Box>

      {isManageOrders && (
        <Box sx={{ mb: 2, display: 'flex', gap: 1, alignItems: 'center', flexWrap: 'wrap' }}>
          <Button
            variant="outlined"
            size="small"
            onClick={toggleSelectAllOnPage}
            disabled={selectableIDsOnPage.length === 0 || isOffline}
          >
            {allSelectableOnPageSelected ? 'Unselect Page' : 'Select Page'}
          </Button>
          <Button
            variant="text"
            size="small"
            onClick={() => setSelectedOrderIDs([])}
            disabled={selectedOrderIDs.length === 0}
          >
            Clear Selection
          </Button>
          <Button
            variant="contained"
            size="small"
            onClick={() => setMergeOpen(true)}
            disabled={selectedOrderIDs.length < 2 || isOffline}
          >
            Merge Selected ({selectedOrderIDs.length})
          </Button>
        </Box>
      )}

      <OfflineDataNotice hasData={rows.length > 0} dataUpdatedAt={dataUpdatedAt} />

      <DataTable
        columns={columns}
        rows={rows}
        rowKey={row => row.id}
        loading={isLoading}
        error={error ? 'Failed to load orders.' : null}
        page={page}
        pageSize={pageSize}
        totalCount={totalCount}
        onPageChange={nextPage => { setPage(nextPage); setSelectedOrderIDs([]); }}
        onPageSizeChange={ps => { setPageSize(ps); setPage(0); setSelectedOrderIDs([]); }}
        emptyTitle="No orders found"
        emptyDescription="Orders will appear here once created."
      />

      <MergeOrdersDialog
        open={mergeOpen}
        selectedCount={selectedOrderIDs.length}
        loading={mergeMutation.isPending}
        isOffline={isOffline}
        onClose={() => setMergeOpen(false)}
        onConfirm={handleMergeConfirm}
      />
    </PageContainer>
  );
}
