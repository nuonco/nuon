import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link } from 'react-router'
import { getInstalls } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { SearchInput } from '@/components/common/SearchInput'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

function getStatus(s: any): string {
  if (!s) return ''
  if (typeof s === 'string') return s
  if (typeof s === 'object' && s.status) return String(s.status)
  return String(s)
}

const CREATOR_TYPE_OPTIONS = [
  { label: 'All', value: '' },
  { label: 'Nuon', value: 'nuon' },
  { label: 'User', value: 'user' },
]

const SORT_OPTIONS = [
  { label: 'Newest', value: 'newest' },
  { label: 'Oldest', value: 'oldest' },
]

export const InstallsList = () => {
  const [search, setSearch] = useState('')
  const [creatorType, setCreatorType] = useState('')
  const [sort, setSort] = useState('newest')
  const [showDeleted, setShowDeleted] = useState(false)
  const [page, setPage] = useState(1)

  const { data, isLoading, error } = useQuery({
    queryKey: ['installs', search, creatorType, sort, showDeleted, page],
    queryFn: () =>
      getInstalls({
        search,
        creator_type: creatorType || undefined,
        sort,
        show_deleted: showDeleted ? 'true' : undefined,
        page,
      }),
    refetchInterval: 20000,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load installs'} />

  const { installs = [], total_pages = 1 } = data || {}

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">Installs</h1>

      <div className="mt-4 flex flex-col gap-4 sm:flex-row sm:items-center sm:flex-wrap">
        <div className="w-full sm:w-64">
          <SearchInput value={search} onChange={(v) => { setSearch(v); setPage(1) }} placeholder="Search installs..." />
        </div>
        <div className="flex gap-2">
          {CREATOR_TYPE_OPTIONS.map((opt) => (
            <button
              key={opt.value}
              onClick={() => { setCreatorType(opt.value); setPage(1) }}
              className={`rounded-md px-3 py-1.5 text-sm font-medium ${
                creatorType === opt.value
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
          <thead className="bg-gray-50 dark:bg-gray-900">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Org</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">App</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Runner</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Sandbox</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Components</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-800 bg-white dark:bg-gray-900">
            {installs.map((install) => {
              const isDeleted = !!install.deleted_at
              const runnerStatus = getStatus(install.runner_status)
              const sandboxStatus = getStatus(install.sandbox_status)
              const componentStatus = getStatus(install.composite_component_status)
              return (
                <tr key={install.id} className={`hover:bg-gray-50 dark:hover:bg-gray-800 ${isDeleted ? 'opacity-50' : ''}`}>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <div className="flex items-center gap-2">
                      <Link to={`/installs/${install.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200 font-medium">
                        {install.name || truncateId(install.id)}
                      </Link>
                      {isDeleted && <Badge variant="status" status="error">Deleted</Badge>}
                    </div>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400 font-mono">{truncateId(install.id)}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Link to={`/orgs/${install.org_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200">
                      {install.org?.name || truncateId(install.org_id)}
                    </Link>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
                    {install.app?.name || truncateId(install.app_id)}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    {runnerStatus && <Badge variant="status" status={runnerStatus}>{runnerStatus}</Badge>}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    {sandboxStatus && <Badge variant="status" status={sandboxStatus}>{sandboxStatus}</Badge>}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    {componentStatus && <Badge variant="status" status={componentStatus}>{componentStatus}</Badge>}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDate(install.created_at)}</td>
                </tr>
              )
            })}
            {installs.length === 0 && (
              <tr>
                <td colSpan={8} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No installs found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={total_pages} onPageChange={setPage} />
    </div>
  )
}
