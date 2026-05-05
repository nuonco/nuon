import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link } from 'react-router'
import { getOrgs } from '@/lib/admin-api'
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

export const OrgsList = () => {
  const [search, setSearch] = useState('')
  const [labelKey, setLabelKey] = useState('')
  const [labelValue, setLabelValue] = useState('')
  const [page, setPage] = useState(1)

  const labelFilter = labelKey && labelValue ? `${labelKey}:${labelValue}` : labelKey || undefined

  const { data, isLoading, error } = useQuery({
    queryKey: ['orgs', search, labelKey, labelValue, page],
    queryFn: () => getOrgs({ search, label: labelFilter, page }),
    refetchInterval: 20000,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load organizations'} />

  const orgs = data?.orgs || []
  const label_options = data?.label_options || []
  const total_pages = data?.total_pages || 1
  const selectedLabelOption = label_options.find((l) => l.key === labelKey)

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900">Organizations</h1>

      <div className="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:flex-wrap">
        <div className="w-full sm:w-64">
          <SearchInput value={search} onChange={(v) => { setSearch(v); setPage(1) }} placeholder="Search orgs..." />
        </div>
        <select
          value={labelKey}
          onChange={(e) => { setLabelKey(e.target.value); setLabelValue(''); setPage(1) }}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300"
        >
          <option value="">All labels</option>
          {label_options.map((l) => (
            <option key={l.key} value={l.key}>{l.key}</option>
          ))}
        </select>
        {labelKey && selectedLabelOption && selectedLabelOption.values.length > 0 && (
          <select
            value={labelValue}
            onChange={(e) => { setLabelValue(e.target.value); setPage(1) }}
            className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300"
          >
            <option value="">Any value</option>
            {selectedLabelOption.values.map((v) => (
              <option key={v} value={v}>{v}</option>
            ))}
          </select>
        )}
      </div>

      <div className="mt-4 overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Labels</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Apps</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Installs</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 bg-white">
            {orgs.map((org) => {
              const orgLabels = org.labels || {}
              return (
                <tr key={org.id} className="hover:bg-gray-50">
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Link to={`/orgs/${org.id}`} className="text-primary-600 hover:text-primary-800 font-medium">
                      {org.name}
                    </Link>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 font-mono">{truncateId(org.id)}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    {org.status && (
                      <Badge variant="status" status={getStatus(org.status)}>{getStatus(org.status)}</Badge>
                    )}
                  </td>
                  <td className="px-4 py-3 text-sm">
                    <div className="flex flex-wrap gap-1">
                      {Object.entries(orgLabels).map(([k, v]) => (
                        <span key={k} className="inline-flex items-center rounded bg-blue-50 border border-blue-200 px-1.5 py-0.5 text-[10px] font-mono">
                          <span className="text-blue-700">{k}</span>
                          <span className="text-blue-400">=</span>
                          <span className="text-blue-600">{String(v)}</span>
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{org.app_count ?? '-'}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{org.install_count ?? '-'}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(org.created_at)}</td>
                </tr>
              )
            })}
            {orgs.length === 0 && (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-sm text-gray-500">No organizations found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={total_pages} onPageChange={setPage} />
    </div>
  )
}
