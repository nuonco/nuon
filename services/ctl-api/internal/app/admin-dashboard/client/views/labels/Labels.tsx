import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link } from 'react-router'
import { getLabels } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { SearchInput } from '@/components/common/SearchInput'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { truncateId } from '@/utils/format'

const ENTITY_TYPE_OPTIONS = [
  { label: 'All', value: '' },
  { label: 'Installs', value: 'install' },
  { label: 'Components', value: 'component' },
  { label: 'Actions', value: 'action' },
]

export const Labels = () => {
  const [search, setSearch] = useState('')
  const [entityType, setEntityType] = useState('')
  const [orgId, setOrgId] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading, error } = useQuery({
    queryKey: ['labels', search, entityType, orgId, page],
    queryFn: () => getLabels({
      search: search || undefined,
      entity_type: entityType || undefined,
      org_id: orgId || undefined,
      page,
    }),
    refetchInterval: 30000,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load labels'} />

  const { results = [], all_keys = [], orgs = [], total_pages = 1, total_count = 0 } = data || {}

  return (
    <div>
      <h1 className="page-heading">Labels</h1>
      <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">{total_count} results{all_keys.length > 0 && ` across ${all_keys.length} label keys`}</p>

      <div className="mt-4 flex flex-col gap-4 sm:flex-row sm:items-center sm:flex-wrap">
        <div className="w-full sm:w-64">
          <SearchInput
            value={search}
            onChange={(v) => { setSearch(v); setPage(1) }}
            placeholder="Search key:value..."
          />
        </div>

        <div className="flex gap-2">
          {ENTITY_TYPE_OPTIONS.map((opt) => (
            <button
              key={opt.value}
              onClick={() => { setEntityType(opt.value); setPage(1) }}
              className={`rounded-md px-3 py-1.5 text-sm font-medium ${
                entityType === opt.value
                  ? 'bg-primary-600 dark:bg-primary-500 text-white'
                  : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'
              }`}
            >
              {opt.label}
            </button>
          ))}
        </div>

        {orgs.length > 0 && (
          <select
            value={orgId}
            onChange={(e) => { setOrgId(e.target.value); setPage(1) }}
            className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 focus:ring-2 focus:ring-primary-600 dark:focus:ring-primary-500"
          >
            <option value="">All organizations</option>
            {orgs.map((org) => (
              <option key={org.id} value={org.id}>{org.name}</option>
            ))}
          </select>
        )}
      </div>

      {all_keys.length > 0 && (
        <div className="mt-3 flex flex-wrap gap-1">
          {all_keys.map((key) => (
            <button
              key={key}
              onClick={() => { setSearch(key); setPage(1) }}
              className="rounded-md bg-gray-100 dark:bg-gray-800 px-2 py-0.5 text-xs text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700"
            >
              {key}
            </button>
          ))}
        </div>
      )}

      <div className="mt-4 table-card">
        <table>
          <thead>
            <tr>
              <th>Entity Type</th>
              <th>Entity Name</th>
              <th>Entity ID</th>
              <th>Labels</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
            {results.map((result) => (
              <tr key={`${result.entity_type}-${result.entity_id}`}>
                <td>
                  <Badge>{result.entity_type}</Badge>
                </td>
                <td>
                  {result.detail_url ? (
                    <Link to={result.detail_url} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 font-medium">
                      {result.entity_name || truncateId(result.entity_id)}
                    </Link>
                  ) : (
                    <span className="text-sm text-gray-900 dark:text-gray-100">{result.entity_name || truncateId(result.entity_id)}</span>
                  )}
                </td>
                <td className="text-gray-500 dark:text-gray-400 font-mono text-xs">{truncateId(result.entity_id)}</td>
                <td>
                  <div className="flex flex-wrap gap-1">
                    {Object.entries(result.labels || {}).map(([key, value]) => (
                      <span key={key} className="inline-flex items-center gap-0.5 rounded-md bg-blue-50 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-800 px-2 py-0.5 text-xs font-mono">
                        <span className="text-blue-700 dark:text-blue-300">{key}</span>
                        <span className="text-blue-400">=</span>
                        <span className="text-blue-600 dark:text-blue-400">{String(value)}</span>
                      </span>
                    ))}
                  </div>
                </td>
              </tr>
            ))}
            {results.length === 0 && (
              <tr>
                <td colSpan={4} className="text-center text-gray-500 dark:text-gray-400 py-6">No labeled entities found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={total_pages} onPageChange={setPage} />
    </div>
  )
}
