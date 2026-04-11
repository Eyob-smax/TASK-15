import { useState } from 'react';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CircularProgress from '@mui/material/CircularProgress';
import Divider from '@mui/material/Divider';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import Paper from '@mui/material/Paper';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import { ConfirmDialog } from '@/components/ConfirmDialog';
import { PageContainer } from '@/components/PageContainer';
import {
  useBiometric,
  useEncryptionKeys,
  useRegisterBiometric,
  useRevokeBiometric,
  useRotateKey,
} from '@/lib/hooks/useAdmin';
import { useNotify } from '@/lib/notifications';

function isModuleDisabled(error: unknown): boolean {
  if (!error) return false;
  const msg = String(error);
  return msg.includes('501') || msg.toUpperCase().includes('MODULE_DISABLED');
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <Box sx={{ py: 0.5 }}>
      <Typography variant="caption" color="text.secondary" display="block">
        {label}
      </Typography>
      <Typography variant="body2">{value || '-'}</Typography>
    </Box>
  );
}

function BiometricLookupSection() {
  const notify = useNotify();
  const [lookupId, setLookupId] = useState('');
  const [activeUserId, setActiveUserId] = useState<string | undefined>(undefined);
  const [revokeOpen, setRevokeOpen] = useState(false);

  const { data, isLoading, error } = useBiometric(activeUserId);
  const revokeMutation = useRevokeBiometric();

  const handleLookup = () => {
    const trimmed = lookupId.trim();
    if (trimmed) {
      setActiveUserId(trimmed);
    }
  };

  const handleRevoke = async () => {
    if (!activeUserId) return;
    try {
      await revokeMutation.mutateAsync(activeUserId);
      notify.success('Biometric enrollment revoked.');
      setRevokeOpen(false);
      setActiveUserId(undefined);
    } catch {
      notify.error('Failed to revoke biometric enrollment.');
    }
  };

  const moduleDisabled = isModuleDisabled(error);

  return (
    <Box>
      <Typography variant="subtitle1" fontWeight={600} gutterBottom>
        User Lookup
      </Typography>
      <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
        <TextField
          label="User ID"
          value={lookupId}
          onChange={e => setLookupId(e.target.value)}
          size="small"
          sx={{ minWidth: 280 }}
          onKeyDown={e => {
            if (e.key === 'Enter') handleLookup();
          }}
        />
        <Button variant="contained" onClick={handleLookup} disabled={!lookupId.trim()}>
          Lookup
        </Button>
      </Box>

      {moduleDisabled && (
        <Alert severity="info" sx={{ mb: 2 }}>
          Biometric module is not enabled on this server.
        </Alert>
      )}

      {activeUserId && isLoading && (
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <CircularProgress size={20} />
          <Typography variant="body2">Loading enrollment...</Typography>
        </Box>
      )}

      {activeUserId && error && !moduleDisabled && <Alert severity="error">{String(error)}</Alert>}

      {activeUserId && data && !error && (
        <Card variant="outlined" sx={{ maxWidth: 400 }}>
          <CardContent>
            <Typography variant="subtitle2" fontWeight={600} gutterBottom>
              Enrollment Record
            </Typography>
            <DetailRow label="User ID" value={activeUserId} />
            <DetailRow label="Template Ref" value={data.template_ref} />
            <DetailRow
              label="Created At"
              value={data.created_at ? new Date(data.created_at).toLocaleString() : '-'}
            />
            <DetailRow label="Status" value={data.is_active ? 'Active' : 'Revoked'} />
            <Box sx={{ mt: 2 }}>
              <Button
                variant="outlined"
                color="error"
                size="small"
                onClick={() => setRevokeOpen(true)}
                disabled={!data.is_active}
              >
                Revoke
              </Button>
            </Box>
          </CardContent>
        </Card>
      )}

      <ConfirmDialog
        open={revokeOpen}
        title="Revoke Biometric"
        message="Revoke this user's biometric enrollment? They will need to re-register."
        confirmLabel="Revoke"
        destructive
        loading={revokeMutation.isPending}
        onConfirm={handleRevoke}
        onCancel={() => setRevokeOpen(false)}
      />
    </Box>
  );
}

function RegisterSection() {
  const notify = useNotify();
  const registerMutation = useRegisterBiometric();
  const [userId, setUserId] = useState('');
  const [templateRef, setTemplateRef] = useState('');

  const handleRegister = async () => {
    const trimmedUser = userId.trim();
    const trimmedTemplate = templateRef.trim();
    if (!trimmedUser || !trimmedTemplate) return;

    try {
      await registerMutation.mutateAsync({ user_id: trimmedUser, template_ref: trimmedTemplate });
      notify.success('Biometric registered successfully.');
      setUserId('');
      setTemplateRef('');
    } catch (err) {
      if (isModuleDisabled(err)) {
        notify.error('Biometric module is not enabled.');
      } else {
        notify.error('Failed to register biometric.');
      }
    }
  };

  return (
    <Box>
      <Typography variant="subtitle1" fontWeight={600} gutterBottom>
        Register Biometric
      </Typography>
      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5, maxWidth: 400 }}>
        <TextField
          label="User ID"
          value={userId}
          onChange={e => setUserId(e.target.value)}
          size="small"
          fullWidth
        />
        <TextField
          label="Template Ref"
          value={templateRef}
          onChange={e => setTemplateRef(e.target.value)}
          size="small"
          fullWidth
        />
        <Box>
          <Button
            variant="contained"
            onClick={handleRegister}
            disabled={registerMutation.isPending || !userId.trim() || !templateRef.trim()}
            startIcon={
              registerMutation.isPending ? <CircularProgress size={16} color="inherit" /> : undefined
            }
          >
            Register
          </Button>
        </Box>
      </Box>
    </Box>
  );
}

function EncryptionKeysSection() {
  const notify = useNotify();
  const { data, isLoading, error } = useEncryptionKeys();
  const rotateMutation = useRotateKey();

  const handleRotate = async () => {
    try {
      await rotateMutation.mutateAsync(undefined);
      notify.success('Encryption key rotated successfully.');
    } catch (err) {
      if (isModuleDisabled(err)) {
        notify.error('Biometric module is not enabled.');
      } else {
        notify.error('Failed to rotate encryption key.');
      }
    }
  };

  const moduleDisabled = isModuleDisabled(error);
  const keys = data ?? [];

  return (
    <Box>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 1.5 }}>
        <Typography variant="subtitle1" fontWeight={600}>
          Encryption Keys
        </Typography>
        <Button
          variant="outlined"
          size="small"
          onClick={handleRotate}
          disabled={rotateMutation.isPending}
          startIcon={rotateMutation.isPending ? <CircularProgress size={16} color="inherit" /> : undefined}
        >
          Rotate Key
        </Button>
      </Box>

      {moduleDisabled && <Alert severity="info">Biometric module is not enabled on this server.</Alert>}

      {error && !moduleDisabled && <Alert severity="error">{String(error)}</Alert>}

      {isLoading && (
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <CircularProgress size={20} />
          <Typography variant="body2">Loading keys...</Typography>
        </Box>
      )}

      {!isLoading && !error && keys.length > 0 && (
        <Paper variant="outlined" sx={{ maxWidth: 500 }}>
          <List dense disablePadding>
            {keys.map((key, i) => (
              <ListItem key={key.id} divider={i < keys.length - 1}>
                <ListItemText
                  primary={
                    <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                      {key.id}
                    </Typography>
                  }
                  secondary={`${key.purpose} • ${new Date(key.created_at).toLocaleString()}`}
                />
              </ListItem>
            ))}
          </List>
        </Paper>
      )}

      {!isLoading && !error && keys.length === 0 && (
        <Typography variant="body2" color="text.secondary">
          No encryption keys found.
        </Typography>
      )}
    </Box>
  );
}

export default function BiometricPage() {
  return (
    <PageContainer
      title="Biometric Management"
      breadcrumbs={[{ label: 'Admin' }, { label: 'Biometrics' }]}
    >
      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        <BiometricLookupSection />
        <Divider />
        <RegisterSection />
        <Divider />
        <EncryptionKeysSection />
      </Box>
    </PageContainer>
  );
}
