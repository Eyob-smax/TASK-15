import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import InboxIcon from '@mui/icons-material/Inbox';
import type { ReactNode } from 'react';

interface EmptyStateProps {
  title?: string;
  description?: string;
  icon?: ReactNode;
  action?: {
    label: string;
    onClick: () => void;
  };
}

export function EmptyState({
  title = 'No data found',
  description,
  icon,
  action,
}: EmptyStateProps) {
  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        py: 8,
        px: 4,
        gap: 2,
        color: 'text.secondary',
      }}
    >
      <Box sx={{ fontSize: 64, color: 'text.disabled' }}>
        {icon ?? <InboxIcon sx={{ fontSize: 'inherit' }} />}
      </Box>
      <Typography variant="h6" color="text.secondary" align="center">
        {title}
      </Typography>
      {description && (
        <Typography variant="body2" color="text.disabled" align="center" maxWidth={400}>
          {description}
        </Typography>
      )}
      {action && (
        <Button variant="contained" onClick={action.onClick} sx={{ mt: 1 }}>
          {action.label}
        </Button>
      )}
    </Box>
  );
}
