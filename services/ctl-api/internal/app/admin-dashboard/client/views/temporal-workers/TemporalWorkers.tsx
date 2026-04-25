import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router'
import { getTemporalWorkers } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatRelativeDate } from '@/utils/format'

export const TemporalWorkers = () => {
  const { data, isLoading, error } = useQuery({
    queryKey: ['temporal-workers'],
    queryFn: () => getTemporalWorkers(),
    refetchInterval: 15000,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load temporal workers'} />

  const workers = data?.namespace_pollers || []
  const temporalUIUrl = data?.temporal_ui_url || ''

  return (
    <div>
      <h1 className="page-heading">Temporal workers</h1>
      <p className="page-subheading">{workers.length} namespaces (auto-refreshing)</p>

      <div className="mt-4 table-card">
        <table>
          <thead>
            <tr>
              <th>Namespace</th>
              <th>Task queue</th>
              <th>Workflow pollers</th>
              <th>Activity pollers</th>
              <th>WF backlog</th>
              <th>Act backlog</th>
              <th>Health</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {workers.map((worker) => {
              const wfPollerCount = worker.workflow_pollers?.length ?? 0
              const actPollerCount = worker.activity_pollers?.length ?? 0
              const isHealthy = !worker.error && wfPollerCount > 0

              return (
                <tr key={worker.namespace}>
                  <td>
                    <Link
                      to={`/temporal-workers/${encodeURIComponent(worker.namespace)}`}
                      className="text-primary-600 hover:text-primary-700 font-medium"
                    >
                      {worker.namespace}
                    </Link>
                  </td>
                  <td className="text-gray-500 font-mono text-xs">{worker.task_queue}</td>
                  <td className="text-gray-900">{wfPollerCount}</td>
                  <td className="text-gray-900">{actPollerCount}</td>
                  <td className="text-gray-500 font-mono text-xs">
                    {worker.workflow_stats?.approximate_backlog_count ?? '-'}
                  </td>
                  <td className="text-gray-500 font-mono text-xs">
                    {worker.activity_stats?.approximate_backlog_count ?? '-'}
                  </td>
                  <td>
                    <Badge variant="status" status={isHealthy ? 'healthy' : 'unhealthy'}>
                      {isHealthy ? 'Healthy' : 'Unhealthy'}
                    </Badge>
                    {worker.error && (
                      <span className="ml-1 text-[11px] text-red-500 truncate max-w-[200px] inline-block align-middle" title={worker.error}>
                        {worker.error.length > 40 ? worker.error.slice(0, 40) + '...' : worker.error}
                      </span>
                    )}
                  </td>
                </tr>
              )
            })}
            {workers.length === 0 && (
              <tr>
                <td colSpan={7} className="text-center text-gray-500 py-6">No temporal workers found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
