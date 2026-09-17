import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link } from 'react-router'
import { getAppBranches } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { SearchInput } from '@/components/common/SearchInput'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, formatRelativeDate, truncateId } from '@/utils/format'

const MANAGED_BY_OPTIONS = [
  { label: 'All', value: '' },
  { label: 'Config', value: 'config' },
  { label: 'Manually', value: 'manually' },
]

const SORT_OPTIONS = [
  { label: 'Newest', value: 'newest' },
  { label: 'Oldest', value: 'oldest' },
]

export const AppBranchesList = () => {
  const [search, setSearch] = useState('')
  const [managedBy, setManagedBy] = useState('')
  const [sort, setSort] = useState('newest')
  const [showDeleted, setShowDeleted] = useState(false)
  const [page, setPage] = useState(1)

  const { data, isLoading, error } = useQuery({
    queryKey: ['app-branches', search, managedBy, sort, showDeleted, page],
    queryFn: () =>
      getAppBranches({
        search,
        managed_by: managedBy || undefined,
        sort,
        show_deleted: showDeleted ? 'true' : undefined,
        page,
      }),
    refetchInterval: 20000,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load app branches'} />

  const { app_branches = [], total_pages = 1 } = data || {}

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">App branches</h1>

      <div className="mt-4 flex flex-col gap-4 sm:flex-row sm:items-center sm:flex-wrap">
        <div className="w-full sm:w-64">
          <SearchInput
            value={search}
            onChange={(v) => { setSearch(v); setPage(1) }}
            placeholder="Search by name, abr/app/org ID..."
          />
        </div>
        <div className="flex gap-2">
          {MANAGED_BY_OPTIONS.map((opt) => (
            <button
              key={opt.value}
              onClick={() => { setManagedBy(opt.value); setPage(1) }}
              className={`rounded-md px-3 py-1.5 text-sm font-medium ${
                managedBy === opt.value
                  ? 'bg-primary-600 dark:bg-primary-500 text-white'
                  : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'
              }`}
            >
              {opt.label}
            </button>
          ))}
        </div>
        <select
          value={sort}
          onChange={(e) => { setSort(e.target.value); setPage(1) }}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 focus:ring-2 focus:ring-primary-600 dark:focus:ring-primary-500"
        >
          {SORT_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <label className="flex items-center gap-1 text-sm text-gray-700 dark:text-gray-300 cursor-pointer">
          <input
            type="checkbox"
            checked={showDeleted}
            onChange={(e) => { setShowDeleted(e.target.checked); setPage(1) }}
            className="rounded border-gray-300 dark:border-gray-700 text-primary-600 dark:text-primary-400 focus:ring-primary-500"
          />
          Show deleted
        </label>
      </div>

      <div className="mt-4 overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
          <thead>
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Org</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">App</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Managed by</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Workflows</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Latest run</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
            {app_branches.map(({ app_branch: branch, org_name, app_name }) => (
              <tr key={branch.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Link to={`/app-branches/${branch.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200 font-medium">
                    {branch.name || truncateId(branch.id)}
                  </Link>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400 font-mono">{truncateId(branch.id)}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  <Link to={`/orgs/${branch.org_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200">
                    {org_name || truncateId(branch.org_id)}
                  </Link>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
                  {app_name || truncateId(branch.app_id)}
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  {branch.managed_by && <Badge>{branch.managed_by}</Badge>}
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{branch.workflow_count ?? 0}</td>
                <td className="whitespace-nowrap px-4 py-3 text-sm">
                  {branch.latest_run ? (
                    <span className="flex items-center gap-2">
                      <Badge variant="status" status={branch.latest_run.status}>{branch.latest_run.status || '-'}</Badge>
                      <span className="text-xs text-gray-500 dark:text-gray-400">{formatRelativeDate(branch.latest_run.created_at)}</span>
                    </span>
                  ) : (
                    <span className="text-xs text-gray-400 dark:text-gray-500">No runs</span>
                  )}
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDate(branch.created_at)}</td>
              </tr>
            ))}
            {app_branches.length === 0 && (
              <tr>
                <td colSpan={8} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No app branches found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={total_pages} onPageChange={setPage} />
    </div>
  )
}
