import Chip, { type ChipProps } from '@mui/material/Chip';
import type { ItemStatus, OrderStatus, CampaignStatus, POStatus } from '@/lib/types';

type Status = ItemStatus | OrderStatus | CampaignStatus | POStatus | string;

const COLOR_MAP: Record<string, ChipProps['color']> = {
  // Item
  draft: 'default',
  published: 'success',
  unpublished: 'warning',
  // Order
  created: 'info',
  paid: 'success',
  cancelled: 'error',
  refunded: 'default',
  auto_closed: 'default',
  // Campaign
  active: 'info',
  succeeded: 'success',
  failed: 'error',
  // PO
  approved: 'success',
  received: 'success',
  returned: 'warning',
  voided: 'default',
};

const LABEL_MAP: Record<string, string> = {
  draft: 'Draft',
  published: 'Published',
  unpublished: 'Unpublished',
  created: 'Created',
  paid: 'Paid',
  cancelled: 'Cancelled',
  refunded: 'Refunded',
  auto_closed: 'Auto-Closed',
  active: 'Active',
  succeeded: 'Succeeded',
  failed: 'Failed',
  approved: 'Approved',
  received: 'Received',
  returned: 'Returned',
  voided: 'Voided',
};

interface StatusChipProps {
  status: Status;
  size?: ChipProps['size'];
}

export function StatusChip({ status, size = 'small' }: StatusChipProps) {
  return (
    <Chip
      label={LABEL_MAP[status] ?? status}
      color={COLOR_MAP[status] ?? 'default'}
      size={size}
      variant="outlined"
    />
  );
}
