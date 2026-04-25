import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link } from 'react-router'
import { getWorkflows } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { SearchInput } from '@/components/common/SearchInput'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

export const WorkflowsList = () => {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading, error } = useQuery({
    queryKey: ['workflows', search, page],
    queryFn: () => getWorkflows({ search, page }),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load workflows'} />

  const { workflows = [], total_pages = 1 } = data || {}

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900">Workflows</h1>

      <div className="mt-4 w-full sm:w-64">
        <SearchInput value={search} onChange={(v) => { setSearch(v); setPage(1) }} placeholder="Search workflows..." />
      </div>

      <div className="mt-4 overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Owner</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 bg-white">
            {workflows.map((wf) => (
              <tr key={wf.id} className="hover:bg-gray-50">
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Link to={`/workflows/${wf.id}`} className="text-primary-600 hover:text-primary-800 font-mono">
                    {truncateId(wf.id)}
                  </Link>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{wf.type}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">
                  <span className="font-mono text-xs">{truncateId(wf.owner_id)}</span>
                  <span className="ml-1 text-xs text-gray-400">({wf.owner_type})</span>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Badge variant="status" status={wf.status}>{wf.status}</Badge>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(wf.created_at)}</td>
              </tr>
            ))}
            {workflows.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-sm text-gray-500">No workflows found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={total_pages} onPageChange={setPage} />
    </div>
  )
}
