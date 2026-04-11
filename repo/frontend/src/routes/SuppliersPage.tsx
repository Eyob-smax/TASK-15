import { useState, useCallback } from 'react';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Alert from '@mui/material/Alert';
import AddIcon from '@mui/icons-material/Add';
import { OfflineDataNotice } from '@/components/OfflineDataNotice';
import { PageContainer } from '@/components/PageContainer';
import { DataTable, type Column } from '@/components/DataTable';
import { FilterBar, type FilterField } from '@/components/FilterBar';
import { RequireRole } from '@/lib/auth';
import { OFFLINE_MUTATION_MESSAGE, useOfflineStatus } from '@/lib/offline';
import { useSupplierList, useCreateSupplier } from '@/lib/hooks/useProcurement';
import type { Supplier } from '@/lib/types';

const FILTER_FIELDS: FilterField[] = [
  { key: 'search', label: 'Search', type: 'text', placeholder: 'Name or contact…' },
];

const COLUMNS: Column<Supplier>[] = [
  { key: 'name', label: 'Name', width: '22%' },
  { key: 'contact_name', label: 'Contact Name', width: '18%' },
  { key: 'contact_email', label: 'Contact Email', width: '22%' },
  { key: 'contact_phone', label: 'Phone', width: '14%' },
  {
    key: 'is_active',
    label: 'Active',
    width: '10%',
    render: row => (
      <Chip
        label={row.is_active ? 'Active' : 'Inactive'}
        color={row.is_active ? 'success' : 'default'}
        size="small"
        variant="outlined"
      />
    ),
  },
  {
    key: 'created_at',
    label: 'Created',
    render: row => new Date(row.created_at).toLocaleDateString(),
  },
];

interface CreateSupplierForm {
  name: string;
  contact_name: string;
  contact_email: string;
  contact_phone: string;
  address: string;
}

const EMPTY_FORM: CreateSupplierForm = {
  name: '',
  contact_name: '',
  contact_email: '',
  contact_phone: '',
  address: '',
};

export default function SuppliersPage() {
  const { isOffline } = useOfflineStatus();
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [filters, setFilters] = useState<Record<string, string>>({});
  const [dialogOpen, setDialogOpen] = useState(false);
  const [form, setForm] = useState<CreateSupplierForm>(EMPTY_FORM);

  const { data, isLoading, error, dataUpdatedAt } = useSupplierList({
    page: page + 1,
    page_size: pageSize,
    search: filters.search || undefined,
  });

  const createSupplier = useCreateSupplier();

  const handleFiltersChange = useCallback((values: Record<string, string>) => {
    setFilters(values);
    setPage(0);
  }, []);

  const handlePageChange = useCallback((p: number) => setPage(p), []);
  const handlePageSizeChange = useCallback((ps: number) => {
    setPageSize(ps);
    setPage(0);
  }, []);

  const handleOpenDialog = useCallback(() => {
    setForm(EMPTY_FORM);
    setDialogOpen(true);
  }, []);

  const handleCloseDialog = useCallback(() => {
    setDialogOpen(false);
  }, []);

  const handleFormChange = useCallback(
    (field: keyof CreateSupplierForm) =>
      (e: React.ChangeEvent<HTMLInputElement>) => {
        setForm(prev => ({ ...prev, [field]: e.target.value }));
      },
    [],
  );

  const handleSubmit = useCallback(async () => {
    if (!form.name.trim() || isOffline) return;
    await createSupplier.mutateAsync({
      name: form.name.trim(),
      contact_name: form.contact_name.trim() || undefined,
      contact_email: form.contact_email.trim() || undefined,
      contact_phone: form.contact_phone.trim() || undefined,
      address: form.address.trim() || undefined,
    });
    setDialogOpen(false);
  }, [form, createSupplier, isOffline]);

  const rows = data?.data ?? [];
  const totalCount = data?.pagination.total_count ?? 0;

  return (
    <PageContainer
      title="Suppliers"
      breadcrumbs={[{ label: 'Procurement', to: '/procurement' }, { label: 'Suppliers' }]}
      actions={
        <RequireRole roles={['administrator', 'operations_manager', 'procurement_specialist']}>
          <Button
            variant="contained"
            size="small"
            startIcon={<AddIcon />}
            onClick={handleOpenDialog}
            disabled={isOffline}
          >
            New Supplier
          </Button>
        </RequireRole>
      }
    >
      <Box sx={{ mb: 2 }}>
        <FilterBar fields={FILTER_FIELDS} onChange={handleFiltersChange} />
      </Box>

      <OfflineDataNotice hasData={rows.length > 0} dataUpdatedAt={dataUpdatedAt} />

      <DataTable
        columns={COLUMNS}
        rows={rows}
        rowKey={row => row.id}
        loading={isLoading}
        error={error ? 'Failed to load suppliers. Please try again.' : null}
        page={page}
        pageSize={pageSize}
        totalCount={totalCount}
        onPageChange={handlePageChange}
        onPageSizeChange={handlePageSizeChange}
        emptyTitle="No suppliers found"
        emptyDescription="Add a supplier to get started."
      />

      {/* Create Supplier Dialog */}
      <Dialog open={dialogOpen} onClose={handleCloseDialog} maxWidth="sm" fullWidth>
        <DialogTitle>New Supplier</DialogTitle>
        <DialogContent>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: 1 }}>
            {isOffline && (
              <Alert severity="warning">{OFFLINE_MUTATION_MESSAGE}</Alert>
            )}
            <TextField
              label="Name"
              value={form.name}
              onChange={handleFormChange('name')}
              required
              fullWidth
              size="small"
              autoFocus
            />
            <TextField
              label="Contact Name"
              value={form.contact_name}
              onChange={handleFormChange('contact_name')}
              fullWidth
              size="small"
            />
            <TextField
              label="Contact Email"
              value={form.contact_email}
              onChange={handleFormChange('contact_email')}
              fullWidth
              size="small"
              type="email"
            />
            <TextField
              label="Contact Phone"
              value={form.contact_phone}
              onChange={handleFormChange('contact_phone')}
              fullWidth
              size="small"
            />
            <TextField
              label="Address"
              value={form.address}
              onChange={handleFormChange('address')}
              fullWidth
              size="small"
              multiline
              rows={2}
            />
            {createSupplier.isError && (
              <Typography variant="caption" color="error">
                Failed to create supplier. Please try again.
              </Typography>
            )}
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog} color="inherit">
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            variant="contained"
            disabled={!form.name.trim() || createSupplier.isPending || isOffline}
          >
            {createSupplier.isPending ? 'Creating…' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </PageContainer>
  );
}
