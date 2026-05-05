import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router'
import { getRunners } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { truncateId } from '@/utils/format'

export const RunnersList = () => {
  const { data, isLoading, error } = useQuery({
    queryKey: ['runners'],
    queryFn: () => getRunners(),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load runners'} />

  const runners = data?.runners || []

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">Runners</h1>

      <div className="mt-4 overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
          <thead className="">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Display Name</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Install</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Version</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Configs</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
            {runners.map((rv) => (
              <tr key={rv.runner.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Link to={`/runners/${rv.runner.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200 font-medium">
                    {rv.runner.name || truncateId(rv.runner.id)}
                  </Link>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{rv.runner.display_name || '-'}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Badge variant="status" status={rv.process_online ? 'online' : 'offline'}>
                    {rv.process_online ? 'Online' : 'Offline'}
                  </Badge>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Link to={`/installs/${rv.install_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200">
                    {rv.install_name || truncateId(rv.install_id)}
                  </Link>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{rv.version || '-'}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{rv.configs?.length ?? 0}</td>
              </tr>
            ))}
            {runners.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No runners found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
