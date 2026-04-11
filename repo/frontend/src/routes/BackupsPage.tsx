import { useState } from 'react';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import Typography from '@mui/material/Typography';
import { PageContainer } from '@/components/PageContainer';
import { DataTable, type Column } from '@/components/DataTable';
import { StatusChip } from '@/components/StatusChip';
import { useBackupList, useTriggerBackup } from '@/lib/hooks/useAdmin';
import { useNotify } from '@/lib/notifications';
import type { BackupRun } from '@/lib/types';

function formatBytes(bytes: number): string {
  if (!bytes || bytes === 0) return '—';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let size = bytes;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }
  return `${size.toFixed(1)} ${units[unitIndex]}`;
}

function shortStr(s: string | null | undefined, len = 8): string {
  if (!s) return '—';
  return s.slice(0, len) + (s.length > len ? '…' : '');
}

export default function BackupsPage() {
  const notify = useNotify();
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);

  const { data, isLoading, error } = useBackupList({ page: page + 1, page_size: pageSize });
  const triggerMutation = useTriggerBackup();

  const handleTrigger = async () => {
    try {
      await triggerMutation.mutateAsync(undefined);
      notify.success('Backup triggered successfully.');
    } catch {
      notify.error('Failed to trigger backup.');
    }
  };

  const columns: Column<BackupRun>[] = [
    {
      key: 'id',
      label: 'ID',
      render: row => (
        <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
          {shortStr(row.id)}
        </Typography>
      ),
    },
    {
      key: 'status',
      label: 'Status',
      render: row => <StatusChip status={row.status} />,
    },
    {
      key: 'file_size',
      label: 'File Size',
      align: 'right',
      render: row => formatBytes(row.file_size),
    },
    {
      key: 'checksum',
      label: 'Checksum',
      render: row => (
        <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
          {shortStr(row.checksum, 12)}
        </Typography>
      ),
    },
    {
      key: 'started_at',
      label: 'Started At',
      render: row => row.started_at ? new Date(row.started_at).toLocaleString() : '—',
    },
    {
      key: 'completed_at',
      label: 'Completed At',
      render: row => row.completed_at ? new Date(row.completed_at).toLocaleString() : '—',
    },
  ];

  const rows = data?.data ?? [];
  const totalCount = data?.pagination?.total_count ?? 0;

  return (
    <PageContainer
      title="Backups"
      breadcrumbs={[{ label: 'Admin' }, { label: 'Backups' }]}
      actions={
        <Button
          variant="contained"
          size="small"
          onClick={handleTrigger}
          disabled={triggerMutation.isPending}
          startIcon={triggerMutation.isPending ? <CircularProgress size={16} color="inherit" /> : undefined}
        >
          Trigger Backup
        </Button>
      }
    >
      <DataTable
        columns={columns}
        rows={rows}
        rowKey={row => row.id}
        loading={isLoading}
        error={error ? 'Failed to load backups.' : null}
        page={page}
        pageSize={pageSize}
        totalCount={totalCount}
        onPageChange={setPage}
        onPageSizeChange={ps => { setPageSize(ps); setPage(0); }}
        emptyTitle="No backups found"
        emptyDescription="Backup runs will appear here once triggered."
      />
    </PageContainer>
  );
}
