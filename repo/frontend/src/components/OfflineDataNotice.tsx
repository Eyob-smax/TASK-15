import Alert from '@mui/material/Alert';
import { useOfflineStatus } from '@/lib/offline';

function formatTimestamp(timestamp: number): string {
  return new Date(timestamp).toLocaleString();
}

interface OfflineDataNoticeProps {
  hasData: boolean;
  dataUpdatedAt?: number;
}

export function OfflineDataNotice({ hasData, dataUpdatedAt }: OfflineDataNoticeProps) {
  const { isOffline } = useOfflineStatus();

  if (!isOffline) {
    return null;
  }

  if (hasData) {
    return (
      <Alert severity="warning" sx={{ mb: 3 }}>
        Offline mode is active. Showing cached data{dataUpdatedAt ? ` from ${formatTimestamp(dataUpdatedAt)}` : ''}.
      </Alert>
    );
  }

  return (
    <Alert severity="info" sx={{ mb: 3 }}>
      Offline mode is active and this view has no cached data yet. Reconnect to load it for the first time.
    </Alert>
  );
}
