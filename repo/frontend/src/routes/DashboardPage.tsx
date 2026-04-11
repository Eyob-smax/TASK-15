import { useState } from 'react';
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import Alert from '@mui/material/Alert';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';
import PeopleIcon from '@mui/icons-material/People';
import TrendingDownIcon from '@mui/icons-material/TrendingDown';
import AutorenewIcon from '@mui/icons-material/Autorenew';
import FitnessCenterIcon from '@mui/icons-material/FitnessCenter';
import SchoolIcon from '@mui/icons-material/School';
import EmojiPeopleIcon from '@mui/icons-material/EmojiPeople';

import { PageContainer } from '@/components/PageContainer';
import { OfflineDataNotice } from '@/components/OfflineDataNotice';
import { StatCard } from '@/components/StatCard';
import { FilterBar, type FilterField } from '@/components/FilterBar';
import { RequireRole } from '@/lib/auth';
import { useDashboardKPIs, type KPIPeriod } from '@/lib/hooks/useDashboard';
import { isOfflineApiError } from '@/lib/api-client';
import { formatPercentage } from '@/lib/format';

import type { KPIMetric } from '@/lib/types';

const PERIOD_OPTIONS: { value: KPIPeriod; label: string }[] = [
  { value: 'daily', label: 'Daily' },
  { value: 'weekly', label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
  { value: 'quarterly', label: 'Quarterly' },
  { value: 'yearly', label: 'Yearly' },
];

const FILTER_FIELDS: FilterField[] = [
  { key: 'location_id', label: 'Location', type: 'text', placeholder: 'Filter by location' },
  { key: 'coach_id', label: 'Coach', type: 'text', placeholder: 'Filter by coach' },
  { key: 'category', label: 'Category', type: 'text', placeholder: 'Item category' },
  { key: 'date_range_start', label: 'From', type: 'date' },
  { key: 'date_range_end', label: 'To', type: 'date' },
];

function toChange(metric: KPIMetric | undefined, negate = false): { percent: number; direction: 'up' | 'down' | 'flat' } | undefined {
  if (!metric) return undefined;
  if (metric.change_percent === undefined) return { percent: 0, direction: 'flat' };
  const cp = metric.change_percent;
  const displayVal = negate ? -cp : cp;
  return {
    percent: Math.abs(displayVal),
    direction: displayVal > 0.001 ? 'up' : displayVal < -0.001 ? 'down' : 'flat',
  };
}

export default function DashboardPage() {
  const [period, setPeriod] = useState<KPIPeriod>('weekly');
  const [filterValues, setFilterValues] = useState<Record<string, string>>({});

  const { data: kpis, isLoading, error, dataUpdatedAt } = useDashboardKPIs({
    period,
    location_id: filterValues.location_id || undefined,
    coach_id: filterValues.coach_id || undefined,
    category: filterValues.category || undefined,
    from: filterValues.date_range_start || undefined,
    to: filterValues.date_range_end || undefined,
  });

  return (
    <PageContainer
      title="Dashboard"
      actions={
        <ToggleButtonGroup
          value={period}
          exclusive
          onChange={(_, val) => val && setPeriod(val)}
          size="small"
          aria-label="period"
        >
          {PERIOD_OPTIONS.map(opt => (
            <ToggleButton key={opt.value} value={opt.value}>
              {opt.label}
            </ToggleButton>
          ))}
        </ToggleButtonGroup>
      }
    >
      {/* Filters — permission-gated for roles with reporting access */}
      <RequireRole roles={['administrator', 'operations_manager', 'coach']}>
        <Box sx={{ mb: 3 }}>
          <FilterBar fields={FILTER_FIELDS} onChange={setFilterValues} />
        </Box>
      </RequireRole>

      <OfflineDataNotice hasData={Boolean(kpis)} dataUpdatedAt={dataUpdatedAt} />

      {/* Error state */}
      {error && !isLoading && (
        <Alert severity="warning" sx={{ mb: 3 }}>
          {isOfflineApiError(error)
            ? 'Dashboard sync is currently offline. Cached KPIs remain available when present.'
            : 'Dashboard data is not yet available. KPIs will appear here once reporting is configured.'}
        </Alert>
      )}

      {/* Membership & engagement KPIs */}
      <Typography variant="overline" color="text.secondary" gutterBottom display="block">
        Membership &amp; Engagement
      </Typography>
      <Grid container spacing={2} sx={{ mb: 3 }}>
        <Grid item xs={12} sm={6} md={4} lg={3}>
          <StatCard
            label="Member Growth"
              value={kpis ? kpis.member_growth.value.toString() : '—'}
              change={kpis ? toChange(kpis.member_growth) : undefined}
              period={period}
              icon={<PeopleIcon />}
              loading={isLoading}
            />
          </Grid>
          <Grid item xs={12} sm={6} md={4} lg={3}>
            <StatCard
              label="Churn Rate"
              value={kpis ? kpis.churn.value.toString() : '—'}
              change={kpis ? toChange(kpis.churn, true) : undefined}
              period={period}
              icon={<TrendingDownIcon />}
              loading={isLoading}
            />
          </Grid>
          <Grid item xs={12} sm={6} md={4} lg={3}>
            <StatCard
              label="Renewal Rate"
              value={kpis ? formatPercentage(kpis.renewal_rate.value) : '—'}
              change={kpis ? toChange(kpis.renewal_rate) : undefined}
              period={period}
              icon={<AutorenewIcon />}
              loading={isLoading}
            />
          </Grid>
          <Grid item xs={12} sm={6} md={4} lg={3}>
            <StatCard
              label="Engagement"
              value={kpis ? Number(kpis.engagement.value).toFixed(2) : '—'}
              change={kpis ? toChange(kpis.engagement) : undefined}
              period={period}
              icon={<FitnessCenterIcon />}
              loading={isLoading}
            />
        </Grid>
      </Grid>

      {/* Operations KPIs — only for roles with reporting access */}
      <RequireRole roles={['administrator', 'operations_manager', 'coach']}>
        <>
          <Divider sx={{ my: 2 }} />
          <Typography variant="overline" color="text.secondary" gutterBottom display="block">
            Operations
          </Typography>
          <Grid container spacing={2} sx={{ mb: 3 }}>
            <Grid item xs={12} sm={6} md={4} lg={3}>
              <StatCard
                label="Class Fill Rate"
                  value={kpis ? formatPercentage(kpis.class_fill_rate.value) : '—'}
                change={kpis ? toChange(kpis.class_fill_rate) : undefined}
                period={period}
                icon={<SchoolIcon />}
                loading={isLoading}
              />
            </Grid>
            <Grid item xs={12} sm={6} md={4} lg={3}>
              <StatCard
                label="Coach Productivity"
                  value={kpis ? Number(kpis.coach_productivity.value).toFixed(2) : '—'}
                change={kpis ? toChange(kpis.coach_productivity) : undefined}
                period={period}
                icon={<EmojiPeopleIcon />}
                loading={isLoading}
              />
            </Grid>
          </Grid>
        </>
      </RequireRole>
    </PageContainer>
  );
}
