import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router'
import { getQueueDetail, getQueueEmitters, getQueueSignals, getQueueInFlightSignals, restartQueue, clearQueue } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

export const QueueDetail = () => {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const [emittersPage, setEmittersPage] = useState(1)
  const [signalsPage, setSignalsPage] = useState(1)

  const { data, isLoading, error } = useQuery({
    queryKey: ['queue', id],
    queryFn: () => getQueueDetail(id!),
    enabled: !!id,
  })

  const { data: emittersData } = useQuery({
    queryKey: ['queue-emitters', id, emittersPage],
    queryFn: () => getQueueEmitters(id!, { page: emittersPage }),
    enabled: !!id,
  })

  const { data: signalsData } = useQuery({
    queryKey: ['queue-signals', id, signalsPage],
    queryFn: () => getQueueSignals(id!, { page: signalsPage }),
    enabled: !!id,
  })

  const { data: inFlightData } = useQuery({
    queryKey: ['queue-in-flight', id],
    queryFn: () => getQueueInFlightSignals(id!),
    enabled: !!id,
  })

  const restartMutation = useMutation({
    mutationFn: () => restartQueue(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['queue', id] }),
  })

  const clearMutation = useMutation({
    mutationFn: () => clearQueue(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['queue', id] }),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load queue'} />
  if (!data) return null

  const { queue } = data

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-gray-900">{queue.name}</h1>
        <p className="mt-1 text-sm text-gray-500 font-mono">{queue.id}</p>
        <p className="mt-1 text-sm text-gray-500">
          Owner: <span className="font-mono">{truncateId(queue.owner_id)}</span> ({queue.owner_type})
        </p>
      </div>

      {/* Actions */}
      <div className="flex gap-2">
        <button
          onClick={() => restartMutation.mutate()}
          disabled={restartMutation.isPending}
          className="rounded-md bg-yellow-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-yellow-700 disabled:opacity-50"
        >
          {restartMutation.isPending ? 'Restarting...' : 'Restart Queue'}
        </button>
        <button
          onClick={() => clearMutation.mutate()}
          disabled={clearMutation.isPending}
          className="rounded-md bg-red-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
        >
          {clearMutation.isPending ? 'Clearing...' : 'Clear Queue'}
        </button>
      </div>

      {/* Emitters */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Emitters</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Owner</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {(emittersData?.emitters || []).map((emitter) => (
                <tr key={emitter.id} className="hover:bg-gray-50">
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Link to={`/queues/${id}/emitters/${emitter.id}`} className="text-primary-600 hover:text-primary-800 font-mono">
                      {truncateId(emitter.id)}
                    </Link>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{emitter.name}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">
                    <span className="font-mono text-xs">{truncateId(emitter.owner_id)}</span>
                    <span className="ml-1 text-xs text-gray-400">({emitter.owner_type})</span>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(emitter.created_at)}</td>
                </tr>
              ))}
              {(!emittersData?.emitters || emittersData.emitters.length === 0) && (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-sm text-gray-500">No emitters</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        {emittersData && (
          <Pagination page={emittersPage} totalPages={emittersData.total_pages} onPageChange={setEmittersPage} />
        )}
      </div>

      {/* Recent Signals */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Recent Signals</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {(signalsData?.signals || []).map((signal) => (
                <tr key={signal.id} className="hover:bg-gray-50">
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Link to={`/queues/${id}/signals/${signal.id}`} className="text-primary-600 hover:text-primary-800 font-mono">
                      {truncateId(signal.id)}
                    </Link>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{signal.type}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Badge variant="status" status={String(signal.status)}>{String(signal.status)}</Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(signal.created_at)}</td>
                </tr>
              ))}
              {(!signalsData?.signals || signalsData.signals.length === 0) && (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-sm text-gray-500">No signals</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        {signalsData && (
          <Pagination page={signalsPage} totalPages={signalsData.total_pages} onPageChange={setSignalsPage} />
        )}
      </div>

      {/* In-Flight Signals */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">In-Flight Signals</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {(inFlightData?.signals || []).map((signal) => (
                <tr key={signal.id}>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Link to={`/queues/${id}/signals/${signal.id}`} className="text-primary-600 hover:text-primary-800 font-mono">
                      {truncateId(signal.id)}
                    </Link>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{signal.type}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Badge variant="status" status={String(signal.status)}>{String(signal.status)}</Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(signal.created_at)}</td>
                </tr>
              ))}
              {(!inFlightData?.signals || inFlightData.signals.length === 0) && (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-sm text-gray-500">No in-flight signals</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
