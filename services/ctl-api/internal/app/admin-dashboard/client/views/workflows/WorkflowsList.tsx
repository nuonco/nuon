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

const WORKFLOW_TYPES = [
  'provision',
  'deprovision',
  'manual_deploy',
  'deploy_components',
  'input_update',
  'reprovision_sandbox',
  'teardown_components',
  'action_workflow_run',
  'sync_secrets',
  'drift_run',
]

export const WorkflowsList = () => {
  const [search, setSearch] = useState('')
  const [type, setType] = useState('')
  const [sort, setSort] = useState<'newest' | 'oldest'>('newest')
  const [page, setPage] = useState(1)

  const { data, isLoading, error } = useQuery({
    queryKey: ['workflows', search, type, sort, page],
    queryFn: () => getWorkflows({ search, type: type || undefined, sort, page }),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load workflows'} />

  const workflows = data?.workflows || []
  const totalPages = data?.total_pages || 1

  return (
    <div>
      <h1 className="page-heading">Workflows</h1>

      <div className="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center">
        <div className="w-full sm:w-64">
          <SearchInput value={search} onChange={(v) => { setSearch(v); setPage(1) }} placeholder="Search by ID or owner ID..." />
        </div>
        <select
          value={type}
          onChange={(e) => { setType(e.target.value); setPage(1) }}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700"
        >
          <option value="">All types</option>
          {WORKFLOW_TYPES.map((t) => (
            <option key={t} value={t}>{t}</option>
          ))}
        </select>
        <select
          value={sort}
          onChange={(e) => { setSort(e.target.value as 'newest' | 'oldest'); setPage(1) }}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700"
        >
          <option value="newest">Newest first</option>
          <option value="oldest">Oldest first</option>
        </select>
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
          <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
            {workflows.map((wf: any) => {
              const status = getStatus(wf.status)
              return (
                <tr key={wf.id}>
                  <td>
                    <Link to={`/workflows/${wf.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 font-mono text-xs">
                      {truncateId(wf.id)}
                    </Link>
                  </td>
                  <td className="font-mono text-xs text-gray-900 dark:text-gray-100">{wf.type}</td>
                  <td className="text-gray-500 dark:text-gray-400">
                    <Link to={`/installs/${wf.owner_id}`} className="font-mono text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
                      {truncateId(wf.owner_id)}
                    </Link>
                    <span className="ml-1 text-[11px] text-gray-400 dark:text-gray-500">({wf.owner_type})</span>
                  </td>
                  <td className="text-gray-500 dark:text-gray-400 text-xs">{wf.steps?.length ?? 0}</td>
                  <td>
                    <Badge variant="status" status={status}>{status || '-'}</Badge>
                  </td>
                  <td className="text-gray-500 dark:text-gray-400 text-xs">
                    {wf.created_by?.email ? (
                      <Link to={`/accounts/${wf.created_by_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">{wf.created_by.email}</Link>
                    ) : (
                      <span className="font-mono">{truncateId(wf.created_by_id)}</span>
                    )}
                  </td>
                  <td className="text-gray-500 dark:text-gray-400">{formatDate(wf.created_at)}</td>
                </tr>
              )
            })}
            {workflows.length === 0 && (
              <tr>
                <td colSpan={7} className="text-center text-gray-500 dark:text-gray-400 py-6">No workflows found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
    </div>
  )
}
