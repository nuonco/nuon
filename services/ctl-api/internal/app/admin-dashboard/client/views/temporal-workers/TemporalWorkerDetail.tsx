import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { getTemporalWorkerDetail } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate } from '@/utils/format'

export const TemporalWorkerDetail = () => {
  const { namespace } = useParams<{ namespace: string }>()

  const { data, isLoading, error } = useQuery({
    queryKey: ['temporal-worker', namespace],
    queryFn: () => getTemporalWorkerDetail(namespace!),
    enabled: !!namespace,
    refetchInterval: 10000,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load temporal worker'} />
  if (!data?.info) return null

  const info = data.info
  const temporalUIUrl = data.temporal_ui_url || ''
  const wfPollerCount = info.workflow_pollers?.length ?? 0
  const actPollerCount = info.activity_pollers?.length ?? 0
  const isHealthy = !info.error && wfPollerCount > 0

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="page-heading">{info.namespace}</h1>
        <div className="mt-2 flex flex-wrap items-center gap-3 text-sm">
          <Badge variant="status" status={isHealthy ? 'healthy' : 'unhealthy'}>
            {isHealthy ? 'Healthy' : 'Unhealthy'}
          </Badge>
          <span className="text-gray-500 dark:text-gray-400">Task queue: <span className="font-mono">{info.task_queue}</span></span>
          <span className="text-gray-500 dark:text-gray-400">Pollers: {wfPollerCount + actPollerCount}</span>
          {temporalUIUrl && (
            <a href={`${temporalUIUrl}/namespaces/${namespace}`} target="_blank" rel="noopener noreferrer" className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 text-xs">
              Open in Temporal UI &rarr;
            </a>
          )}
        </div>
        {info.error && (
          <div className="mt-2 rounded-md bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 p-3 text-sm text-red-700 dark:text-red-300">{info.error}</div>
        )}
      </div>

      {/* Workflow Pollers */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Workflow pollers ({wfPollerCount})</h2>
        <div className="mt-2 table-card">
          <table>
            <thead>
              <tr>
                <th>Identity</th>
                <th>Last access</th>
                <th>Rate/s</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {(info.workflow_pollers || []).map((poller, i) => (
                <tr key={i}>
                  <td className="text-gray-900 dark:text-gray-100 break-all text-xs font-mono">{poller.identity}</td>
                  <td className="text-gray-500 dark:text-gray-400">{formatDate(poller.last_access_time)}</td>
                  <td className="text-gray-500 dark:text-gray-400 font-mono">{poller.rate_per_second?.toFixed(2)}</td>
                </tr>
              ))}
              {wfPollerCount === 0 && (
                <tr><td colSpan={3} className="text-center text-gray-500 dark:text-gray-400 py-4">No workflow pollers</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Activity Pollers */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Activity pollers ({actPollerCount})</h2>
        <div className="mt-2 table-card">
          <table>
            <thead>
              <tr>
                <th>Identity</th>
                <th>Last access</th>
                <th>Rate/s</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {(info.activity_pollers || []).map((poller, i) => (
                <tr key={i}>
                  <td className="text-gray-900 dark:text-gray-100 break-all text-xs font-mono">{poller.identity}</td>
                  <td className="text-gray-500 dark:text-gray-400">{formatDate(poller.last_access_time)}</td>
                  <td className="text-gray-500 dark:text-gray-400 font-mono">{poller.rate_per_second?.toFixed(2)}</td>
                </tr>
              ))}
              {actPollerCount === 0 && (
                <tr><td colSpan={3} className="text-center text-gray-500 dark:text-gray-400 py-4">No activity pollers</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Queue Stats */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {info.workflow_stats && (
          <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
            <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Workflow queue stats</h2>
            <dl className="mt-3 space-y-2 text-sm">
              <div className="flex justify-between">
                <dt className="text-gray-500 dark:text-gray-400">Backlog count</dt>
                <dd className="font-mono text-gray-900 dark:text-gray-100">{info.workflow_stats.approximate_backlog_count}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500 dark:text-gray-400">Backlog age</dt>
                <dd className="font-mono text-gray-900 dark:text-gray-100">{info.workflow_stats.approximate_backlog_age ? `${(info.workflow_stats.approximate_backlog_age / 1000000000).toFixed(1)}s` : '-'}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500 dark:text-gray-400">Tasks add rate</dt>
                <dd className="font-mono text-gray-900 dark:text-gray-100">{info.workflow_stats.tasks_add_rate?.toFixed(2)}/s</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500 dark:text-gray-400">Tasks dispatch rate</dt>
                <dd className="font-mono text-gray-900 dark:text-gray-100">{info.workflow_stats.tasks_dispatch_rate?.toFixed(2)}/s</dd>
              </div>
            </dl>
          </div>
        )}
        {info.activity_stats && (
          <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
            <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Activity queue stats</h2>
            <dl className="mt-3 space-y-2 text-sm">
              <div className="flex justify-between">
                <dt className="text-gray-500 dark:text-gray-400">Backlog count</dt>
                <dd className="font-mono text-gray-900 dark:text-gray-100">{info.activity_stats.approximate_backlog_count}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500 dark:text-gray-400">Backlog age</dt>
                <dd className="font-mono text-gray-900 dark:text-gray-100">{info.activity_stats.approximate_backlog_age ? `${(info.activity_stats.approximate_backlog_age / 1000000000).toFixed(1)}s` : '-'}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500 dark:text-gray-400">Tasks add rate</dt>
                <dd className="font-mono text-gray-900 dark:text-gray-100">{info.activity_stats.tasks_add_rate?.toFixed(2)}/s</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500 dark:text-gray-400">Tasks dispatch rate</dt>
                <dd className="font-mono text-gray-900 dark:text-gray-100">{info.activity_stats.tasks_dispatch_rate?.toFixed(2)}/s</dd>
              </div>
            </dl>
          </div>
        )}
        {!info.workflow_stats && !info.activity_stats && (
          <p className="text-sm text-gray-500 dark:text-gray-400">No queue stats available</p>
        )}
      </div>
    </div>
  )
}
