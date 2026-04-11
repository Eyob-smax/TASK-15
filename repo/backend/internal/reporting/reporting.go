// Package reporting provides KPI aggregation, report data generation, and
// export file creation for the FitCommerce platform.
//
// This package will contain the following implementations (Prompt 7):
//
//   - KPI aggregation: computes dashboard metrics including member growth rate,
//     churn rate, renewal rate, member engagement scores, class fill rates, and
//     coach productivity indices. Supports filtering by location and time period.
//
//   - CSV generation: renders report data into CSV format with proper escaping,
//     UTF-8 BOM for Excel compatibility, and configurable column selection.
//
//   - PDF generation: renders report data into formatted PDF documents with
//     headers, footers, tables, and branding elements.
//
//   - Export file management: handles temporary file creation, storage path
//     resolution, and cleanup of expired export files.
//
// All report generators accept a ReportDefinition and filter parameters, query
// the necessary repositories, and produce the output in the requested format.
package reporting
