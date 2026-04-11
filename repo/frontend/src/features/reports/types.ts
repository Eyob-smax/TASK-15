import type {
  ReportDefinition,
  ExportJob,
  ExportFormat,
  ExportStatus,
} from '@/lib/types';

export type {
  ReportDefinition,
  ExportJob,
  ExportFormat,
  ExportStatus,
};

export interface ReportFilters {
  report_type?: string;
  search?: string;
  is_active?: boolean;
  page?: number;
  page_size?: number;
}

export interface ExportRequestData {
  report_id: string;
  format: ExportFormat;
  parameters?: Record<string, string>;
}

export interface ReportParameterDefinition {
  name: string;
  label: string;
  type: 'string' | 'number' | 'date' | 'select';
  required: boolean;
  options?: { label: string; value: string }[];
  default_value?: string;
}
