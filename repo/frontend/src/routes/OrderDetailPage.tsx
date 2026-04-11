import { useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";
import Dialog from "@mui/material/Dialog";
import DialogTitle from "@mui/material/DialogTitle";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import Divider from "@mui/material/Divider";
import Grid from "@mui/material/Grid";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import Paper from "@mui/material/Paper";
import Skeleton from "@mui/material/Skeleton";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import { OfflineDataNotice } from "@/components/OfflineDataNotice";
import { PageContainer } from "@/components/PageContainer";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { StatusChip } from "@/components/StatusChip";
import { RequireRole, useAuth } from "@/lib/auth";
import { OFFLINE_MUTATION_MESSAGE, useOfflineStatus } from "@/lib/offline";
import {
  useOrder,
  useOrderTimeline,
  useCancelOrder,
  usePayOrder,
  useRefundOrder,
  useAddOrderNote,
  useSplitOrder,
} from "@/lib/hooks/useOrders";
import { useNotify } from "@/lib/notifications";
import type { Order } from "@/lib/types";

function DetailRow({
  label,
  value,
}: {
  label: string;
  value: string | number | null | undefined;
}) {
  return (
    <Box sx={{ py: 0.5 }}>
      <Typography variant="caption" color="text.secondary" display="block">
        {label}
      </Typography>
      <Typography variant="body2">{value ?? "—"}</Typography>
    </Box>
  );
}

function AddNoteDialog({
  open,
  onClose,
  orderId,
  isOffline,
}: {
  open: boolean;
  onClose: () => void;
  orderId: string;
  isOffline: boolean;
}) {
  const notify = useNotify();
  const addNoteMutation = useAddOrderNote();
  const [note, setNote] = useState("");

  const handleSubmit = async () => {
    if (!note.trim()) return;
    try {
      await addNoteMutation.mutateAsync({ id: orderId, note });
      notify.success("Note added.");
      setNote("");
      onClose();
    } catch {
      notify.error("Failed to add note.");
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="xs" fullWidth>
      <DialogTitle>Add Note</DialogTitle>
      <DialogContent>
        {isOffline && (
          <Alert severity="warning" sx={{ mt: 1, mb: 2 }}>
            {OFFLINE_MUTATION_MESSAGE}
          </Alert>
        )}
        <TextField
          label="Note"
          value={note}
          onChange={(e) => setNote(e.target.value)}
          fullWidth
          size="small"
          multiline
          rows={3}
          sx={{ mt: 1 }}
        />
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose} disabled={addNoteMutation.isPending}>
          Cancel
        </Button>
        <Button
          variant="contained"
          onClick={handleSubmit}
          disabled={addNoteMutation.isPending || !note.trim() || isOffline}
          startIcon={
            addNoteMutation.isPending ? (
              <CircularProgress size={16} color="inherit" />
            ) : undefined
          }
        >
          Add Note
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function PayDialog({
  open,
  onClose,
  orderId,
  isOffline,
}: {
  open: boolean;
  onClose: () => void;
  orderId: string;
  isOffline: boolean;
}) {
  const notify = useNotify();
  const payMutation = usePayOrder();
  const [marker, setMarker] = useState("");

  const handleSubmit = async () => {
    if (!marker.trim()) return;
    try {
      await payMutation.mutateAsync({ id: orderId, settlementMarker: marker });
      notify.success("Order marked as paid.");
      setMarker("");
      onClose();
    } catch {
      notify.error("Failed to mark order as paid.");
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="xs" fullWidth>
      <DialogTitle>Record Payment</DialogTitle>
      <DialogContent>
        {isOffline && (
          <Alert severity="warning" sx={{ mt: 1, mb: 2 }}>
            {OFFLINE_MUTATION_MESSAGE}
          </Alert>
        )}
        <TextField
          label="Settlement Marker"
          value={marker}
          onChange={(e) => setMarker(e.target.value)}
          fullWidth
          size="small"
          sx={{ mt: 1 }}
          placeholder="Transaction reference or marker"
        />
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose} disabled={payMutation.isPending}>
          Cancel
        </Button>
        <Button
          variant="contained"
          onClick={handleSubmit}
          disabled={payMutation.isPending || !marker.trim() || isOffline}
          startIcon={
            payMutation.isPending ? (
              <CircularProgress size={16} color="inherit" />
            ) : undefined
          }
        >
          Record Payment
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function parseSplitQuantities(input: string): number[] {
  return input
    .split(",")
    .map((value) => Number.parseInt(value.trim(), 10))
    .filter((value) => Number.isFinite(value));
}

function SplitDialog({
  open,
  onClose,
  order,
  isOffline,
}: {
  open: boolean;
  onClose: () => void;
  order: Order;
  isOffline: boolean;
}) {
  const notify = useNotify();
  const splitMutation = useSplitOrder();
  const [quantitiesText, setQuantitiesText] = useState("");

  useEffect(() => {
    if (!open) return;
    if (order.quantity > 1) {
      setQuantitiesText(`1,${order.quantity - 1}`);
      return;
    }
    setQuantitiesText(String(order.quantity));
  }, [open, order.quantity]);

  const quantities = useMemo(
    () => parseSplitQuantities(quantitiesText),
    [quantitiesText],
  );
  const hasAtLeastTwo = quantities.length >= 2;
  const allPositive = quantities.every((quantity) => quantity > 0);
  const total = quantities.reduce((sum, quantity) => sum + quantity, 0);
  const isValid = hasAtLeastTwo && allPositive && total === order.quantity;

  let helperText = `Enter at least 2 positive integers. Sum must equal ${order.quantity}.`;
  if (quantitiesText.trim()) {
    if (!hasAtLeastTwo) {
      helperText = "Provide at least 2 quantities.";
    } else if (!allPositive) {
      helperText = "Quantities must be positive integers.";
    } else if (total !== order.quantity) {
      helperText = `Current total is ${total}; expected ${order.quantity}.`;
    }
  }

  const handleSubmit = async () => {
    if (!isValid) return;
    try {
      await splitMutation.mutateAsync({ id: order.id, quantities });
      notify.success("Order split successfully.");
      onClose();
    } catch {
      notify.error("Failed to split order.");
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="xs" fullWidth>
      <DialogTitle>Split Order</DialogTitle>
      <DialogContent>
        {isOffline && (
          <Alert severity="warning" sx={{ mt: 1, mb: 2 }}>
            {OFFLINE_MUTATION_MESSAGE}
          </Alert>
        )}
        <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
          Original quantity: {order.quantity}
        </Typography>
        <TextField
          autoFocus
          margin="dense"
          label="Split Quantities"
          placeholder="e.g. 1,2,3"
          value={quantitiesText}
          onChange={(e) => setQuantitiesText(e.target.value)}
          fullWidth
          size="small"
          error={Boolean(quantitiesText.trim()) && !isValid}
          helperText={helperText}
        />
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose} disabled={splitMutation.isPending}>
          Cancel
        </Button>
        <Button
          variant="contained"
          onClick={handleSubmit}
          disabled={splitMutation.isPending || !isValid || isOffline}
          startIcon={
            splitMutation.isPending ? (
              <CircularProgress size={16} color="inherit" />
            ) : undefined
          }
        >
          Split
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default function OrderDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const { isOffline } = useOfflineStatus();
  const notify = useNotify();

  const [cancelOpen, setCancelOpen] = useState(false);
  const [refundOpen, setRefundOpen] = useState(false);
  const [payOpen, setPayOpen] = useState(false);
  const [splitOpen, setSplitOpen] = useState(false);
  const [noteOpen, setNoteOpen] = useState(false);

  const { data: order, isLoading, error, dataUpdatedAt } = useOrder(id);
  const { data: timeline = [], isLoading: timelineLoading } =
    useOrderTimeline(id);
  const cancelMutation = useCancelOrder();
  const refundMutation = useRefundOrder();

  const isManageOrders =
    user?.role === "administrator" || user?.role === "operations_manager";
  const isOwnOrder = order?.user_id === user?.id;
  const orderStatus = order?.status;
  const orderQuantity = order?.quantity ?? 0;

  const canCancel =
    Boolean(order) &&
    ((isManageOrders &&
      (orderStatus === "created" || orderStatus === "paid")) ||
      (!isManageOrders && isOwnOrder && orderStatus === "created"));

  const canSplit =
    Boolean(order) &&
    isManageOrders &&
    orderStatus !== "cancelled" &&
    orderStatus !== "refunded" &&
    orderStatus !== "auto_closed" &&
    orderQuantity > 1;

  const handleCancel = async () => {
    if (!id || isOffline) return;
    try {
      await cancelMutation.mutateAsync(id);
      notify.success("Order cancelled.");
      setCancelOpen(false);
    } catch {
      notify.error("Failed to cancel order.");
    }
  };

  const handleRefund = async () => {
    if (!id || isOffline) return;
    try {
      await refundMutation.mutateAsync(id);
      notify.success("Order refunded.");
      setRefundOpen(false);
    } catch {
      notify.error("Failed to refund order.");
    }
  };

  if (isLoading) {
    return (
      <PageContainer
        title="Order Details"
        breadcrumbs={[{ label: "Orders", to: "/orders" }, { label: "Details" }]}
      >
        <Skeleton variant="rectangular" height={300} />
      </PageContainer>
    );
  }

  if (!order) {
    return (
      <PageContainer
        title="Order Details"
        breadcrumbs={[{ label: "Orders", to: "/orders" }, { label: "Details" }]}
      >
        <Alert severity="error">Failed to load order details.</Alert>
      </PageContainer>
    );
  }

  const shortId = (value: string) => value.slice(0, 8) + "…";

  return (
    <PageContainer
      title={`Order ${shortId(order.id)}`}
      breadcrumbs={[
        { label: "Orders", to: "/orders" },
        { label: shortId(order.id) },
      ]}
      actions={
        <Box sx={{ display: "flex", gap: 1 }}>
          {canSplit && (
            <Button
              variant="outlined"
              size="small"
              onClick={() => setSplitOpen(true)}
              disabled={isOffline}
            >
              Split
            </Button>
          )}
          {canCancel && (
            <Button
              variant="outlined"
              size="small"
              color="error"
              onClick={() => setCancelOpen(true)}
              disabled={isOffline}
            >
              Cancel Order
            </Button>
          )}
          <RequireRole roles={["administrator", "operations_manager"]}>
            {order.status === "paid" && (
              <Button
                variant="outlined"
                size="small"
                onClick={() => setRefundOpen(true)}
                disabled={isOffline}
              >
                Refund
              </Button>
            )}
            {order.status === "created" && (
              <Button
                variant="contained"
                size="small"
                color="success"
                onClick={() => setPayOpen(true)}
                disabled={isOffline}
              >
                Record Payment
              </Button>
            )}
            <Button
              variant="outlined"
              size="small"
              onClick={() => setNoteOpen(true)}
              disabled={isOffline}
            >
              Add Note
            </Button>
          </RequireRole>
        </Box>
      }
    >
      <OfflineDataNotice
        hasData={Boolean(order)}
        dataUpdatedAt={dataUpdatedAt}
      />

      {error && (
        <Alert severity="warning" sx={{ mb: 3 }}>
          Order sync is temporarily unavailable. Showing the latest cached
          details when possible.
        </Alert>
      )}

      <Grid container spacing={3}>
        <Grid item xs={12} md={7}>
          <Paper variant="outlined" sx={{ p: 3 }}>
            <Box sx={{ display: "flex", alignItems: "center", gap: 2, mb: 2 }}>
              <Typography variant="subtitle1" fontWeight={600}>
                Order Details
              </Typography>
              <StatusChip status={order.status} />
            </Box>

            <Grid container spacing={2}>
              <Grid item xs={6}>
                <DetailRow label="Order ID" value={order.id} />
                <DetailRow label="Item ID" value={order.item_id} />
                <DetailRow label="Campaign ID" value={order.campaign_id} />
                <DetailRow label="Quantity" value={order.quantity} />
              </Grid>
              <Grid item xs={6}>
                <DetailRow
                  label="Unit Price"
                  value={`$${order.unit_price.toFixed(2)}`}
                />
                <DetailRow
                  label="Total"
                  value={`$${order.total_amount.toFixed(2)}`}
                />
                <DetailRow
                  label="Auto-close At"
                  value={
                    order.auto_close_at
                      ? new Date(order.auto_close_at).toLocaleString()
                      : undefined
                  }
                />
              </Grid>
              <Grid item xs={12}>
                <DetailRow
                  label="Settlement Marker"
                  value={order.settlement_marker || undefined}
                />
                {order.notes && <DetailRow label="Notes" value={order.notes} />}
                {order.paid_at && (
                  <DetailRow
                    label="Paid At"
                    value={new Date(order.paid_at).toLocaleString()}
                  />
                )}
                {order.cancelled_at && (
                  <DetailRow
                    label="Cancelled At"
                    value={new Date(order.cancelled_at).toLocaleString()}
                  />
                )}
                {order.refunded_at && (
                  <DetailRow
                    label="Refunded At"
                    value={new Date(order.refunded_at).toLocaleString()}
                  />
                )}
              </Grid>
            </Grid>
          </Paper>
        </Grid>

        <Grid item xs={12} md={5}>
          <Paper variant="outlined" sx={{ p: 3 }}>
            <Typography variant="subtitle1" fontWeight={600} gutterBottom>
              Timeline
            </Typography>
            <Divider sx={{ mb: 1 }} />

            {timelineLoading ? (
              <Box>
                {[1, 2, 3].map((value) => (
                  <Skeleton key={value} height={40} />
                ))}
              </Box>
            ) : timeline.length === 0 ? (
              <Typography variant="body2" color="text.secondary">
                No timeline entries yet.
              </Typography>
            ) : (
              <List dense disablePadding>
                {timeline.map((entry) => (
                  <ListItem
                    key={entry.id}
                    disableGutters
                    sx={{ alignItems: "flex-start" }}
                  >
                    <ListItemText
                      primary={
                        <Box
                          sx={{
                            display: "flex",
                            justifyContent: "space-between",
                          }}
                        >
                          <Typography
                            variant="body2"
                            fontWeight={500}
                            textTransform="capitalize"
                          >
                            {entry.action.replace(/_/g, " ")}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {new Date(entry.created_at).toLocaleString()}
                          </Typography>
                        </Box>
                      }
                      secondary={entry.description}
                    />
                  </ListItem>
                ))}
              </List>
            )}
          </Paper>
        </Grid>
      </Grid>

      <ConfirmDialog
        open={cancelOpen}
        title="Cancel Order"
        message="Cancel this order? Inventory will be released."
        confirmLabel="Cancel Order"
        destructive
        loading={cancelMutation.isPending}
        onConfirm={handleCancel}
        onCancel={() => setCancelOpen(false)}
      />

      <ConfirmDialog
        open={refundOpen}
        title="Refund Order"
        message="Issue a refund for this order? Inventory will be released."
        confirmLabel="Refund"
        loading={refundMutation.isPending}
        onConfirm={handleRefund}
        onCancel={() => setRefundOpen(false)}
      />

      <PayDialog
        open={payOpen}
        onClose={() => setPayOpen(false)}
        orderId={order.id}
        isOffline={isOffline}
      />
      <SplitDialog
        open={splitOpen}
        onClose={() => setSplitOpen(false)}
        order={order}
        isOffline={isOffline}
      />
      <AddNoteDialog
        open={noteOpen}
        onClose={() => setNoteOpen(false)}
        orderId={order.id}
        isOffline={isOffline}
      />
    </PageContainer>
  );
}
