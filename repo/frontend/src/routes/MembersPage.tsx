import { useState, useCallback } from 'react';
import { useQuery } from '@tanstack/react-query';
import { PageContainer } from '@/components/PageContainer';
import { DataTable, type Column } from '@/components/DataTable';
import { StatusChip } from '@/components/StatusChip';
import { apiClient, type PaginatedResponse } from '@/lib/api-client';
import type { Member } from '@/lib/types';

function shortId(id: string) {
  return id.slice(0, 8) + '…';
}

const COLUMNS: Column<Member>[] = [
  { key: 'id', label: 'Member ID', render: row => shortId(row.id) },
  { key: 'user_id', label: 'User ID', render: row => shortId(row.user_id) },
  { key: 'location_id', label: 'Location ID', render: row => shortId(row.location_id) },
  {
    key: 'membership_status',
    label: 'Status',
    render: row => <StatusChip status={row.membership_status} />,
  },
  {
    key: 'joined_at',
    label: 'Joined',
    render: row => new Date(row.joined_at).toLocaleDateString(),
  },
  {
    key: 'renewal_date',
    label: 'Renewal',
    render: row => row.renewal_date ? new Date(row.renewal_date).toLocaleDateString() : '—',
  },
];

export default function MembersPage() {
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);

  const { data, isLoading, error } = useQuery({
    queryKey: ['members', page + 1, pageSize],
    queryFn: () =>
      apiClient.get<PaginatedResponse<Member>>('/members', {
        page: page + 1,
        page_size: pageSize,
      }),
  });

  const handlePageChange = useCallback((p: number) => setPage(p), []);
  const handlePageSizeChange = useCallback((ps: number) => {
    setPageSize(ps);
    setPage(0);
  }, []);

  const rows = data?.data ?? [];
  const totalCount = data?.pagination.total_count ?? 0;

  return (
    <PageContainer title="Members" breadcrumbs={[{ label: 'Members' }]}>
      <DataTable
        columns={COLUMNS}
        rows={rows}
        rowKey={row => row.id}
        loading={isLoading}
        error={error ? 'Failed to load members.' : null}
        page={page}
        pageSize={pageSize}
        totalCount={totalCount}
        onPageChange={handlePageChange}
        onPageSizeChange={handlePageSizeChange}
        emptyTitle="No members found"
        emptyDescription="Members will appear here once enrolled."
      />
    </PageContainer>
  );
}
