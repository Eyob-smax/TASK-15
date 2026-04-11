import { useState } from 'react';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import TextField from '@mui/material/TextField';
import { PageContainer } from '@/components/PageContainer';
import { DataTable, type Column } from '@/components/DataTable';
import { StatusChip } from '@/components/StatusChip';
import { ConfirmDialog } from '@/components/ConfirmDialog';
import {
  useUserList,
  useCreateUser,
  useDeactivateUser,
} from '@/lib/hooks/useAdmin';
import { useNotify } from '@/lib/notifications';
import type { User, UserRole } from '@/lib/types';

const ROLE_OPTIONS: { value: UserRole; label: string }[] = [
  { value: 'administrator', label: 'Administrator' },
  { value: 'operations_manager', label: 'Operations Manager' },
  { value: 'procurement_specialist', label: 'Procurement Specialist' },
  { value: 'coach', label: 'Coach' },
  { value: 'member', label: 'Member' },
];

interface CreateUserForm {
  email: string;
  password: string;
  display_name: string;
  role: UserRole;
  location_id: string;
}

const EMPTY_FORM: CreateUserForm = {
  email: '',
  password: '',
  display_name: '',
  role: 'member',
  location_id: '',
};

function CreateUserDialog({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const notify = useNotify();
  const createMutation = useCreateUser();
  const [form, setForm] = useState<CreateUserForm>(EMPTY_FORM);

  const handleChange = (field: keyof CreateUserForm, value: string) => {
    setForm(prev => ({ ...prev, [field]: value }));
  };

  const handleSubmit = async () => {
    if (!form.email.trim() || !form.password.trim() || !form.display_name.trim()) return;
    try {
      await createMutation.mutateAsync({
        email: form.email.trim(),
        password: form.password,
        display_name: form.display_name.trim(),
        role: form.role,
        location_id: form.location_id.trim() || undefined,
      });
      notify.success('User created successfully.');
      setForm(EMPTY_FORM);
      onClose();
    } catch {
      notify.error('Failed to create user.');
    }
  };

  const isValid = form.email.trim() && form.password.trim() && form.display_name.trim();

  return (
    <Dialog open={open} onClose={onClose} maxWidth="xs" fullWidth>
      <DialogTitle>Create User</DialogTitle>
      <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: '16px !important' }}>
        <TextField
          label="Email"
          type="email"
          value={form.email}
          onChange={e => handleChange('email', e.target.value)}
          fullWidth
          size="small"
          required
        />
        <TextField
          label="Password"
          type="password"
          value={form.password}
          onChange={e => handleChange('password', e.target.value)}
          fullWidth
          size="small"
          required
        />
        <TextField
          label="Display Name"
          value={form.display_name}
          onChange={e => handleChange('display_name', e.target.value)}
          fullWidth
          size="small"
          required
        />
        <FormControl fullWidth size="small">
          <InputLabel>Role</InputLabel>
          <Select
            value={form.role}
            label="Role"
            onChange={e => handleChange('role', e.target.value as UserRole)}
          >
            {ROLE_OPTIONS.map(opt => (
              <MenuItem key={opt.value} value={opt.value}>{opt.label}</MenuItem>
            ))}
          </Select>
        </FormControl>
        <TextField
          label="Location ID (optional)"
          value={form.location_id}
          onChange={e => handleChange('location_id', e.target.value)}
          fullWidth
          size="small"
        />
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose} disabled={createMutation.isPending}>Cancel</Button>
        <Button
          variant="contained"
          onClick={handleSubmit}
          disabled={createMutation.isPending || !isValid}
          startIcon={createMutation.isPending ? <CircularProgress size={16} color="inherit" /> : undefined}
        >
          Create
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default function UsersPage() {
  const notify = useNotify();
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [createOpen, setCreateOpen] = useState(false);
  const [deactivateTarget, setDeactivateTarget] = useState<User | null>(null);

  const { data, isLoading, error } = useUserList({ page: page + 1, page_size: pageSize });
  const deactivateMutation = useDeactivateUser();

  const handleDeactivate = async () => {
    if (!deactivateTarget) return;
    try {
      await deactivateMutation.mutateAsync(deactivateTarget.id);
      notify.success('User deactivated.');
      setDeactivateTarget(null);
    } catch {
      notify.error('Failed to deactivate user.');
    }
  };

  const columns: Column<User>[] = [
    {
      key: 'display_name',
      label: 'Display Name',
    },
    {
      key: 'email',
      label: 'Email',
    },
    {
      key: 'role',
      label: 'Role',
      render: row => <StatusChip status={row.role} />,
    },
    {
      key: 'status',
      label: 'Status',
      render: row => <StatusChip status={row.status} />,
    },
    {
      key: 'created_at',
      label: 'Created At',
      render: row => new Date(row.created_at).toLocaleDateString(),
    },
    {
      key: 'actions',
      label: '',
      render: row => (
        <Button
          size="small"
          variant="outlined"
          color="error"
          disabled={row.status === 'inactive'}
          onClick={() => setDeactivateTarget(row)}
        >
          Deactivate
        </Button>
      ),
    },
  ];

  const rows = data?.data ?? [];
  const totalCount = data?.pagination.total_count ?? 0;

  return (
    <PageContainer
      title="Users"
      breadcrumbs={[{ label: 'Admin' }, { label: 'Users' }]}
      actions={
        <Button variant="contained" size="small" onClick={() => setCreateOpen(true)}>
          Create User
        </Button>
      }
    >
      <DataTable
        columns={columns}
        rows={rows}
        rowKey={row => row.id}
        loading={isLoading}
        error={error ? 'Failed to load users.' : null}
        page={page}
        pageSize={pageSize}
        totalCount={totalCount}
        onPageChange={setPage}
        onPageSizeChange={ps => { setPageSize(ps); setPage(0); }}
        emptyTitle="No users found"
        emptyDescription="Users will appear here once created."
      />

      <CreateUserDialog open={createOpen} onClose={() => setCreateOpen(false)} />

      <ConfirmDialog
        open={!!deactivateTarget}
        title="Deactivate User"
        message={`Deactivate ${deactivateTarget?.display_name ?? 'this user'}? They will no longer be able to log in.`}
        confirmLabel="Deactivate"
        destructive
        loading={deactivateMutation.isPending}
        onConfirm={handleDeactivate}
        onCancel={() => setDeactivateTarget(null)}
      />
    </PageContainer>
  );
}
