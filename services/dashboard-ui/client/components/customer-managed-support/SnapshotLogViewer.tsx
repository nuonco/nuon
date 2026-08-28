import { useEffect, useMemo, useState } from 'react'
import { SSELogs } from '@/components/log-stream/SSELogs/SSELogs'
import { useLogFilters } from '@/hooks/use-log-filters'
import type { TCustomerManagedSnapshotJobLog } from '@/lib'
import type { TOTELLog } from '@/types'

const severityNumbers: Record<string, number> = {
  trace: 1,
  debug: 5,
  info: 9,
  warn: 13,
  warning: 13,
  error: 17,
  fatal: 21,
}

export const CustomerManagedSnapshotLogViewer = ({
  log,
  capturedAt,
}: {
  log: TCustomerManagedSnapshotJobLog
  capturedAt?: string
}) => {
  const [activeLogId, setActiveLogId] = useState<string>()
  const records = useMemo<TOTELLog[]>(
    () =>
      log.entries.map((entry, index) => {
        const attributes = Object.fromEntries(
          Object.entries(entry.fields ?? {}).map(([key, value]) => [
            key,
            typeof value === 'string' ? value : JSON.stringify(value),
          ])
        )
        const level = entry.level?.toLowerCase() ?? ''
        return {
          id: `${log.job_id}-${index}`,
          body: entry.msg ?? '',
          timestamp: entry.time ?? log.started_at ?? capturedAt,
          severity_text: level
            ? `${level.charAt(0).toUpperCase()}${level.slice(1)}`
            : 'Info',
          severity_number: severityNumbers[level] ?? 9,
          service_name:
            attributes['service.name'] ?? attributes.service_name ?? '',
          scope_name: 'oteljob',
          runner_job_id: log.job_id,
          log_attributes: attributes,
        }
      }),
    [capturedAt, log]
  )
  const filters = useLogFilters(records)
  const activeLog = records.find(({ id }) => id === activeLogId)

  useEffect(() => {
    if (!records.some(({ id }) => id === activeLogId)) setActiveLogId(undefined)
  }, [records, activeLogId])

  return (
    <SSELogs
      filteredLogs={filters.filteredLogs ?? undefined}
      filters={filters}
      activeLog={activeLog}
      handleActiveLog={setActiveLogId}
      isLoading={false}
      isConnected={false}
      showDownload={false}
    />
  )
}
