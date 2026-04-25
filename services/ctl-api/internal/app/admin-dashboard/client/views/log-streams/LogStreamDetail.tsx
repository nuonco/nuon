import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router'
import { getLogStreamDetail, getLogStreamLogs } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

export const LogStreamDetail = () => {
  const { id } = useParams<{ id: string }>()
  const [logsPage, setLogsPage] = useState(1)

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

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-gray-900">Log Stream</h1>
        <p className="mt-1 text-sm text-gray-500 font-mono">{log_stream.id}</p>
        <div className="mt-1 text-sm text-gray-500">
          Org: <Link to={`/orgs/${log_stream.org_id}`} className="text-primary-600 hover:text-primary-800 font-mono">{truncateId(log_stream.org_id)}</Link>
        </div>
        <div className="mt-1 text-sm text-gray-500">
          Owner: <span className="font-mono text-xs">{truncateId(log_stream.owner_id)}</span> ({log_stream.owner_type})
        </div>
        <div className="mt-1 text-sm text-gray-500">Created: {formatDate(log_stream.created_at)}</div>
      </div>

      {/* Logs */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Logs</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Timestamp</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Severity</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Body</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {(logsData?.logs || []).map((log, i) => (
                <tr key={i}>
                  <td className="whitespace-nowrap px-4 py-3 text-xs text-gray-500 font-mono">{formatDate(log.timestamp)}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-xs">
                    <Badge variant="status" status={log.severity_text === 'ERROR' ? 'error' : log.severity_text === 'WARN' ? 'warning' : 'healthy'}>
                      {log.severity_text}
                    </Badge>
                  </td>
                  <td className="px-4 py-3 text-xs text-gray-900 max-w-lg truncate">{log.body}</td>
                </tr>
              ))}
              {(!logsData?.logs || logsData.logs.length === 0) && (
                <tr>
                  <td colSpan={3} className="px-4 py-8 text-center text-sm text-gray-500">No logs</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        {logsData && (
          <Pagination page={logsPage} totalPages={logsData.total_pages} onPageChange={setLogsPage} />
        )}
      </div>
    </div>
  )
}
