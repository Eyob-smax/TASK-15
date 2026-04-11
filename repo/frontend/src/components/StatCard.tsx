import type { ReactNode } from 'react';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import TrendingDownIcon from '@mui/icons-material/TrendingDown';
import TrendingFlatIcon from '@mui/icons-material/TrendingFlat';
import Skeleton from '@mui/material/Skeleton';

interface StatCardProps {
  label: string;
  value: string | number;
  change?: {
    percent: number;
    direction: 'up' | 'down' | 'flat';
  };
  period?: string;
  icon?: ReactNode;
  loading?: boolean;
}

const directionColor = {
  up: 'success.main',
  down: 'error.main',
  flat: 'text.secondary',
} as const;

const DirectionIcon = ({ direction }: { direction: 'up' | 'down' | 'flat' }) => {
  if (direction === 'up') return <TrendingUpIcon fontSize="small" />;
  if (direction === 'down') return <TrendingDownIcon fontSize="small" />;
  return <TrendingFlatIcon fontSize="small" />;
};

export function StatCard({ label, value, change, period, icon, loading }: StatCardProps) {
  return (
    <Card>
      <CardContent>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <Typography variant="body2" color="text.secondary" fontWeight={500}>
            {label}
          </Typography>
          {icon && (
            <Box sx={{ color: 'primary.main', opacity: 0.7 }}>
              {icon}
            </Box>
          )}
        </Box>

        {loading ? (
          <>
            <Skeleton variant="text" width="60%" height={48} />
            <Skeleton variant="text" width="40%" height={20} />
          </>
        ) : (
          <>
            <Typography variant="h4" component="div" fontWeight={700} sx={{ mt: 1, mb: 0.5 }}>
              {value}
            </Typography>
            {(change || period) && (
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                {change && (
                  <Box
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      color: directionColor[change.direction],
                      gap: 0.25,
                    }}
                  >
                    <DirectionIcon direction={change.direction} />
                    <Typography variant="caption" fontWeight={600}>
                      {Math.abs(change.percent).toFixed(1)}%
                    </Typography>
                  </Box>
                )}
                {period && (
                  <Typography variant="caption" color="text.secondary">
                    {period}
                  </Typography>
                )}
              </Box>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
}
