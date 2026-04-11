import { useState } from 'react';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardActions from '@mui/material/CardActions';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Divider from '@mui/material/Divider';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import { OfflineDataNotice } from '@/components/OfflineDataNotice';
import { downloadFile } from '@/lib/api-client';
import { PageContainer } from '@/components/PageContainer';
import { OFFLINE_MUTATION_MESSAGE, useOfflineStatus } from '@/lib/offline';
import { useReportList, useRunExport } from '@/lib/hooks/useReports';
import { useNotify } from '@/lib/notifications';
import type { ReportDefinition, ExportJob } from '@/lib/types';

function formatExportStatus(status: string): 'default' | 'warning' | 'success' | 'error' {
  switch (status) {
    case 'completed': return 'success';
    case 'failed': return 'error';
    case 'processing': return 'warning';
    default: return 'default';
  }
}

function ReportCard({
  report,
  onExport,
}: {
  report: ReportDefinition;
  onExport: (reportId: string, format: 'csv' | 'pdf') => Promise<void>;
}) {
  const [csvPending, setCsvPending] = useState(false);
  const [pdfPending, setPdfPending] = useState(false);

  const handleExport = async (format: 'csv' | 'pdf') => {
    if (format === 'csv') setCsvPending(true);
    else setPdfPending(true);
    try {
      await onExport(report.id, format);
    } finally {
      if (format === 'csv') setCsvPending(false);
      else setPdfPending(false);
    }
  };

  return (
    <Card variant="outlined">
      <CardContent sx={{ pb: 1 }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 1, mb: 0.5 }}>
          <Typography variant="subtitle2" fontWeight={600}>
            {report.name}
          </Typography>
          <Chip
            label={report.report_type}
            size="small"
            variant="outlined"
            sx={{ fontSize: '0.7rem', height: 20 }}
          />
        </Box>
        {report.description && (
          <Typography variant="body2" color="text.secondary">
            {report.description}
          </Typography>
        )}
      </CardContent>
      <CardActions sx={{ px: 2, pb: 2 }}>
        <Button
          size="small"
          variant="outlined"
          onClick={() => handleExport('csv')}
          disabled={csvPending || pdfPending}
          startIcon={csvPending ? <CircularProgress size={14} color="inherit" /> : undefined}
        >
          Export CSV
        </Button>
        <Button
          size="small"
          variant="outlined"
          onClick={() => handleExport('pdf')}
          disabled={csvPending || pdfPending}
          startIcon={pdfPending ? <CircularProgress size={14} color="inherit" /> : undefined}
        >
          Export PDF
        </Button>
      </CardActions>
    </Card>
  );
}

function ExportJobItem({
  job,
  onDownload,
  isOffline,
}: {
  job: ExportJob;
  onDownload: (job: ExportJob) => Promise<void>;
  isOffline: boolean;
}) {
  return (
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        py: 1,
        px: 2,
        borderBottom: 1,
        borderColor: 'divider',
        '&:last-child': { borderBottom: 0 },
      }}
    >
      <Box>
        <Typography variant="body2" sx={{ fontFamily: 'monospace', fontSize: '0.8rem' }}>
          {job.filename}
        </Typography>
        <Typography variant="caption" color="text.secondary">
          {job.format.toUpperCase()} · {new Date(job.created_at).toLocaleString()}
        </Typography>
      </Box>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <Chip
          label={job.status}
          size="small"
          color={formatExportStatus(job.status)}
          variant="outlined"
        />
        {job.status === 'completed' && (
          <Button size="small" variant="text" onClick={() => onDownload(job)} disabled={isOffline}>
            Download
          </Button>
        )}
      </Box>
    </Box>
  );
}

export default function ReportsPage() {
  const notify = useNotify();
  const { isOffline } = useOfflineStatus();
  const { data: reportData, isLoading, error, dataUpdatedAt } = useReportList();
  const runExportMutation = useRunExport();
  const [exportJobs, setExportJobs] = useState<ExportJob[]>([]);
  const [filters, setFilters] = useState({
    location_id: '',
    coach_id: '',
    category: '',
    from: '',
    to: '',
    status: '',
  });

  const exportParameters = Object.fromEntries(
    Object.entries(filters).filter(([, value]) => value !== ''),
  );

  const handleDownload = async (job: ExportJob) => {
    try {
      await downloadFile(`/exports/${job.id}/download`, job.filename);
      notify.success(`Downloaded ${job.filename}.`);
    } catch {
      notify.error('Failed to download export.');
    }
  };

  const handleExport = async (reportId: string, format: 'csv' | 'pdf') => {
    try {
      const job = await runExportMutation.mutateAsync({
        report_id: reportId,
        format,
        parameters: exportParameters,
      });
      if (job) {
        setExportJobs(prev => [job, ...prev.filter(existing => existing.id !== job.id)]);
      }
      if (job?.status === 'completed') {
        await handleDownload(job);
      } else {
        notify.success(`Export job created (${format.toUpperCase()}).`);
      }
    } catch {
      notify.error('Failed to create export job.');
    }
  };

  const reports: ReportDefinition[] = Array.isArray(reportData)
    ? reportData
    : reportData?.data ?? [];

  return (
    <PageContainer title="Reports" breadcrumbs={[{ label: 'Reports' }]}>
      <OfflineDataNotice hasData={reports.length > 0} dataUpdatedAt={dataUpdatedAt} />

      {isOffline && (
        <Alert severity="warning" sx={{ mb: 3 }}>
          {OFFLINE_MUTATION_MESSAGE} Export requests are queued locally; downloads still require reconnect.
        </Alert>
      )}

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {String(error)}
        </Alert>
      )}

      {isLoading && (
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 3 }}>
          <CircularProgress size={20} />
          <Typography variant="body2">Loading reports...</Typography>
        </Box>
      )}

      {!isLoading && !error && reports.length === 0 && (
        <Typography variant="body2" color="text.secondary">
          No reports are available.
        </Typography>
      )}

      <Box
        sx={{
          display: 'grid',
          gridTemplateColumns: { xs: '1fr', md: 'repeat(3, minmax(0, 1fr))' },
          gap: 1.5,
          mb: 3,
        }}
      >
        <TextField
          label="Location ID"
          size="small"
          value={filters.location_id}
          onChange={(event) => setFilters(prev => ({ ...prev, location_id: event.target.value }))}
        />
        <TextField
          label="Coach ID"
          size="small"
          value={filters.coach_id}
          onChange={(event) => setFilters(prev => ({ ...prev, coach_id: event.target.value }))}
        />
        <TextField
          label="Category"
          size="small"
          value={filters.category}
          onChange={(event) => setFilters(prev => ({ ...prev, category: event.target.value }))}
        />
        <TextField
          label="From"
          type="date"
          size="small"
          value={filters.from}
          onChange={(event) => setFilters(prev => ({ ...prev, from: event.target.value }))}
          InputLabelProps={{ shrink: true }}
        />
        <TextField
          label="To"
          type="date"
          size="small"
          value={filters.to}
          onChange={(event) => setFilters(prev => ({ ...prev, to: event.target.value }))}
          InputLabelProps={{ shrink: true }}
        />
        <TextField
          select
          label="Status"
          size="small"
          value={filters.status}
          onChange={(event) => setFilters(prev => ({ ...prev, status: event.target.value }))}
        >
          <MenuItem value="">Any status</MenuItem>
          <MenuItem value="active">Active</MenuItem>
          <MenuItem value="expired">Expired</MenuItem>
          <MenuItem value="cancelled">Cancelled</MenuItem>
          <MenuItem value="completed">Completed</MenuItem>
        </TextField>
      </Box>

      {reports.length > 0 && (
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5, mb: 4 }}>
          {reports.map(report => (
            <ReportCard key={report.id} report={report} onExport={handleExport} />
          ))}
        </Box>
      )}

      {exportJobs.length > 0 && (
        <>
          <Divider sx={{ mb: 2 }} />
          <Typography variant="subtitle1" fontWeight={600} gutterBottom>
            Recent Export Jobs
          </Typography>
          <Box
            component="section"
            sx={{ border: 1, borderColor: 'divider', borderRadius: 1, maxWidth: 600 }}
          >
            {exportJobs.slice(0, 10).map(job => (
              <ExportJobItem key={job.id} job={job} onDownload={handleDownload} isOffline={isOffline} />
            ))}
          </Box>
        </>
      )}
    </PageContainer>
  );
}
