import { useState } from 'react';
import Box from '@mui/material/Box';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import { PageContainer } from '@/components/PageContainer';
import { DataTable, type Column } from '@/components/DataTable';
import { useAuditLog, useSecurityEvents } from '@/lib/hooks/useAdmin';

// The API returns a different shape than the AuditEvent type in lib/types.ts
interface ApiAuditEvent {
  id: string;
  event_type: string;
  entity_type: string;
  entity_id: string;
  actor_id: string;
  details: Record<string, unknown> | string | null;
  integrity_hash: string;
  created_at: string;
}

function shortId(id: string | null | undefined): string {
  if (!id) return '—';
  return id.slice(0, 8) + '…';
}

const AUDIT_COLUMNS: Column<ApiAuditEvent>[] = [
  {
    key: 'event_type',
    label: 'Event Type',
    render: row => (
      <Typography variant="body2" sx={{ fontFamily: 'monospace', fontSize: '0.8rem' }}>
        {row.event_type}
      </Typography>
    ),
  },
  {
    key: 'entity_type',
    label: 'Entity Type',
  },
  {
    key: 'entity_id',
    label: 'Entity ID',
    render: row => (
      <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
        {shortId(row.entity_id)}
      </Typography>
    ),
  },
  {
    key: 'actor_id',
    label: 'Actor ID',
    render: row => (
      <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
        {shortId(row.actor_id)}
      </Typography>
    ),
  },
  {
    key: 'created_at',
    label: 'Created At',
    render: row => new Date(row.created_at).toLocaleString(),
  },
];

function AllEventsTab() {
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [eventTypeFilter, setEventTypeFilter] = useState('');
  const [appliedFilter, setAppliedFilter] = useState('');

  const { data, isLoading, error } = useAuditLog({
    page: page + 1,
    page_size: pageSize,
    event_type: appliedFilter || undefined,
  });

  const handleFilterKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      setAppliedFilter(eventTypeFilter.trim());
      setPage(0);
    }
  };

  const rows = (data?.data ?? []) as ApiAuditEvent[];
  const totalCount = data?.pagination?.total_count ?? 0;

  return (
    <Box>
      <Box sx={{ mb: 2 }}>
        <TextField
          label="Filter by Event Type"
          value={eventTypeFilter}
          onChange={e => setEventTypeFilter(e.target.value)}
          onKeyDown={handleFilterKeyDown}
          size="small"
          placeholder="Press Enter to apply"
          sx={{ minWidth: 260 }}
        />
      </Box>
      <DataTable
        columns={AUDIT_COLUMNS}
        rows={rows}
        rowKey={row => row.id}
        loading={isLoading}
        error={error ? 'Failed to load audit events.' : null}
        page={page}
        pageSize={pageSize}
        totalCount={totalCount}
        onPageChange={setPage}
        onPageSizeChange={ps => { setPageSize(ps); setPage(0); }}
        emptyTitle="No audit events found"
        emptyDescription="Audit events will appear here once actions are performed."
      />
    </Box>
  );
}

function SecurityEventsTab() {
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);

  const { data, isLoading, error } = useSecurityEvents({
    page: page + 1,
    page_size: pageSize,
  });

  const rows = (data?.data ?? []) as ApiAuditEvent[];
  const totalCount = data?.pagination?.total_count ?? 0;

  return (
    <DataTable
      columns={AUDIT_COLUMNS}
      rows={rows}
      rowKey={row => row.id}
      loading={isLoading}
      error={error ? 'Failed to load security events.' : null}
      page={page}
      pageSize={pageSize}
      totalCount={totalCount}
      onPageChange={setPage}
      onPageSizeChange={ps => { setPageSize(ps); setPage(0); }}
      emptyTitle="No security events found"
      emptyDescription="Security-related audit events will appear here."
    />
  );
}

export default function AuditPage() {
  const [tab, setTab] = useState(0);

  return (
    <PageContainer
      title="Audit Log"
      breadcrumbs={[{ label: 'Admin' }, { label: 'Audit Log' }]}
    >
      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 3 }}>
        <Tabs value={tab} onChange={(_, v) => setTab(v)} aria-label="audit log tabs">
          <Tab label="All Events" id="audit-tab-0" aria-controls="audit-panel-0" />
          <Tab label="Security Events" id="audit-tab-1" aria-controls="audit-panel-1" />
        </Tabs>
      </Box>

      <Box role="tabpanel" id="audit-panel-0" hidden={tab !== 0}>
        {tab === 0 && <AllEventsTab />}
      </Box>
      <Box role="tabpanel" id="audit-panel-1" hidden={tab !== 1}>
        {tab === 1 && <SecurityEventsTab />}
      </Box>
    </PageContainer>
  );
}
