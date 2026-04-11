import { useState, useCallback } from 'react';
import { useQuery } from '@tanstack/react-query';
import { PageContainer } from '@/components/PageContainer';
import { DataTable, type Column } from '@/components/DataTable';
import { apiClient, type PaginatedResponse } from '@/lib/api-client';
import type { Location } from '@/lib/types';

const COLUMNS: Column<Location>[] = [
  { key: 'name', label: 'Name', width: '25%' },
  { key: 'address', label: 'Address' },
  { key: 'timezone', label: 'Timezone' },
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

export default function LocationsPage() {
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);

  const { data, isLoading, error } = useQuery({
    queryKey: ['locations', page + 1, pageSize],
    queryFn: () =>
      apiClient.get<PaginatedResponse<Location>>('/locations', {
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
    <PageContainer title="Locations" breadcrumbs={[{ label: 'Locations' }]}>
      <DataTable
        columns={COLUMNS}
        rows={rows}
        rowKey={row => row.id}
        loading={isLoading}
        error={error ? 'Failed to load locations.' : null}
        page={page}
        pageSize={pageSize}
        totalCount={totalCount}
        onPageChange={handlePageChange}
        onPageSizeChange={handlePageSizeChange}
        emptyTitle="No locations found"
        emptyDescription="Locations will appear here once created."
      />
    </PageContainer>
  );
}
