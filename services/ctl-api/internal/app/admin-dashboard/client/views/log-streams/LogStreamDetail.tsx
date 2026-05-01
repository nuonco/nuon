import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router'
import { getLogStreamDetail, getLogStreamLogs } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

function severityColor(severity: string): string {
  switch (severity?.toUpperCase()) {
    case 'ERROR':
    case 'FATAL':
      return 'text-red-600'
    case 'WARN':
    case 'WARNING':
      return 'text-orange-500'
    case 'INFO':
      return 'text-blue-600'
    case 'DEBUG':
    case 'TRACE':
      return 'text-gray-400'
    default:
      return 'text-gray-600'
  }
}

function severityBadgeStatus(severity: string): string {
  switch (severity?.toUpperCase()) {
    case 'ERROR':
    case 'FATAL':
      return 'error'
    case 'WARN':
    case 'WARNING':
      return 'warning'
    case 'INFO':
      return 'healthy'
    case 'DEBUG':
    case 'TRACE':
      return 'unknown'
    default:
      return 'unknown'
  }
}

function AttributeTable({ title, attrs }: { title: string; attrs: Record<string, string> | null | undefined }) {
  if (!attrs || Object.keys(attrs).length === 0) return null
  return (
    <div>
      <h4 className="text-xs font-semibold text-gray-500 uppercase mb-1">{title}</h4>
      <div className="rounded border border-gray-200 overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-3 py-1.5 text-left text-xs font-medium text-gray-500">Key</th>
              <th className="px-3 py-1.5 text-left text-xs font-medium text-gray-500">Value</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {Object.entries(attrs).map(([key, value]) => (
              <tr key={key}>
                <td className="px-3 py-1.5 text-xs font-mono text-gray-700 whitespace-nowrap">{key}</td>
                <td className="px-3 py-1.5 text-xs font-mono text-gray-900 break-all">{value}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

export const LogStreamDetail = () => {
  const { id } = useParams<{ id: string }>()
  const [logsPage, setLogsPage] = useState(1)
  const [expandedRows, setExpandedRows] = useState<Set<number>>(new Set())

  const { data, isLoading, error } = useQuery({
    queryKey: ['log-stream', id],
    queryFn: () => getLogStreamDetail(id!),
    enabled: !!id,
  })

  const { data: logsData } = useQuery({
    queryKey: ['log-stream-logs', id, logsPage],
    queryFn: () => getLogStreamLogs(id!, { page: logsPage }),
    enabled: !!id,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load log stream'} />
  if (!data) return null

  const { log_stream } = data

  const toggleRow = (index: number) => {
    setExpandedRows((prev) => {
      const next = new Set(prev)
      if (next.has(index)) {
        next.delete(index)
      } else {
        next.add(index)
      }
      return next
    })
  }

  const logs = logsData?.logs || data?.logs || []
  const totalPages = logsData?.total_pages || data?.total_pages || 1

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <nav className="text-sm text-gray-500">
        <Link to="/log-streams" className="text-primary-600 hover:text-primary-800">Log Streams</Link>
        <span className="mx-1">/</span>
        <span className="font-mono">{truncateId(log_stream.id)}</span>
      </nav>

      {/* Header */}
      <div>
        <h1 className="text-xl font-bold text-gray-900">Log Stream</h1>
        <p className="mt-1 text-sm text-gray-500 font-mono">{log_stream.id}</p>
        <div className="mt-2 grid grid-cols-1 gap-1 sm:grid-cols-2 lg:grid-cols-4 text-sm text-gray-500">
          <div>
            Org: <Link to={`/orgs/${log_stream.org_id}`} className="text-primary-600 hover:text-primary-800 font-mono">{truncateId(log_stream.org_id)}</Link>
          </div>
          <div>
            Owner: <span className="font-mono text-xs">{truncateId(log_stream.owner_id)}</span>{' '}
            <Badge variant="default">{log_stream.owner_type}</Badge>
          </div>
          <div>Created: {formatDate(log_stream.created_at)}</div>
        </div>
      </div>

      {/* Logs */}
      <div className="table-card rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">
          Logs
          {logs.length > 0 && (
            <span className="ml-2 text-xs font-normal text-gray-500">
              (page {logsData?.page || data?.page || logsPage} of {totalPages})
            </span>
          )}
        </h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="w-8 px-2 py-3"></th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Timestamp</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Severity</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Service</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Body</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {logs.map((log: any, i: number) => {
                const isExpanded = expandedRows.has(i)
                return (
                  <>
                    <tr
                      key={`row-${i}`}
                      className="hover:bg-gray-50 cursor-pointer"
                      onClick={() => toggleRow(i)}
                    >
                      <td className="px-2 py-3 text-xs text-gray-400">
                        {isExpanded ? '\u25BC' : '\u25B6'}
                      </td>
                      <td className="whitespace-nowrap px-4 py-3 text-xs text-gray-500 font-mono">
                        {formatDate(log.timestamp)}
                      </td>
                      <td className="whitespace-nowrap px-4 py-3 text-xs">
                        <span className={`font-semibold ${severityColor(log.severity_text)}`}>
                          {log.severity_text || '-'}
                        </span>
                      </td>
                      <td className="whitespace-nowrap px-4 py-3 text-xs text-gray-700">
                        {log.service_name || '-'}
                      </td>
                      <td className="px-4 py-3 text-xs text-gray-900 max-w-lg truncate">
                        {log.body}
                      </td>
                    </tr>
                    {isExpanded && (
                      <tr key={`detail-${i}`}>
                        <td colSpan={5} className="px-4 py-4 bg-gray-50">
                          <div className="space-y-4">
                            {/* Core details */}
                            <div className="grid grid-cols-2 gap-x-8 gap-y-2 sm:grid-cols-3 lg:grid-cols-4">
                              {[
                                { label: 'Service Name', value: log.service_name },
                                { label: 'Scope Name', value: log.scope_name },
                                { label: 'Trace ID', value: log.trace_id },
                                { label: 'Span ID', value: log.span_id },
                                { label: 'Runner Job ID', value: log.runner_job_id },
                                { label: 'Runner Job Execution Step', value: log.runner_job_execution_step },
                                { label: 'Severity Number', value: log.severity_number?.toString() },
                                { label: 'Scope Version', value: log.scope_version },
                              ]
                                .filter(({ value }) => value)
                                .map(({ label, value }) => (
                                  <div key={label}>
                                    <div className="text-xs text-gray-500">{label}</div>
                                    <div className="text-xs font-mono text-gray-900 break-all">{value}</div>
                                  </div>
                                ))}
                            </div>

                            {/* Full body */}
                            {log.body && log.body.length > 80 && (
                              <div>
                                <h4 className="text-xs font-semibold text-gray-500 uppercase mb-1">Full Body</h4>
                                <pre className="text-xs font-mono text-gray-900 bg-white border border-gray-200 rounded p-3 overflow-x-auto whitespace-pre-wrap break-all">
                                  {log.body}
                                </pre>
                              </div>
                            )}

                            {/* Attribute tables */}
                            <AttributeTable title="Resource Attributes" attrs={log.resource_attributes} />
                            <AttributeTable title="Scope Attributes" attrs={log.scope_attributes} />
                            <AttributeTable title="Log Attributes" attrs={log.log_attributes} />
                          </div>
                        </td>
                      </tr>
                    )}
                  </>
                )
              })}
              {logs.length === 0 && (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-sm text-gray-500">No logs</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        <Pagination page={logsData?.page || logsPage} totalPages={totalPages} onPageChange={setLogsPage} />
      </div>
    </div>
  )
}
