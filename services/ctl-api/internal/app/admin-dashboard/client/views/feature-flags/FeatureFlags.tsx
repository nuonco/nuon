import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link } from 'react-router'
import { getFeatureFlags } from '@/lib/admin-api'
import { SearchInput } from '@/components/common/SearchInput'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import type { TFeatureFlag } from '@/types/admin.types'

type Sort = 'newest' | 'rollout' | 'drift' | 'name'
type DefaultFilter = '' | 'on' | 'off'

const SORT_OPTIONS: { label: string; value: Sort }[] = [
  { label: 'Newest', value: 'newest' },
  { label: 'Most orgs', value: 'rollout' },
  { label: 'Most drift', value: 'drift' },
  { label: 'Name', value: 'name' },
]

const DEFAULT_OPTIONS: { label: string; value: DefaultFilter }[] = [
  { label: 'All defaults', value: '' },
  { label: 'Default on', value: 'on' },
  { label: 'Default off', value: 'off' },
]

const sortFlags = (flags: TFeatureFlag[], sort: Sort) => {
  if (sort === 'newest') return flags
  const sorted = [...flags]
  if (sort === 'name') return sorted.sort((a, b) => a.name.localeCompare(b.name))
  if (sort === 'drift') return sorted.sort((a, b) => b.drift_count - a.drift_count)
  return sorted.sort((a, b) => b.enabled_count - a.enabled_count)
}

export const FeatureFlags = () => {
  const [search, setSearch] = useState('')
  const [sort, setSort] = useState<Sort>('newest')
  const [defaultFilter, setDefaultFilter] = useState<DefaultFilter>('')

  const { data, isLoading, error } = useQuery({
    queryKey: ['feature-flags'],
    queryFn: getFeatureFlags,
    refetchInterval: 60000,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load feature flags'} />

  const { flags = [], total_orgs = 0 } = data || {}

  const term = search.trim().toLowerCase()
  const filtered = flags.filter((f) => {
    if (defaultFilter === 'on' && !f.effective_default) return false
    if (defaultFilter === 'off' && f.effective_default) return false
    if (!term) return true
    return f.name.toLowerCase().includes(term) || f.description.toLowerCase().includes(term)
  })
  const visible = sortFlags(filtered, sort)

  const defaultOnCount = flags.filter((f) => f.effective_default).length
  const autoEnabledCount = flags.filter((f) => f.auto_enabled).length

  return (
    <div>
      <h1 className="page-heading">Feature flags</h1>
      <p className="page-subheading">
        {flags.length} flags across {total_orgs} orgs · {defaultOnCount} on by default for new orgs
        {autoEnabledCount > 0 && `, ${autoEnabledCount} of them forced on by this deployment's auto_enabled_features config`}
      </p>

      <div className="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:flex-wrap">
        <div className="w-full sm:w-72">
          <SearchInput value={search} onChange={setSearch} placeholder="Search flags..." />
        </div>

        <div className="flex gap-2">
          {DEFAULT_OPTIONS.map((opt) => (
            <button
              key={opt.value}
              onClick={() => setDefaultFilter(opt.value)}
              className={`rounded-md px-3 py-1.5 text-sm font-medium ${
                defaultFilter === opt.value
                  ? 'bg-primary-600 dark:bg-primary-500 text-white'
                  : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'
              }`}
            >
              {opt.label}
            </button>
          ))}
        </div>

        <div className="flex gap-2 sm:ml-auto">
          {SORT_OPTIONS.map((opt) => (
            <button
              key={opt.value}
              onClick={() => setSort(opt.value)}
              className={`rounded-md px-3 py-1.5 text-sm font-medium ${
                sort === opt.value
                  ? 'bg-primary-600 dark:bg-primary-500 text-white'
                  : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'
              }`}
            >
              {opt.label}
            </button>
          ))}
        </div>
      </div>

      <div className="mt-4 table-card">
        <table>
          <thead>
            <tr>
              <th>Flag</th>
              <th className="w-32">New orgs</th>
              <th className="w-64">Orgs enabled</th>
              <th className="w-40">Differs from default</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
            {visible.map((flag) => {
              const pct = total_orgs > 0 ? Math.round((flag.enabled_count / total_orgs) * 100) : 0
              const driftState = flag.effective_default ? 'disabled' : 'enabled'
              return (
                <tr key={flag.name}>
                  <td>
                    <Link
                      to={`/orgs?feature=${encodeURIComponent(flag.name)}&feature_state=enabled`}
                      title={`View the ${flag.enabled_count} orgs with ${flag.name} enabled`}
                      className="font-mono text-sm text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300"
                    >
                      {flag.name}
                    </Link>
                    {flag.description && (
                      <div className="mt-0.5 max-w-3xl text-xs text-gray-500 dark:text-gray-400">{flag.description}</div>
                    )}
                  </td>
                  <td>
                    <span
                      className={`inline-flex rounded px-1.5 py-0.5 text-[10px] font-mono ${
                        flag.effective_default
                          ? 'bg-green-50 dark:bg-green-900/30 border border-green-200 dark:border-green-800 text-green-700 dark:text-green-300'
                          : 'bg-gray-100 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400'
                      }`}
                    >
                      {flag.effective_default ? 'on' : 'off'}
                    </span>
                    {flag.auto_enabled && !flag.default && (
                      <div
                        className="mt-0.5 text-[10px] text-amber-600 dark:text-amber-400"
                        title="Off in the code defaults, but this deployment's auto_enabled_features config turns it on for newly created orgs. Existing orgs are unaffected."
                      >
                        auto_enabled_features
                      </div>
                    )}
                  </td>
                  <td>
                    <div className="flex items-center gap-2">
                      <div className="h-1.5 w-24 overflow-hidden rounded-full bg-gray-200 dark:bg-gray-800">
                        <div className="h-full rounded-full bg-primary-500" style={{ width: `${pct}%` }} />
                      </div>
                      <Link
                        to={`/orgs?feature=${encodeURIComponent(flag.name)}&feature_state=enabled`}
                        className="font-mono text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300"
                      >
                        {flag.enabled_count}/{total_orgs}
                      </Link>
                      <span className="text-xs text-gray-400 dark:text-gray-500">{pct}%</span>
                    </div>
                    {flag.unset_count > 0 && (
                      <div className="mt-0.5 text-[10px] text-gray-400 dark:text-gray-500">
                        {flag.effective_default
                          ? `incl. ${flag.unset_count} with no stored value, assumed on`
                          : `${flag.unset_count} with no stored value, assumed off`}
                      </div>
                    )}
                  </td>
                  <td>
                    {flag.drift_count === 0 ? (
                      <span className="text-xs text-gray-400 dark:text-gray-500">none</span>
                    ) : (
                      <Link
                        to={`/orgs?feature=${encodeURIComponent(flag.name)}&feature_state=${driftState}`}
                        className="text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300"
                      >
                        {flag.drift_count} {driftState}
                      </Link>
                    )}
                  </td>
                </tr>
              )
            })}
            {visible.length === 0 && (
              <tr>
                <td colSpan={4} className="text-center text-gray-500 dark:text-gray-400 py-6">No feature flags found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
