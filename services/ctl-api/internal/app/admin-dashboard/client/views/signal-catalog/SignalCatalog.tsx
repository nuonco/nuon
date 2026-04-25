import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router'
import { getSignalCatalog } from '@/lib/admin-api'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'

export const SignalCatalog = () => {
  const { data, isLoading, error } = useQuery({
    queryKey: ['signal-catalog'],
    queryFn: () => getSignalCatalog(),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load signal catalog'} />

  const signalTypes = data?.signal_types || []

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900">Signal Catalog</h1>
      <p className="mt-1 text-sm text-gray-500">{signalTypes.length} signal types</p>

      <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {signalTypes.map((type) => (
          <Link
            key={type}
            to={`/signal-catalog/${encodeURIComponent(type)}`}
            className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm transition-shadow hover:shadow-md"
          >
            <h3 className="text-sm font-medium text-gray-900 break-all">{type}</h3>
          </Link>
        ))}
        {signalTypes.length === 0 && (
          <p className="text-sm text-gray-500 col-span-full">No signal types found</p>
        )}
      </div>
    </div>
  )
}
