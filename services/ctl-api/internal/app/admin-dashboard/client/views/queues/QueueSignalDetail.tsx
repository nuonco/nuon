import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { getQueueSignalDetail } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { JsonViewer } from '@/components/common/JsonViewer'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

export const QueueSignalDetail = () => {
  const { id: queueId, signalId } = useParams<{ id: string; signalId: string }>()

  const { data, isLoading, error } = useQuery({
    queryKey: ['queue-signal', queueId, signalId],
    queryFn: () => getQueueSignalDetail(queueId!, signalId!),
    enabled: !!queueId && !!signalId,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load signal'} />
  if (!data) return null

  const { signal, workflow_info } = data

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-gray-900">Signal Detail</h1>
        <p className="mt-1 text-sm text-gray-500 font-mono">{signal.id}</p>
      </div>

      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <dl className="grid grid-cols-2 gap-4 text-sm sm:grid-cols-3">
          <div>
            <dt className="text-gray-500">Type</dt>
            <dd className="mt-1 text-gray-900">{signal.type}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Status</dt>
            <dd className="mt-1"><Badge variant="status" status={String(signal.status)}>{String(signal.status)}</Badge></dd>
          </div>
          <div>
            <dt className="text-gray-500">Owner</dt>
            <dd className="mt-1 font-mono text-xs text-gray-900">{truncateId(signal.owner_id)}</dd>
            <dd className="text-xs text-gray-400">{signal.owner_type}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Queue</dt>
            <dd className="mt-1 font-mono text-xs text-gray-900">{truncateId(signal.queue_id)}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Created</dt>
            <dd className="mt-1 text-gray-900">{formatDate(signal.created_at)}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Updated</dt>
            <dd className="mt-1 text-gray-900">{formatDate(signal.updated_at)}</dd>
          </div>
        </dl>
      </div>

      {/* Signal Payload */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Signal Payload</h2>
        <div className="mt-2">
          <JsonViewer data={signal} collapsed />
        </div>
      </div>

      {/* Workflow Info */}
      {workflow_info && (
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h2 className="text-sm font-semibold text-gray-900">Workflow Info</h2>
          <div className="mt-2">
            <JsonViewer data={workflow_info} collapsed />
          </div>
        </div>
      )}
    </div>
  )
}
