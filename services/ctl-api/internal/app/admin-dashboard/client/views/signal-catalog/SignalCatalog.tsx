import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link } from 'react-router'
import { getSignalCatalog } from '@/lib/admin-api'
import { SearchInput } from '@/components/common/SearchInput'
import { Badge } from '@/components/common/Badge'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'

type SignalInfo = {
  Type: string
  Namespace?: string
  AutoRetry?: boolean
  MaxRetries?: number
  HasMaxAutoRetries?: boolean
  HasCloneSteps?: boolean
  HasNoOpCheck?: boolean
  HasPolicyEval?: boolean
  HasSkipCleanup?: boolean
  HasOnApprove?: boolean
  HasOnRetry?: boolean
  HasOnSkip?: boolean
  HasOnDeny?: boolean
  SkipGroup?: boolean
  HasFetchSteps?: boolean
  HasQueue?: boolean
  Queue?: string
  IsParallelizable?: boolean
  HasStepContext?: boolean
  HasLifecycle?: boolean
  Operation?: string
}

function CapabilityBadges({ info }: { info: SignalInfo }) {
  const caps: string[] = []
  if (info.HasCloneSteps) caps.push('clone-steps')
  if (info.HasNoOpCheck) caps.push('noop-check')
  if (info.HasPolicyEval) caps.push('policy-eval')
  if (info.HasSkipCleanup) caps.push('skip-cleanup')
  if (info.HasOnApprove) caps.push('on-approve')
  if (info.HasOnRetry) caps.push('on-retry')
  if (info.HasOnSkip) caps.push('on-skip')
  if (info.HasOnDeny) caps.push('on-deny')
  if (info.HasFetchSteps) caps.push('fetch-steps')
  if (info.IsParallelizable) caps.push('parallelizable')
  if (info.HasStepContext) caps.push('step-context')
  if (info.HasLifecycle) caps.push('lifecycle')

  return (
    <div className="flex flex-wrap gap-1">
      {caps.map((c) => (
        <Badge key={c}>{c}</Badge>
      ))}
      {info.HasQueue && info.Queue && <Badge>queue: {info.Queue}</Badge>}
      {caps.length === 0 && !info.HasQueue && <span className="text-xs text-gray-400 dark:text-gray-500">—</span>}
    </div>
  )
}

export const SignalCatalog = () => {
  const [search, setSearch] = useState('')

  const { data, isLoading, error } = useQuery({
    queryKey: ['signal-catalog', search],
    queryFn: () => getSignalCatalog(),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load signal catalog'} />

  const grouped = (data?.grouped || {}) as Record<string, SignalInfo[]>
  const namespaces = data?.namespaces || []

  const lowerSearch = search.toLowerCase()
  const matches = (info: SignalInfo, ns: string) =>
    !search ||
    ns.toLowerCase().includes(lowerSearch) ||
    String(info.Type || '').toLowerCase().includes(lowerSearch) ||
    String(info.Operation || '').toLowerCase().includes(lowerSearch)

  const visibleNamespaces = namespaces
    .map((ns) => ({ ns, signals: (grouped[ns] || []).filter((info) => matches(info, ns)) }))
    .filter(({ signals }) => signals.length > 0)

  const totalSignals = Object.values(grouped).reduce((sum, arr) => sum + (arr?.length || 0), 0)

  return (
    <div>
      <h1 className="page-heading">Signal catalog</h1>
      <p className="page-subheading">{totalSignals} signal types across {namespaces.length} namespaces</p>

      <div className="mt-4 w-full sm:w-64">
        <SearchInput value={search} onChange={setSearch} placeholder="Filter signals..." />
      </div>

      <div className="mt-4 space-y-8">
        {visibleNamespaces.map(({ ns, signals }) => (
          <div key={ns}>
            <h2 className="border-b border-gray-200 dark:border-gray-800 pb-2 mb-3 text-sm font-semibold uppercase tracking-wider text-primary-600 dark:text-primary-400">
              {ns}
              <span className="ml-2 font-normal text-gray-400 dark:text-gray-500">({signals.length})</span>
            </h2>
            <div className="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-800">
              <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
                <thead className="">
                  <tr>
                    <th className="px-4 py-2 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">Type</th>
                    <th className="px-4 py-2 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">Operation</th>
                    <th className="px-4 py-2 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">Auto-retry</th>
                    <th className="px-4 py-2 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">Max retries</th>
                    <th className="px-4 py-2 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">Capabilities</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
                  {signals.map((info) => (
                    <tr key={String(info.Type)} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                      <td className="whitespace-nowrap px-4 py-2 text-xs font-mono">
                        <Link
                          to={`/signal-catalog/${encodeURIComponent(String(info.Type))}`}
                          className="text-primary-600 dark:text-primary-400 hover:underline"
                        >
                          {String(info.Type)}
                        </Link>
                      </td>
                      <td className="whitespace-nowrap px-4 py-2 text-xs font-mono text-gray-700 dark:text-gray-300">
                        {info.Operation || <span className="text-gray-400 dark:text-gray-500">—</span>}
                      </td>
                      <td className="whitespace-nowrap px-4 py-2 text-xs">
                        {info.AutoRetry ? (
                          <Badge variant="status" status="online">Yes</Badge>
                        ) : (
                          <span className="text-gray-400 dark:text-gray-500">No</span>
                        )}
                      </td>
                      <td className="whitespace-nowrap px-4 py-2 text-xs font-mono">{info.MaxRetries ?? 0}</td>
                      <td className="px-4 py-2 text-xs">
                        <CapabilityBadges info={info} />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        ))}
        {visibleNamespaces.length === 0 && (
          <p className="text-sm text-gray-500 dark:text-gray-400">No signal types found</p>
        )}
      </div>
    </div>
  )
}
