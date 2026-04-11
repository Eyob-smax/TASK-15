import type {
  GroupBuyCampaign,
  GroupBuyParticipant,
  CampaignStatus,
} from '@/lib/types';

export type {
  GroupBuyCampaign,
  GroupBuyParticipant,
  CampaignStatus,
};

export interface CampaignFilters {
  status?: CampaignStatus;
  item_id?: string;
  search?: string;
  created_after?: string;
  created_before?: string;
  page?: number;
  page_size?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface JoinFormData {
  quantity: number;
}

export interface CampaignProgressData {
  campaign: GroupBuyCampaign;
  participants: GroupBuyParticipant[];
  progress_percent: number;
  remaining_quantity: number;
  time_remaining_seconds: number;
}
