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

function getStatus(status: any): string {
  if (!status) return ''
  if (typeof status === 'string') return status
  if (typeof status === 'object' && status.status) return String(status.status)
  return String(status)
}

export const WorkflowsList = () => {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading, error } = useQuery({
    queryKey: ['workflows', search, page],
    queryFn: () => getWorkflows({ search, page }),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load workflows'} />

  const workflows = data?.workflows || []
  const totalPages = data?.total_pages || 1

  return (
    <div>
      <h1 className="page-heading">Workflows</h1>

      <div className="mt-4 w-full sm:w-64">
        <SearchInput value={search} onChange={(v) => { setSearch(v); setPage(1) }} placeholder="Search by ID or owner ID..." />
      </div>

      <div className="mt-4 table-card">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>Type</th>
              <th>Owner</th>
              <th>Steps</th>
              <th>Status</th>
              <th>Created by</th>
              <th>Created</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {workflows.map((wf: any) => {
              const status = getStatus(wf.status)
              return (
                <tr key={wf.id}>
                  <td>
                    <Link to={`/workflows/${wf.id}`} className="text-primary-600 hover:text-primary-700 font-mono text-xs">
                      {truncateId(wf.id)}
                    </Link>
                  </td>
                  <td className="font-mono text-xs text-gray-900">{wf.type}</td>
                  <td className="text-gray-500">
                    <Link to={`/installs/${wf.owner_id}`} className="font-mono text-xs text-primary-600 hover:text-primary-700">
                      {truncateId(wf.owner_id)}
                    </Link>
                    <span className="ml-1 text-[11px] text-gray-400">({wf.owner_type})</span>
                  </td>
                  <td className="text-gray-500 text-xs">{wf.steps?.length ?? 0}</td>
                  <td>
                    <Badge variant="status" status={status}>{status || '-'}</Badge>
                  </td>
                  <td className="text-gray-500 text-xs">
                    {wf.created_by?.email ? (
                      <Link to={`/accounts/${wf.created_by_id}`} className="text-primary-600 hover:text-primary-700">{wf.created_by.email}</Link>
                    ) : (
                      <span className="font-mono">{truncateId(wf.created_by_id)}</span>
                    )}
                  </td>
                  <td className="text-gray-500">{formatDate(wf.created_at)}</td>
                </tr>
              )
            })}
            {workflows.length === 0 && (
              <tr>
                <td colSpan={7} className="text-center text-gray-500 py-6">No workflows found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
    </div>
  )
}
