import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router'
import { getTemporalWorkers } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'

export const TemporalWorkers = () => {
  const { data, isLoading, error } = useQuery({
    queryKey: ['temporal-workers'],
    queryFn: () => getTemporalWorkers(),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load temporal workers'} />

  const workers = data?.workers || []

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900">Temporal Workers</h1>

      <div className="mt-4 overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Namespace</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Task Queue</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Workflow Pollers</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Activity Pollers</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Health</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 bg-white">
            {workers.map((worker) => (
              <tr key={worker.namespace} className="hover:bg-gray-50">
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Link
                    to={`/temporal-workers/${encodeURIComponent(worker.namespace)}`}
                    className="text-primary-600 hover:text-primary-800 font-medium"
                  >
                    {worker.namespace}
                  </Link>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{worker.task_queue}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{worker.workflow_pollers?.length ?? 0}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{worker.activity_pollers?.length ?? 0}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Badge variant="status" status={worker.is_healthy ? 'healthy' : 'unhealthy'}>
                    {worker.is_healthy ? 'Healthy' : 'Unhealthy'}
                  </Badge>
                  {worker.error && (
                    <span className="ml-2 text-xs text-red-500">{worker.error}</span>
                  )}
                </td>
              </tr>
            ))}
            {workers.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-sm text-gray-500">No temporal workers found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
