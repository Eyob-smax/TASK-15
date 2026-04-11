import { useState, useCallback } from 'react';
import { useQuery } from '@tanstack/react-query';
import { PageContainer } from '@/components/PageContainer';
import { DataTable, type Column } from '@/components/DataTable';
import { apiClient, type PaginatedResponse } from '@/lib/api-client';
import type { Coach } from '@/lib/types';

function shortId(id: string) {
  return id.slice(0, 8) + '…';
}

const COLUMNS: Column<Coach>[] = [
  { key: 'id', label: 'Coach ID', render: row => shortId(row.id) },
  { key: 'user_id', label: 'User ID', render: row => shortId(row.user_id) },
  { key: 'location_id', label: 'Location ID', render: row => shortId(row.location_id) },
  { key: 'specialization', label: 'Specialization' },
  {
    key: 'is_active',
    label: 'Active',
    render: row => (row.is_active ? 'Yes' : 'No'),
  },
  {
    key: 'created_at',
    label: 'Created',
    render: row => new Date(row.created_at).toLocaleDateString(),
  },
];

export default function CoachesPage() {
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);

  const { data, isLoading, error } = useQuery({
    queryKey: ['coaches', page + 1, pageSize],
    queryFn: () =>
      apiClient.get<PaginatedResponse<Coach>>('/coaches', {
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
    <PageContainer title="Coaches" breadcrumbs={[{ label: 'Coaches' }]}>
      <DataTable
        columns={COLUMNS}
        rows={rows}
        rowKey={row => row.id}
        loading={isLoading}
        error={error ? 'Failed to load coaches.' : null}
        page={page}
        pageSize={pageSize}
        totalCount={totalCount}
        onPageChange={handlePageChange}
        onPageSizeChange={handlePageSizeChange}
        emptyTitle="No coaches found"
        emptyDescription="Coaches will appear here once assigned."
      />
    </PageContainer>
  );
}
