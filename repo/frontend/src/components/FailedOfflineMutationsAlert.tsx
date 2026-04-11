import { useEffect, useState } from 'react';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import {
  loadPendingOfflineMutations,
  removeOfflineMutation,
  type OfflineMutationEntry,
} from '@/lib/offline-cache';
import { useOfflineStatus } from '@/lib/offline';

export function FailedOfflineMutationsAlert() {
  const { isOnline } = useOfflineStatus();
  const [failedMutations, setFailedMutations] = useState<OfflineMutationEntry[]>([]);

  useEffect(() => {
    loadPendingOfflineMutations().then((mutations) => {
      setFailedMutations(mutations.filter((m) => m.status === 'failed'));
    });
  }, [isOnline]);

  const dismiss = async (id: string) => {
    await removeOfflineMutation(id);
    setFailedMutations((prev) => prev.filter((m) => m.id !== id));
  };

  if (failedMutations.length === 0) {
    return null;
  }

  return (
    <Alert severity="error" square>
      <Typography variant="body2" fontWeight={600} gutterBottom>
        {failedMutations.length} queued action{failedMutations.length > 1 ? 's' : ''} failed to sync.
      </Typography>
      {failedMutations.map((entry) => (
        <Box
          key={entry.id}
          sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1, mt: 0.5 }}
        >
          <Typography variant="caption">
            <strong>{entry.type}</strong>
            {entry.lastError ? `: ${entry.lastError}` : ''}
          </Typography>
          <Button
            size="small"
            color="error"
            onClick={() => dismiss(entry.id)}
            sx={{ minWidth: 0, py: 0, px: 1 }}
          >
            Dismiss
          </Button>
        </Box>
      ))}
    </Alert>
  );
}
