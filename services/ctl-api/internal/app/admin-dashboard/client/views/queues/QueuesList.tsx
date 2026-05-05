import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useSearchParams } from 'react-router'
import { getQueues } from '@/lib/admin-api'
import { Pagination } from '@/components/common/Pagination'
import { SearchInput } from '@/components/common/SearchInput'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

export const QueuesList = () => {
  const [searchParams] = useSearchParams()
  const ownerID = searchParams.get('owner_id') || undefined
  const ownerType = searchParams.get('owner_type') || undefined
  const [search, setSearch] = useState('')
  const [name, setName] = useState('')
  const [namespace, setNamespace] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading, error } = useQuery({
    queryKey: ['queues', search, name, namespace, ownerID, ownerType, page],
    queryFn: () => getQueues({
      search,
      name: name || undefined,
      namespace: namespace || undefined,
      owner_id: ownerID,
      owner_type: ownerType,
      page,
    }),
    refetchInterval: 20000,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load queues'} />

  const { queues = [], total_pages = 1 } = data || {}

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">Queues</h1>

      <div className="mt-4 flex flex-col gap-4 sm:flex-row sm:items-center">
        <div className="w-full sm:w-64">
          <SearchInput value={search} onChange={(v) => { setSearch(v); setPage(1) }} placeholder="Search queues..." />
        </div>
        <input
          type="text"
          value={name}
          onChange={(e) => { setName(e.target.value); setPage(1) }}
          placeholder="Filter by name..."
          className="block w-48 rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 focus:ring-2 focus:ring-primary-600 dark:focus:ring-primary-500"
        />
        <input
          type="text"
          value={namespace}
          onChange={(e) => { setNamespace(e.target.value); setPage(1) }}
          placeholder="Filter by namespace..."
          className="block w-48 rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 focus:ring-2 focus:ring-primary-600 dark:focus:ring-primary-500"
        />
      </div>

      <div className="mt-4 overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
          <thead className="bg-gray-50 dark:bg-gray-900">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Owner</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Emitters</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-800 bg-white dark:bg-gray-900">
            {queues.map((queue) => (
              <tr key={queue.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Link to={`/queues/${queue.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200 font-mono">
                    {truncateId(queue.id)}
                  </Link>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{queue.name}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
                  <span className="font-mono text-xs">{truncateId(queue.owner_id)}</span>
                  <span className="ml-1 text-xs text-gray-400 dark:text-gray-500">({queue.owner_type})</span>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{queue.emitters?.length ?? 0}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDate(queue.created_at)}</td>
              </tr>
            ))}
            {queues.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No queues found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={total_pages} onPageChange={setPage} />
    </div>
  )
}
