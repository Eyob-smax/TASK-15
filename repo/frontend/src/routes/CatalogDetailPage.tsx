import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Divider from '@mui/material/Divider';
import Grid from '@mui/material/Grid';
import Paper from '@mui/material/Paper';
import Skeleton from '@mui/material/Skeleton';
import Typography from '@mui/material/Typography';
import Alert from '@mui/material/Alert';
import EditIcon from '@mui/icons-material/Edit';
import PublishIcon from '@mui/icons-material/Publish';
import UnpublishedIcon from '@mui/icons-material/Unpublished';
import { OfflineDataNotice } from '@/components/OfflineDataNotice';
import { PageContainer } from '@/components/PageContainer';
import { ConfirmDialog } from '@/components/ConfirmDialog';
import { StatusChip } from '@/components/StatusChip';
import { RequireRole } from '@/lib/auth';
import { useOfflineStatus } from '@/lib/offline';
import { useItem, usePublishItem, useUnpublishItem } from '@/lib/hooks/useItems';
import { useNotify } from '@/lib/notifications';
import { CONDITION_LABELS, BILLING_MODEL_LABELS } from '@/lib/constants';

function DetailRow({ label, value }: { label: string; value: string | number | null | undefined }) {
  return (
    <Box sx={{ py: 0.75 }}>
      <Typography variant="caption" color="text.secondary" display="block">
        {label}
      </Typography>
      <Typography variant="body2">{value ?? '—'}</Typography>
    </Box>
  );
}

function formatWindowLabel(startTime: string, endTime: string) {
  return `${new Date(startTime).toLocaleString()} - ${new Date(endTime).toLocaleString()}`;
}

export default function CatalogDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const notify = useNotify();
  const { isOffline } = useOfflineStatus();

  const [publishOpen, setPublishOpen] = useState(false);
  const [unpublishOpen, setUnpublishOpen] = useState(false);

  const { data: item, isLoading, error, dataUpdatedAt } = useItem(id);
  const publishMutation = usePublishItem();
  const unpublishMutation = useUnpublishItem();

  const handlePublish = async () => {
    if (!id || isOffline) return;
    try {
      await publishMutation.mutateAsync(id);
      notify.success('Item published successfully.');
      setPublishOpen(false);
    } catch {
      notify.error('Failed to publish item.');
    }
  };

  const handleUnpublish = async () => {
    if (!id || isOffline) return;
    try {
      await unpublishMutation.mutateAsync(id);
      notify.success('Item unpublished.');
      setUnpublishOpen(false);
    } catch {
      notify.error('Failed to unpublish item.');
    }
  };

  if (isLoading) {
    return (
      <PageContainer title="Item Details" breadcrumbs={[{ label: 'Catalog', to: '/catalog' }, { label: 'Details' }]}>
        <Skeleton variant="rectangular" height={300} />
      </PageContainer>
    );
  }

  if (!item) {
    return (
      <PageContainer title="Item Details" breadcrumbs={[{ label: 'Catalog', to: '/catalog' }, { label: 'Details' }]}>
        <Alert severity="error">Failed to load item details.</Alert>
      </PageContainer>
    );
  }

  const isPublished = item.status === 'published';
  const isDraft = item.status === 'draft' || item.status === 'unpublished';

  return (
    <PageContainer
      title={item.name}
      breadcrumbs={[{ label: 'Catalog', to: '/catalog' }, { label: item.name }]}
      actions={
        <>
          <RequireRole roles={['administrator', 'operations_manager']}>
            <Box sx={{ display: 'flex', gap: 1 }}>
              {isDraft && (
                <Button
                  variant="contained"
                  size="small"
                  startIcon={<PublishIcon />}
                  onClick={() => setPublishOpen(true)}
                  color="success"
                  disabled={isOffline}
                >
                  Publish
                </Button>
              )}
              {isPublished && (
                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<UnpublishedIcon />}
                  onClick={() => setUnpublishOpen(true)}
                  disabled={isOffline}
                >
                  Unpublish
                </Button>
              )}
              <Button
                variant="outlined"
                size="small"
                startIcon={<EditIcon />}
                onClick={() => navigate(`/catalog/${id}/edit`)}
                disabled={isOffline}
              >
                Edit
              </Button>
            </Box>
          </RequireRole>
          <RequireRole roles={['member']}>
            {isPublished && (
              <Button
                variant="contained"
                size="small"
                onClick={() => navigate(`/group-buys?item_id=${id}`)}
                disabled={isOffline}
              >
                Start Group Buy
              </Button>
            )}
          </RequireRole>
        </>
      }
    >
      <OfflineDataNotice hasData={Boolean(item)} dataUpdatedAt={dataUpdatedAt} />

      {error && (
        <Alert severity="warning" sx={{ mb: 3 }}>
          Catalog sync is temporarily unavailable. Showing the latest cached item details when possible.
        </Alert>
      )}

      <Paper variant="outlined" sx={{ p: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
          <Typography variant="h6">{item.name}</Typography>
          <StatusChip status={item.status} />
        </Box>

        {item.description && (
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            {item.description}
          </Typography>
        )}

        <Divider sx={{ my: 2 }} />

        <Grid container spacing={3}>
          <Grid item xs={12} sm={6} md={3}>
            <DetailRow label="Category" value={item.category} />
            <DetailRow label="Brand" value={item.brand} />
            <DetailRow label="SKU" value={item.sku} />
            <DetailRow label="Condition" value={CONDITION_LABELS[item.condition] ?? item.condition} />
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <DetailRow label="Billing Model" value={BILLING_MODEL_LABELS[item.billing_model] ?? item.billing_model} />
            <DetailRow label="Price" value={`$${item.unit_price.toFixed(2)}`} />
            <DetailRow label="Refundable Deposit" value={`$${item.refundable_deposit.toFixed(2)}`} />
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <DetailRow label="Quantity" value={item.quantity} />
            <DetailRow label="Location ID" value={item.location_id} />
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <DetailRow label="Version" value={item.version} />
          </Grid>
        </Grid>

        {(item.availability_windows?.length ?? 0) > 0 && (
          <>
            <Divider sx={{ my: 2 }} />
            <Typography variant="subtitle2" fontWeight={600} gutterBottom>
              Availability Windows
            </Typography>
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
              {item.availability_windows?.map((window) => (
                <Typography key={window.id} variant="body2">
                  {formatWindowLabel(window.start_time, window.end_time)}
                </Typography>
              ))}
            </Box>
          </>
        )}

        {(item.blackout_windows?.length ?? 0) > 0 && (
          <>
            <Divider sx={{ my: 2 }} />
            <Typography variant="subtitle2" fontWeight={600} gutterBottom>
              Blackout Windows
            </Typography>
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
              {item.blackout_windows?.map((window) => (
                <Typography key={window.id} variant="body2">
                  {formatWindowLabel(window.start_time, window.end_time)}
                </Typography>
              ))}
            </Box>
          </>
        )}
      </Paper>

      <ConfirmDialog
        open={publishOpen}
        title="Publish Item"
        message={`Publish "${item.name}"? It will become visible for ordering and group buys.`}
        confirmLabel="Publish"
        loading={publishMutation.isPending}
        onConfirm={handlePublish}
        onCancel={() => setPublishOpen(false)}
      />

      <ConfirmDialog
        open={unpublishOpen}
        title="Unpublish Item"
        message={`Unpublish "${item.name}"? It will no longer be available for new orders.`}
        confirmLabel="Unpublish"
        loading={unpublishMutation.isPending}
        onConfirm={handleUnpublish}
        onCancel={() => setUnpublishOpen(false)}
      />
    </PageContainer>
  );
}
