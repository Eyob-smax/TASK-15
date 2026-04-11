import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import { PageContainer } from '@/components/PageContainer';
import { DataTable, type Column } from '@/components/DataTable';
import { apiClient } from '@/lib/api-client';
import type { LandedCostEntry } from '@/lib/types';

interface SingleResponse<T> { data: T }

const COLUMNS: Column<LandedCostEntry>[] = [
  { key: 'item_id', label: 'Item ID', render: row => row.item_id.slice(0, 8) + '…' },
  { key: 'purchase_order_id', label: 'PO ID', render: row => row.purchase_order_id.slice(0, 8) + '…' },
  { key: 'period', label: 'Period' },
  { key: 'cost_component', label: 'Component' },
  { key: 'raw_amount', label: 'Raw Amount', align: 'right', render: row => `$${row.raw_amount.toFixed(2)}` },
  { key: 'allocated_amount', label: 'Allocated', align: 'right', render: row => `$${row.allocated_amount.toFixed(2)}` },
  { key: 'allocation_method', label: 'Method' },
];

export default function LandedCostsPage() {
  const [itemId, setItemId] = useState('');
  const [period, setPeriod] = useState('');
  const [queryKey, setQueryKey] = useState<{ itemId: string; period: string } | null>(null);

  const { data, isLoading, error } = useQuery({
    queryKey: ['landed-costs', queryKey],
    queryFn: async () => {
      if (!queryKey?.itemId) return { data: [] };
      const response = await apiClient.get<SingleResponse<LandedCostEntry[]>>(
        '/procurement/landed-costs',
        { item_id: queryKey.itemId, period: queryKey.period || undefined },
      );
      return response;
    },
    enabled: Boolean(queryKey?.itemId),
  });

  const rows = data?.data ?? [];

  const handleSearch = () => {
    if (itemId.trim()) {
      setQueryKey({ itemId: itemId.trim(), period: period.trim() });
    }
  };

  return (
    <PageContainer
      title="Landed Costs"
      breadcrumbs={[{ label: 'Procurement', to: '/procurement' }, { label: 'Landed Costs' }]}
    >
      <Box sx={{ display: 'flex', gap: 2, mb: 3, alignItems: 'flex-end' }}>
        <TextField
          label="Item ID (UUID)"
          value={itemId}
          onChange={e => setItemId(e.target.value)}
          size="small"
          sx={{ minWidth: 300 }}
          placeholder="e.g. 550e8400-e29b-41d4-a716-446655440000"
        />
        <TextField
          label="Period"
          value={period}
          onChange={e => setPeriod(e.target.value)}
          size="small"
          sx={{ minWidth: 140 }}
          placeholder="e.g. 2026-Q1"
        />
        <Button variant="contained" size="small" onClick={handleSearch} disabled={!itemId.trim()}>
          Search
        </Button>
      </Box>

      <DataTable
        columns={COLUMNS}
        rows={rows}
        rowKey={row => row.id}
        loading={isLoading}
        error={error ? 'Failed to load landed costs.' : null}
        page={0}
        pageSize={rows.length || 20}
        totalCount={rows.length}
        onPageChange={() => {}}
        onPageSizeChange={() => {}}
        emptyTitle={queryKey ? 'No landed costs found' : 'Enter an Item ID to search'}
        emptyDescription={queryKey ? 'No cost entries for this item/period combination.' : 'Landed cost entries will appear after a search.'}
      />
    </PageContainer>
  );
}
