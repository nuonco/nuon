import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link } from 'react-router'
import { getLogStreams } from '@/lib/admin-api'
import { SearchInput } from '@/components/common/SearchInput'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

export const LogStreams = () => {
  const [search, setSearch] = useState('')

  const { data, isLoading, error } = useQuery({
    queryKey: ['log-streams', search],
    queryFn: () => getLogStreams({ search }),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load log streams'} />

  const logStreams = data?.log_streams || []

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900">Log Streams</h1>

      <div className="mt-4 w-full sm:w-64">
        <SearchInput value={search} onChange={setSearch} placeholder="Search log streams..." />
      </div>

      <div className="mt-4 overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Org ID</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Owner</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 bg-white">
            {logStreams.map((ls) => (
              <tr key={ls.id} className="hover:bg-gray-50">
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Link to={`/log-streams/${ls.id}`} className="text-primary-600 hover:text-primary-800 font-mono">
                    {truncateId(ls.id)}
                  </Link>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Link to={`/orgs/${ls.org_id}`} className="text-primary-600 hover:text-primary-800 font-mono">
                    {truncateId(ls.org_id)}
                  </Link>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">
                  <span className="font-mono text-xs">{truncateId(ls.owner_id)}</span>
                  <span className="ml-1 text-xs text-gray-400">({ls.owner_type})</span>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(ls.created_at)}</td>
              </tr>
            ))}
            {logStreams.length === 0 && (
              <tr>
                <td colSpan={4} className="px-4 py-8 text-center text-sm text-gray-500">No log streams found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
