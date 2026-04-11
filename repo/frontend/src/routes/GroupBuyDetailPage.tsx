import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import Divider from '@mui/material/Divider';
import Grid from '@mui/material/Grid';
import LinearProgress from '@mui/material/LinearProgress';
import Paper from '@mui/material/Paper';
import Skeleton from '@mui/material/Skeleton';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { OfflineDataNotice } from '@/components/OfflineDataNotice';
import { PageContainer } from '@/components/PageContainer';
import { ConfirmDialog } from '@/components/ConfirmDialog';
import { StatusChip } from '@/components/StatusChip';
import { RequireRole } from '@/lib/auth';
import { useItem } from '@/lib/hooks/useItems';
import { OFFLINE_MUTATION_MESSAGE, useOfflineStatus } from '@/lib/offline';
import {
  useCampaign,
  useJoinCampaign,
  useCancelCampaign,
  useEvaluateCampaign,
} from '@/lib/hooks/useCampaigns';
import { useNotify } from '@/lib/notifications';
import { joinCampaignSchema, type JoinCampaignFormData } from '@/lib/validation';

export default function GroupBuyDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const notify = useNotify();
  const { isOffline } = useOfflineStatus();

  const [cancelOpen, setCancelOpen] = useState(false);
  const [evaluateOpen, setEvaluateOpen] = useState(false);

  const { data: campaign, isLoading, error, dataUpdatedAt } = useCampaign(id);
  const { data: item, isLoading: itemLoading } = useItem(campaign?.item_id);
  const joinMutation = useJoinCampaign();
  const cancelMutation = useCancelCampaign();
  const evaluateMutation = useEvaluateCampaign();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<JoinCampaignFormData>({
    resolver: zodResolver(joinCampaignSchema),
    defaultValues: { quantity: 1 },
  });

  const handleJoin = async (data: JoinCampaignFormData) => {
    if (!id) return;
    try {
      await joinMutation.mutateAsync({ id, quantity: data.quantity });
      notify.success('Joined campaign successfully! Check your orders.');
      reset();
    } catch {
      notify.error('Failed to join campaign.');
    }
  };

  const handleCancel = async () => {
    if (!id) return;
    try {
      await cancelMutation.mutateAsync(id);
      notify.success('Campaign cancelled.');
      setCancelOpen(false);
      navigate('/group-buys');
    } catch {
      notify.error('Failed to cancel campaign.');
    }
  };

  const handleEvaluate = async () => {
    if (!id) return;
    try {
      await evaluateMutation.mutateAsync(id);
      notify.success('Campaign evaluated.');
      setEvaluateOpen(false);
    } catch {
      notify.error('Failed to evaluate campaign.');
    }
  };

  if (isLoading) {
    return (
      <PageContainer title="Campaign Details" breadcrumbs={[{ label: 'Group Buys', to: '/group-buys' }, { label: 'Details' }]}>
        <Skeleton variant="rectangular" height={300} />
      </PageContainer>
    );
  }

  if (!campaign) {
    return (
      <PageContainer title="Campaign Details" breadcrumbs={[{ label: 'Group Buys', to: '/group-buys' }, { label: 'Details' }]}>
        <Alert severity="error">Failed to load campaign details.</Alert>
      </PageContainer>
    );
  }

  const pct = campaign.min_quantity > 0
    ? Math.min((campaign.current_committed_qty / campaign.min_quantity) * 100, 100)
    : 0;

  const isActive = campaign.status === 'active';
  const campaignLabel = item?.name ? `${item.name} Group Buy` : `Campaign ${campaign.item_id.slice(0, 8)}`;
  const outcomeMessage = campaign.status === 'succeeded'
    ? 'This campaign met the minimum quantity at cutoff and succeeded.'
    : campaign.status === 'failed'
      ? 'This campaign did not reach the minimum quantity at cutoff and failed.'
      : campaign.status === 'cancelled'
        ? 'This campaign was cancelled before completion.'
        : null;

  return (
    <PageContainer
      title={campaignLabel}
      breadcrumbs={[{ label: 'Group Buys', to: '/group-buys' }, { label: campaignLabel }]}
      actions={
        <RequireRole roles={['administrator', 'operations_manager']}>
          <Box sx={{ display: 'flex', gap: 1 }}>
            {isActive && (
              <>
                <Button
                  variant="outlined"
                  size="small"
                  onClick={() => setEvaluateOpen(true)}
                  disabled={isOffline}
                >
                  Evaluate
                </Button>
                <Button
                  variant="outlined"
                  size="small"
                  color="error"
                  onClick={() => setCancelOpen(true)}
                  disabled={isOffline}
                >
                  Cancel
                </Button>
              </>
            )}
          </Box>
        </RequireRole>
      }
    >
      <OfflineDataNotice hasData={Boolean(campaign)} dataUpdatedAt={dataUpdatedAt} />

      {error && (
        <Alert severity="warning" sx={{ mb: 3 }}>
          Campaign sync is temporarily unavailable. Showing the latest cached details when possible.
        </Alert>
      )}

      {outcomeMessage && (
        <Alert severity={campaign.status === 'succeeded' ? 'success' : 'warning'} sx={{ mb: 3 }}>
          {outcomeMessage}
        </Alert>
      )}

      <Grid container spacing={3}>
        {/* Campaign info */}
        <Grid item xs={12} md={8}>
          <Paper variant="outlined" sx={{ p: 3 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
              <Typography variant="h6">{campaignLabel}</Typography>
              <StatusChip status={campaign.status} />
            </Box>

            <Divider sx={{ my: 2 }} />

            {/* Progress */}
            <Box sx={{ mb: 3 }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
                <Typography variant="body2" fontWeight={600}>
                  Participation Progress
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {campaign.current_committed_qty} / {campaign.min_quantity} min
                </Typography>
              </Box>
              <LinearProgress
                variant="determinate"
                value={pct}
                color={pct >= 100 ? 'success' : 'primary'}
                sx={{ height: 10, borderRadius: 5 }}
              />
              <Typography variant="caption" color="text.secondary" sx={{ mt: 0.5, display: 'block' }}>
                {Math.round(pct)}% of minimum reached
              </Typography>
            </Box>

            <Grid container spacing={2}>
              <Grid item xs={6}>
                <Typography variant="caption" color="text.secondary">Cutoff Time</Typography>
                <Typography variant="body2">{new Date(campaign.cutoff_time).toLocaleString()}</Typography>
              </Grid>
              <Grid item xs={6}>
                <Typography variant="caption" color="text.secondary">Item</Typography>
                <Typography variant="body2">{itemLoading ? 'Loading item...' : item?.name ?? campaign.item_id}</Typography>
              </Grid>
              <Grid item xs={6}>
                <Typography variant="caption" color="text.secondary">Category</Typography>
                <Typography variant="body2">{item?.category ?? 'Not available'}</Typography>
              </Grid>
              <Grid item xs={6}>
                <Typography variant="caption" color="text.secondary">Deposit</Typography>
                <Typography variant="body2">{item ? `$${item.refundable_deposit.toFixed(2)}` : 'Not available'}</Typography>
              </Grid>
              {campaign.evaluated_at && (
                <Grid item xs={6}>
                  <Typography variant="caption" color="text.secondary">Evaluated At</Typography>
                  <Typography variant="body2">{new Date(campaign.evaluated_at).toLocaleString()}</Typography>
                </Grid>
              )}
            </Grid>
          </Paper>
        </Grid>

        {/* Join form — members only */}
        {isActive && (
          <Grid item xs={12} md={4}>
            <RequireRole
              roles={['member']}
              fallback={
                <Paper variant="outlined" sx={{ p: 3 }}>
                  <Typography variant="body2" color="text.secondary">
                    Only members can join group buy campaigns.
                  </Typography>
                </Paper>
              }
            >
              <Paper variant="outlined" sx={{ p: 3 }}>
                {isOffline && (
                  <Alert severity="warning" sx={{ mb: 2 }}>{OFFLINE_MUTATION_MESSAGE}</Alert>
                )}
                <Typography variant="subtitle1" fontWeight={600} gutterBottom>
                  Join This Campaign
                </Typography>
                <Box component="form" onSubmit={handleSubmit(handleJoin)} noValidate>
                  <TextField
                    {...register('quantity', { valueAsNumber: true })}
                    label="Quantity"
                    type="number"
                    fullWidth
                    size="small"
                    inputProps={{ min: 1, step: 1 }}
                    error={Boolean(errors.quantity)}
                    helperText={errors.quantity?.message}
                    sx={{ mb: 2 }}
                  />
                  <Button
                    type="submit"
                    variant="contained"
                    fullWidth
                    disabled={isSubmitting || joinMutation.isPending || isOffline}
                    startIcon={
                      (isSubmitting || joinMutation.isPending)
                        ? <CircularProgress size={16} color="inherit" />
                        : undefined
                    }
                  >
                    Join Campaign
                  </Button>
                </Box>
              </Paper>
            </RequireRole>
          </Grid>
        )}
      </Grid>

      <ConfirmDialog
        open={cancelOpen}
        title="Cancel Campaign"
        message={`Cancel this campaign? This action cannot be undone.`}
        confirmLabel="Cancel Campaign"
        destructive
        loading={cancelMutation.isPending}
        onConfirm={handleCancel}
        onCancel={() => setCancelOpen(false)}
      />

      <ConfirmDialog
        open={evaluateOpen}
        title="Evaluate Campaign"
        message={`Evaluate this campaign now? It will be marked as succeeded or failed based on current participation.`}
        confirmLabel="Evaluate"
        loading={evaluateMutation.isPending}
        onConfirm={handleEvaluate}
        onCancel={() => setEvaluateOpen(false)}
      />
    </PageContainer>
  );
}
