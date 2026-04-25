import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link } from 'react-router'
import { getSignalCatalog } from '@/lib/admin-api'
import { SearchInput } from '@/components/common/SearchInput'
import { Badge } from '@/components/common/Badge'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'

export const SignalCatalog = () => {
  const [search, setSearch] = useState('')

  const { data, isLoading, error } = useQuery({
    queryKey: ['signal-catalog', search],
    queryFn: () => getSignalCatalog(),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load signal catalog'} />

  const grouped = data?.grouped || {}
  const namespaces = data?.namespaces || []

  const lowerSearch = search.toLowerCase()
  const filteredNamespaces = search
    ? namespaces.filter((ns) => {
        const signals = grouped[ns] || []
        return ns.toLowerCase().includes(lowerSearch) ||
          signals.some((s: any) =>
            String(s.Type || '').toLowerCase().includes(lowerSearch) ||
            String(s.Operation || '').toLowerCase().includes(lowerSearch)
          )
      })
    : namespaces

  const totalSignals = Object.values(grouped).reduce((sum, arr) => sum + (arr?.length || 0), 0)

  return (
    <div>
      <h1 className="page-heading">Signal catalog</h1>
      <p className="page-subheading">{totalSignals} signal types across {namespaces.length} namespaces</p>

      <div className="mt-4 w-full sm:w-64">
        <SearchInput value={search} onChange={setSearch} placeholder="Filter signals..." />
      </div>

      <div className="mt-4 space-y-6">
        {filteredNamespaces.map((ns) => {
          const signals = grouped[ns] || []
          const filteredSignals = search
            ? signals.filter((s: any) =>
                String(s.Type || '').toLowerCase().includes(lowerSearch) ||
                String(s.Operation || '').toLowerCase().includes(lowerSearch) ||
                ns.toLowerCase().includes(lowerSearch)
              )
            : signals

          if (filteredSignals.length === 0) return null

          return (
            <div key={ns}>
              <h2 className="text-sm font-semibold text-gray-700 mb-2">{ns} <span className="text-gray-400 font-normal">({filteredSignals.length})</span></h2>
              <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
                {filteredSignals.map((info: any) => (
                  <Link
                    key={String(info.Type)}
                    to={`/signal-catalog/${encodeURIComponent(String(info.Type))}`}
                    className="group rounded-lg border border-gray-200 bg-white p-3 transition-all duration-100 hover:border-primary-200 hover:shadow-sm"
                  >
                    <p className="text-xs font-medium text-gray-900 group-hover:text-primary-700 break-all font-mono">{String(info.Type)}</p>
                    {info.Operation && (
                      <p className="mt-1 text-[11px] text-gray-500">op: {info.Operation}</p>
                    )}
                  </Link>
                ))}
              </div>
            </div>
          )
        })}
        {filteredNamespaces.length === 0 && (
          <p className="text-sm text-gray-500">No signal types found</p>
        )}
      </div>
    </div>
  )
}
