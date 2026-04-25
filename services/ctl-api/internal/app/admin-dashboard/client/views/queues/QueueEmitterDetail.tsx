import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { getQueueEmitterDetail } from '@/lib/admin-api'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

export const QueueEmitterDetail = () => {
  const { id: queueId, emitterId } = useParams<{ id: string; emitterId: string }>()

  const { data, isLoading, error } = useQuery({
    queryKey: ['queue-emitter', queueId, emitterId],
    queryFn: () => getQueueEmitterDetail(queueId!, emitterId!),
    enabled: !!queueId && !!emitterId,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load emitter'} />
  if (!data) return null

  const { emitter } = data

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-gray-900">Emitter Detail</h1>
        <p className="mt-1 text-sm text-gray-500 font-mono">{emitter.id}</p>
      </div>

      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <dl className="grid grid-cols-2 gap-4 text-sm sm:grid-cols-3">
          <div>
            <dt className="text-gray-500">Name</dt>
            <dd className="mt-1 text-gray-900">{emitter.name}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Queue</dt>
            <dd className="mt-1 font-mono text-xs text-gray-900">{truncateId(emitter.queue_id)}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Owner</dt>
            <dd className="mt-1 font-mono text-xs text-gray-900">{truncateId(emitter.owner_id)}</dd>
            <dd className="text-xs text-gray-400">{emitter.owner_type}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Created</dt>
            <dd className="mt-1 text-gray-900">{formatDate(emitter.created_at)}</dd>
          </div>
        </dl>
      </div>
    </div>
  )
}
