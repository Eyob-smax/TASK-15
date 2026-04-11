import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, type ApiEnvelope } from '@/lib/api-client';
import { enqueueOfflineMutation } from '@/lib/offline-cache';
import type { ReportDefinition, ExportJob, ExportFormat } from '@/lib/types';

// ─── Shared ──────────────────────────────────────────────────────────────────

// ─── Body types ───────────────────────────────────────────────────────────────

interface RunExportBody {
  report_id: string;
  format: ExportFormat;
  parameters?: Record<string, string>;
}

// ─── Report hooks ─────────────────────────────────────────────────────────────

export function useReportList() {
  return useQuery({
    queryKey: ['reports'],
    queryFn: () =>
      apiClient.get<{ data: ReportDefinition[] }>('/reports'),
  });
}

export function useReport(id: string | undefined) {
  return useQuery({
    queryKey: ['reports', id, 'data'],
    queryFn: async () => {
      const response = await apiClient.get<ApiEnvelope<unknown>>(`/reports/${id}/data`);
      return response.data;
    },
    enabled: Boolean(id),
  });
}

// ─── Export hooks ─────────────────────────────────────────────────────────────

export function useRunExport() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: RunExportBody) => {
      if (typeof navigator !== 'undefined' && !navigator.onLine) {
        const queuedID = globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-run-export`;
        await enqueueOfflineMutation({
          id: queuedID,
          type: 'run-export',
          payload: body as unknown as Record<string, unknown>,
          createdAt: Date.now(),
        });
        return {
          id: queuedID,
          report_id: body.report_id,
          format: body.format,
          filename: `queued_${body.report_id}.${body.format}`,
          status: 'pending',
          created_by: 'offline-queue',
          created_at: new Date().toISOString(),
          completed_at: null,
        } as ExportJob;
      }
      const response = await apiClient.post<ApiEnvelope<ExportJob>>('/exports', body);
      return response.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['exports'] });
    },
  });
}

export function useExport(id: string | undefined) {
  return useQuery({
    queryKey: ['exports', id],
    queryFn: async () => {
      const response = await apiClient.get<ApiEnvelope<ExportJob>>(`/exports/${id}`);
      return response.data;
    },
    enabled: Boolean(id),
  });
}
