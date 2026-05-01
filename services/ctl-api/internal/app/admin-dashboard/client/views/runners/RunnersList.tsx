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
      <h1 className="text-xl font-bold text-gray-900">Runners</h1>

      <div className="mt-4 overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Display Name</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Install</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Version</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Configs</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 bg-white">
            {runners.map((rv) => (
              <tr key={rv.runner.id} className="hover:bg-gray-50">
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Link to={`/runners/${rv.runner.id}`} className="text-primary-600 hover:text-primary-800 font-medium">
                    {rv.runner.name || truncateId(rv.runner.id)}
                  </Link>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{rv.runner.display_name || '-'}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Badge variant="status" status={rv.process_online ? 'online' : 'offline'}>
                    {rv.process_online ? 'Online' : 'Offline'}
                  </Badge>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Link to={`/installs/${rv.install_id}`} className="text-primary-600 hover:text-primary-800">
                    {rv.install_name || truncateId(rv.install_id)}
                  </Link>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{rv.version || '-'}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{rv.configs?.length ?? 0}</td>
              </tr>
            ))}
            {runners.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-sm text-gray-500">No runners found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
