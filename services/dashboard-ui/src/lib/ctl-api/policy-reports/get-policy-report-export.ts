import { api } from '@/lib/api'
import { buildQueryParams } from '@/utils/build-query-params'

export type TPolicyReportExportFormat = 'opa' | 'sarif' | 'pdf'

export type TPolicyReportDownload = {
  content: string
  filename: string
}

export const getPolicyReportExport = ({
  reportId,
  orgId,
  format = 'opa',
}: {
  reportId: string
  orgId: string
  format?: TPolicyReportExportFormat
}) =>
  api<TPolicyReportDownload>({
    path: `policy-reports/${reportId}/export${buildQueryParams({ format })}`,
    orgId,
  })
