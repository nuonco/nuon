import { useQuery } from '@tanstack/react-query'
import { getLabels } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'

export const Labels = () => {
  const { data, isLoading, error } = useQuery({
    queryKey: ['labels'],
    queryFn: () => getLabels(),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load labels'} />

  const labels = data?.labels || {}
  const keys = Object.keys(labels).sort()

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900">Labels</h1>
      <p className="mt-1 text-sm text-gray-500">{keys.length} label keys</p>

      <div className="mt-4 space-y-4">
        {keys.map((key) => (
          <div key={key} className="rounded-lg border border-gray-200 bg-white p-4">
            <h2 className="text-sm font-semibold text-gray-900">{key}</h2>
            <div className="mt-2 flex flex-wrap gap-2">
              {labels[key].map((value) => (
                <Badge key={value}>{value}</Badge>
              ))}
            </div>
          </div>
        ))}
        {keys.length === 0 && (
          <p className="text-sm text-gray-500">No labels found</p>
        )}
      </div>
    </div>
  )
}
