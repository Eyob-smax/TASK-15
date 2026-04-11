import Alert from '@mui/material/Alert';
import { useOfflineStatus } from '@/lib/offline';

function formatTimestamp(timestamp: number): string {
  return new Date(timestamp).toLocaleString();
}

export function OfflineStatusBanner() {
  const { isOffline, lastSyncAt } = useOfflineStatus();

  if (!isOffline) {
    return null;
  }

  return (
    <Alert severity="warning" square>
      Offline mode is active. Cached reads remain available and changes are disabled until you reconnect.
      {lastSyncAt ? ` Last successful sync: ${formatTimestamp(lastSyncAt)}.` : ''}
    </Alert>
  );
}
