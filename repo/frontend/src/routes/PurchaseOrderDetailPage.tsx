import { useState } from 'react';
import { useParams } from 'react-router-dom';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Grid from '@mui/material/Grid';
import MenuItem from '@mui/material/MenuItem';
import Paper from '@mui/material/Paper';
import Skeleton from '@mui/material/Skeleton';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import { OfflineDataNotice } from '@/components/OfflineDataNotice';
import { PageContainer } from '@/components/PageContainer';
import { ConfirmDialog } from '@/components/ConfirmDialog';
import { StatusChip } from '@/components/StatusChip';
import { RequireRole } from '@/lib/auth';
import { useOfflineStatus } from '@/lib/offline';
import {
  usePO,
  useApprovePO,
  useReceivePO,
  useReturnPO,
  useVoidPO,
  useVarianceList,
  useResolveVariance,
} from '@/lib/hooks/useProcurement';
import { useNotify } from '@/lib/notifications';
import type { PurchaseOrderLine, VarianceRecord } from '@/lib/types';

function DetailRow({ label, value }: { label: string; value: string | number | null | undefined }) {
  return (
    <Box sx={{ py: 0.5 }}>
      <Typography variant="caption" color="text.secondary" display="block">
        {label}
      </Typography>
      <Typography variant="body2">{value ?? '—'}</Typography>
    </Box>
  );
}

function shortId(id: string | null | undefined): string {
  if (!id) return '—';
  return id.slice(0, 8) + '…';
}

interface ReceivedLine {
  lineId: string;
  quantity: number;
  unitPrice: number;
}

function ReceiveDialog({
  open,
  onClose,
  poId,
  lines,
  isOffline,
}: {
  open: boolean;
  onClose: () => void;
  poId: string;
  lines: PurchaseOrderLine[];
  isOffline: boolean;
}) {
  const notify = useNotify();
  const receiveMutation = useReceivePO();
  const [receivedLines, setReceivedLines] = useState<Record<string, ReceivedLine>>(() => {
    const init: Record<string, ReceivedLine> = {};
    lines.forEach(l => {
      init[l.id] = { lineId: l.id, quantity: l.ordered_quantity, unitPrice: l.ordered_unit_price };
    });
    return init;
  });

  const handleChange = (lineId: string, field: 'quantity' | 'unitPrice', value: string) => {
    setReceivedLines(prev => ({
      ...prev,
      [lineId]: { ...prev[lineId], [field]: parseFloat(value) || 0 },
    }));
  };

  const handleSubmit = async () => {
    try {
      const payload = Object.values(receivedLines).map(l => ({
        po_line_id: l.lineId,
        received_quantity: l.quantity,
        received_unit_price: l.unitPrice,
      }));
      await receiveMutation.mutateAsync({ id: poId, lines: payload });
      isOffline
        ? notify.info('Action queued — will sync when you reconnect.')
        : notify.success('Purchase order marked as received.');
      onClose();
    } catch {
      notify.error('Failed to receive purchase order.');
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>Receive Purchase Order</DialogTitle>
      <DialogContent>
        {isOffline && (
          <Alert severity="info" sx={{ mb: 2 }}>Offline — this action will be queued and applied when you reconnect.</Alert>
        )}
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          Enter the received quantities and unit prices for each line.
        </Typography>
        <TableContainer>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Item ID</TableCell>
                <TableCell align="right">Ordered Qty</TableCell>
                <TableCell align="right">Received Qty</TableCell>
                <TableCell align="right">Received Unit Price</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {lines.map(line => (
                <TableRow key={line.id}>
                  <TableCell>
                    <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                      {shortId(line.item_id)}
                    </Typography>
                  </TableCell>
                  <TableCell align="right">{line.ordered_quantity}</TableCell>
                  <TableCell align="right">
                    <TextField
                      type="number"
                      size="small"
                      value={receivedLines[line.id]?.quantity ?? line.ordered_quantity}
                      onChange={e => handleChange(line.id, 'quantity', e.target.value)}
                      inputProps={{ min: 0, style: { textAlign: 'right', width: 80 } }}
                    />
                  </TableCell>
                  <TableCell align="right">
                    <TextField
                      type="number"
                      size="small"
                      value={receivedLines[line.id]?.unitPrice ?? line.ordered_unit_price}
                      onChange={e => handleChange(line.id, 'unitPrice', e.target.value)}
                      inputProps={{ min: 0, step: 0.01, style: { textAlign: 'right', width: 100 } }}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose} disabled={receiveMutation.isPending}>Cancel</Button>
        <Button
          variant="contained"
          onClick={handleSubmit}
          disabled={receiveMutation.isPending}
          startIcon={receiveMutation.isPending ? <CircularProgress size={16} color="inherit" /> : undefined}
        >
          Confirm Receipt
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function ResolveVarianceDialog({
  open,
  onClose,
  varianceId,
  isOffline,
}: {
  open: boolean;
  onClose: () => void;
  varianceId: string;
  isOffline: boolean;
}) {
  const notify = useNotify();
  const resolveMutation = useResolveVariance();
  const [action, setAction] = useState<'adjustment' | 'return'>('adjustment');
  const [notes, setNotes] = useState('');
  const [quantityChange, setQuantityChange] = useState('');

  const handleSubmit = async () => {
    if (!notes.trim()) return;
    if (action === 'adjustment' && !quantityChange.trim()) return;
    try {
      await resolveMutation.mutateAsync({
        id: varianceId,
        action,
        resolution_notes: notes.trim(),
        quantity_change: action === 'adjustment' ? Number(quantityChange) : undefined,
      });
      isOffline
        ? notify.info('Action queued — will sync when you reconnect.')
        : notify.success('Variance resolved.');
      setAction('adjustment');
      setNotes('');
      setQuantityChange('');
      onClose();
    } catch {
      notify.error('Failed to resolve variance.');
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="xs" fullWidth>
      <DialogTitle>Resolve Variance</DialogTitle>
      <DialogContent>
        {isOffline && (
          <Alert severity="info" sx={{ mt: 1, mb: 2 }}>Offline — this action will be queued and applied when you reconnect.</Alert>
        )}
        <TextField
          select
          label="Resolution Action"
          value={action}
          onChange={e => setAction(e.target.value as 'adjustment' | 'return')}
          fullWidth
          size="small"
          sx={{ mt: 1, mb: 2 }}
        >
          <MenuItem value="adjustment">Adjustment</MenuItem>
          <MenuItem value="return">Return</MenuItem>
        </TextField>
        {action === 'adjustment' && (
          <TextField
            label="Quantity Change"
            value={quantityChange}
            onChange={e => setQuantityChange(e.target.value)}
            type="number"
            fullWidth
            size="small"
            sx={{ mb: 2 }}
          />
        )}
        <TextField
          label="Resolution Notes"
          value={notes}
          onChange={e => setNotes(e.target.value)}
          fullWidth
          size="small"
          multiline
          rows={3}
          sx={{ mt: 1 }}
          required
        />
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose} disabled={resolveMutation.isPending}>Cancel</Button>
        <Button
          variant="contained"
          onClick={handleSubmit}
          disabled={resolveMutation.isPending || !notes.trim() || (action === 'adjustment' && !quantityChange.trim())}
          startIcon={resolveMutation.isPending ? <CircularProgress size={16} color="inherit" /> : undefined}
        >
          Resolve
        </Button>
      </DialogActions>
    </Dialog>
  );
}

const TERMINAL_STATUSES = ['received', 'returned', 'voided'];

export default function PurchaseOrderDetailPage() {
  const { id } = useParams<{ id: string }>();
  const notify = useNotify();
  const { isOffline } = useOfflineStatus();

  const [receiveOpen, setReceiveOpen] = useState(false);
  const [voidOpen, setVoidOpen] = useState(false);
  const [returnOpen, setReturnOpen] = useState(false);
  const [resolveVarianceId, setResolveVarianceId] = useState<string | null>(null);

  const { data: poData, isLoading, error, dataUpdatedAt } = usePO(id);
  const approveMutation = useApprovePO();
  const returnMutation = useReturnPO();
  const voidMutation = useVoidPO();

  // Variance list — fetch all and filter by PO if needed
  const { data: varianceData } = useVarianceList();

  const handleApprove = async () => {
    if (!id) return;
    try {
      await approveMutation.mutateAsync(id);
      isOffline
        ? notify.info('Action queued — will sync when you reconnect.')
        : notify.success('Purchase order approved.');
    } catch {
      notify.error('Failed to approve purchase order.');
    }
  };

  const handleReturn = async () => {
    if (!id) return;
    try {
      await returnMutation.mutateAsync(id);
      isOffline
        ? notify.info('Action queued — will sync when you reconnect.')
        : notify.success('Purchase order returned.');
      setReturnOpen(false);
    } catch {
      notify.error('Failed to return purchase order.');
    }
  };

  const handleVoid = async () => {
    if (!id) return;
    try {
      await voidMutation.mutateAsync(id);
      isOffline
        ? notify.info('Action queued — will sync when you reconnect.')
        : notify.success('Purchase order voided.');
      setVoidOpen(false);
    } catch {
      notify.error('Failed to void purchase order.');
    }
  };

  if (isLoading) {
    return (
      <PageContainer
        title="Purchase Order"
        breadcrumbs={[
          { label: 'Procurement' },
          { label: 'Purchase Orders', to: '/procurement/purchase-orders' },
          { label: id ? shortId(id) : '…' },
        ]}
      >
        <Skeleton variant="rectangular" height={300} />
      </PageContainer>
    );
  }

  if (!poData) {
    return (
      <PageContainer
        title="Purchase Order"
        breadcrumbs={[
          { label: 'Procurement' },
          { label: 'Purchase Orders', to: '/procurement/purchase-orders' },
          { label: id ? shortId(id) : '…' },
        ]}
      >
        <Alert severity="error">Failed to load purchase order details.</Alert>
      </PageContainer>
    );
  }

  const po = poData;
  const lines: PurchaseOrderLine[] = poData.lines ?? [];
  const poLineIDs = new Set(lines.map(line => line.id));
  const variances: VarianceRecord[] = (Array.isArray(varianceData)
    ? varianceData
    : varianceData?.data ?? []).filter(variance => poLineIDs.has(variance.po_line_id));

  const isTerminal = TERMINAL_STATUSES.includes(po.status);

  return (
    <PageContainer
      title="Purchase Order"
      breadcrumbs={[
        { label: 'Procurement' },
        { label: 'Purchase Orders', to: '/procurement/purchase-orders' },
        { label: shortId(po.id) },
      ]}
      actions={
        <RequireRole roles={['administrator', 'procurement_specialist']}>
          <Box sx={{ display: 'flex', gap: 1 }}>
            {po.status === 'created' && (
              <Button
                variant="contained"
                size="small"
                color="success"
                onClick={handleApprove}
                disabled={approveMutation.isPending}
                startIcon={approveMutation.isPending ? <CircularProgress size={16} color="inherit" /> : undefined}
              >
                Approve
              </Button>
            )}
            {po.status === 'approved' && (
              <Button variant="contained" size="small" onClick={() => setReceiveOpen(true)}>
                Receive
              </Button>
            )}
            {po.status === 'received' && (
              <Button variant="outlined" size="small" color="warning" onClick={() => setReturnOpen(true)}>
                Return
              </Button>
            )}
            {!isTerminal && (
              <Button variant="outlined" size="small" color="error" onClick={() => setVoidOpen(true)}>
                Void
              </Button>
            )}
          </Box>
        </RequireRole>
      }
    >
      <OfflineDataNotice hasData={Boolean(poData)} dataUpdatedAt={dataUpdatedAt} />

      {error && (
        <Alert severity="warning" sx={{ mb: 3 }}>
          Purchase order sync is temporarily unavailable. Showing the latest cached details when possible.
        </Alert>
      )}

      <Grid container spacing={3}>
        {/* PO Info Card */}
        <Grid item xs={12} md={6}>
          <Paper variant="outlined" sx={{ p: 3 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
              <Typography variant="subtitle1" fontWeight={600}>Order Details</Typography>
              <StatusChip status={po.status} />
            </Box>
            <Grid container spacing={2}>
              <Grid item xs={6}>
                <DetailRow label="Supplier ID" value={shortId(po.supplier_id)} />
                <DetailRow label="Created By" value={shortId(po.created_by)} />
                <DetailRow label="Approved By" value={shortId(po.approved_by)} />
              </Grid>
              <Grid item xs={6}>
                <DetailRow
                  label="Total Amount"
                  value={`$${po.total_amount.toFixed(2)}`}
                />
                <DetailRow label="Created At" value={po.created_at ? new Date(po.created_at).toLocaleString() : undefined} />
                <DetailRow label="Approved At" value={po.approved_at ? new Date(po.approved_at).toLocaleString() : undefined} />
                <DetailRow label="Received At" value={po.received_at ? new Date(po.received_at).toLocaleString() : undefined} />
              </Grid>
            </Grid>
          </Paper>
        </Grid>

        {/* Lines Table */}
        <Grid item xs={12}>
          <Paper variant="outlined">
            <Box sx={{ px: 2, py: 1.5, borderBottom: 1, borderColor: 'divider' }}>
              <Typography variant="subtitle1" fontWeight={600}>Order Lines</Typography>
            </Box>
            <TableContainer>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Item ID</TableCell>
                    <TableCell align="right">Ordered Qty</TableCell>
                    <TableCell align="right">Ordered Unit Price</TableCell>
                    <TableCell align="right">Received Qty</TableCell>
                    <TableCell align="right">Received Unit Price</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {lines.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={5}>
                        <Typography variant="body2" color="text.secondary" sx={{ py: 1 }}>
                          No lines found.
                        </Typography>
                      </TableCell>
                    </TableRow>
                  ) : (
                    lines.map(line => (
                      <TableRow key={line.id} hover>
                        <TableCell>
                          <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                            {shortId(line.item_id)}
                          </Typography>
                        </TableCell>
                        <TableCell align="right">{line.ordered_quantity}</TableCell>
                        <TableCell align="right">${line.ordered_unit_price.toFixed(2)}</TableCell>
                        <TableCell align="right">{line.received_quantity ?? 0}</TableCell>
                        <TableCell align="right">
                          {line.received_unit_price != null ? `$${line.received_unit_price.toFixed(2)}` : '—'}
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          </Paper>
        </Grid>

        {/* Variances Section */}
        <Grid item xs={12}>
          <Paper variant="outlined">
            <Box sx={{ px: 2, py: 1.5, borderBottom: 1, borderColor: 'divider' }}>
              <Typography variant="subtitle1" fontWeight={600}>Variances</Typography>
            </Box>
            <TableContainer>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Type</TableCell>
                    <TableCell align="right">Expected</TableCell>
                    <TableCell align="right">Actual</TableCell>
                    <TableCell>Status</TableCell>
                    <TableCell>Overdue</TableCell>
                    <TableCell />
                  </TableRow>
                </TableHead>
                <TableBody>
                  {variances.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={6}>
                        <Typography variant="body2" color="text.secondary" sx={{ py: 1 }}>
                          No variances recorded.
                        </Typography>
                      </TableCell>
                    </TableRow>
                  ) : (
                    variances.map(v => (
                      <TableRow key={v.id} hover>
                        <TableCell>{v.type}</TableCell>
                        <TableCell align="right">{v.expected_value}</TableCell>
                        <TableCell align="right">{v.actual_value}</TableCell>
                        <TableCell>
                          <StatusChip status={v.status} />
                        </TableCell>
                        <TableCell>
                          {v.is_overdue ? (
                            <Chip label="Overdue" color="error" size="small" />
                          ) : (
                            <Chip label="No" size="small" variant="outlined" />
                          )}
                        </TableCell>
                        <TableCell>
                          {(v.status === 'open' || v.status === 'escalated') && (
                            <RequireRole roles={['administrator', 'procurement_specialist']}>
                              <Button
                                size="small"
                                variant="outlined"
                                onClick={() => setResolveVarianceId(v.id)}
                              >
                                Resolve
                              </Button>
                            </RequireRole>
                          )}
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          </Paper>
        </Grid>
      </Grid>

      {/* Dialogs */}
      {id && po.status === 'approved' && (
        <ReceiveDialog
          open={receiveOpen}
          onClose={() => setReceiveOpen(false)}
          poId={id}
          lines={lines}
          isOffline={isOffline}
        />
      )}

      <ConfirmDialog
        open={returnOpen}
        title="Return Purchase Order"
        message="Return this purchase order? This will reverse the received status."
        confirmLabel="Return"
        destructive
        loading={returnMutation.isPending}
        onConfirm={handleReturn}
        onCancel={() => setReturnOpen(false)}
      />

      <ConfirmDialog
        open={voidOpen}
        title="Void Purchase Order"
        message="Void this purchase order? This action cannot be undone."
        confirmLabel="Void"
        destructive
        loading={voidMutation.isPending}
        onConfirm={handleVoid}
        onCancel={() => setVoidOpen(false)}
      />

      {resolveVarianceId && (
        <ResolveVarianceDialog
          open={!!resolveVarianceId}
          onClose={() => setResolveVarianceId(null)}
          varianceId={resolveVarianceId}
          isOffline={isOffline}
        />
      )}
    </PageContainer>
  );
}
