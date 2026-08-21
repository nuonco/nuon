import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useSearchParams } from 'react-router'
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
  const [searchParams, setSearchParams] = useSearchParams()

  const feature = searchParams.get('feature') || ''
  const featureState = searchParams.get('feature_state') || 'enabled'

  const setFeatureFilter = (nextFeature: string, nextState: string) => {
    const next = new URLSearchParams(searchParams)
    if (nextFeature) {
      next.set('feature', nextFeature)
      next.set('feature_state', nextState)
    } else {
      next.delete('feature')
      next.delete('feature_state')
    }
    setSearchParams(next)
    setPage(1)
  }

  // Arriving from the feature flags page means the flag is what you care about,
  // so keep that context when drilling into an org.
  const orgDetailPath = (orgID: string) =>
    feature
      ? `/orgs/${orgID}?tab=features&flag=${encodeURIComponent(feature)}`
      : `/orgs/${orgID}`

  const labelFilter = labelKey && labelValue ? `${labelKey}:${labelValue}` : labelKey || undefined

  const { data, isLoading, error } = useQuery({
    queryKey: ['orgs', search, labelKey, labelValue, feature, featureState, page],
    queryFn: () => getOrgs({
      search,
      label: labelFilter,
      feature: feature || undefined,
      feature_state: feature ? featureState : undefined,
      page,
    }),
    refetchInterval: 20000,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load organizations'} />

  const orgs = data?.orgs || []
  const label_options = data?.label_options || []
  const feature_options = data?.feature_options || []
  const feature_counts = data?.feature_counts || {}
  const total_pages = data?.total_pages || 1
  const selectedLabelOption = label_options.find((l) => l.key === labelKey)
  const selectedFeature = feature_options.find((f) => f.name === feature)

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">Organizations</h1>

      <div className="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:flex-wrap">
        <div className="w-full sm:w-64">
          <SearchInput value={search} onChange={(v) => { setSearch(v); setPage(1) }} placeholder="Search orgs..." />
        </div>
<select
  value={labelKey}
  onChange={(e) => { setLabelKey(e.target.value); setLabelValue(''); setPage(1) }}
  className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 dark:bg-gray-900"
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
    className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 dark:bg-gray-900"
  >
    <option value="">Any value</option>
    {selectedLabelOption.values.map((v) => (
      <option key={v} value={v}>{v}</option>
    ))}
  </select>
)}
<select
  value={feature}
  onChange={(e) => setFeatureFilter(e.target.value, featureState)}
  title={selectedFeature?.description}
  className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 dark:bg-gray-900"
>
  <option value="">All feature flags</option>
  {feature_options.map((f) => (
    <option key={f.name} value={f.name}>{f.name}</option>
  ))}
</select>
{feature && (
  <select
    value={featureState}
    onChange={(e) => setFeatureFilter(feature, e.target.value)}
    className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 dark:bg-gray-900"
  >
    <option value="enabled">enabled</option>
    <option value="disabled">disabled</option>
  </select>
)}
</div>

      {feature && (
        <div className="mt-2 max-w-4xl text-xs text-gray-500 dark:text-gray-400">
          <span className="font-mono text-gray-600 dark:text-gray-300">{feature}</span> {featureState}
          {selectedFeature?.description && <>: {selectedFeature.description}</>}
          <div className="mt-1 flex gap-3">
            <button onClick={() => setFeatureFilter('', 'enabled')} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
              Clear filter
            </button>
            <Link to="/feature-flags" className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
              All feature flags
            </Link>
          </div>
        </div>
      )}

      <div className="mt-4 overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
          <thead className="">
            <tr>
<th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
<th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
<th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
<th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Labels</th>
<th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Flags</th>
<th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Apps</th>
<th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Installs</th>
<th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
            {orgs.map((org) => {
              const orgLabels = org.labels || {}
              const enabledFlagCount = feature_counts[org.id] ?? 0
              return (
                <tr key={org.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Link to={orgDetailPath(org.id)} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200 font-medium">
                      {org.name}
                    </Link>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400 font-mono">{truncateId(org.id)}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    {org.status && (
                      <Badge variant="status" status={getStatus(org.status)}>{getStatus(org.status)}</Badge>
                    )}
                  </td>
                  <td className="px-4 py-3 text-sm">
                    <div className="flex flex-wrap gap-1">
                      {Object.entries(orgLabels).map(([k, v]) => (
                        <span key={k} className="inline-flex items-center rounded bg-blue-50 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-800 px-1.5 py-0.5 text-[10px] font-mono">
                          <span className="text-blue-700 dark:text-blue-300">{k}</span>
                          <span className="text-blue-400">=</span>
                          <span className="text-blue-600 dark:text-blue-400">{String(v)}</span>
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Link
                      to={feature ? orgDetailPath(org.id) : `/orgs/${org.id}?tab=features`}
                      className="font-mono text-xs text-gray-500 dark:text-gray-400 hover:text-primary-600 dark:hover:text-primary-400"
                    >
                      {enabledFlagCount}/{feature_options.length}
                    </Link>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{org.app_count ?? '-'}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{org.install_count ?? '-'}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDate(org.created_at)}</td>
                </tr>
              )
            })}
            {orgs.length === 0 && (
              <tr>
                <td colSpan={8} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No organizations found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={total_pages} onPageChange={setPage} />
    </div>
  )
}
