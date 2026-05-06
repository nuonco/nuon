import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router'
import { getQueueDetail, getQueueEmitters, getQueueInFlightSignals, restartQueue, forceRestartQueue, clearQueue } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { StatusHistory } from '@/components/common/StatusHistory'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, formatRelativeDate, truncateId } from '@/utils/format'

function getStatus(s: any): string {
  if (!s) return ''
  if (typeof s === 'string') return s
  if (typeof s === 'object' && s.status) return String(s.status)
  return String(s)
}

function queueStatusLabel(status: { Ready?: boolean; Paused?: boolean; Stopped?: boolean } | null): string {
  if (!status) return 'unknown'
  if (status.Stopped) return 'stopped'
  if (status.Paused) return 'paused'
  if (status.Ready) return 'ready'
  return 'unknown'
}

function queueStatusBadgeStatus(status: { Ready?: boolean; Paused?: boolean; Stopped?: boolean } | null): string {
  if (!status) return 'unknown'
  if (status.Stopped) return 'error'
  if (status.Paused) return 'warning'
  if (status.Ready) return 'healthy'
  return 'unknown'
}

export const QueueDetail = () => {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const [emittersPage, setEmittersPage] = useState(1)

  const { data, isLoading, error } = useQuery({
    queryKey: ['queue', id],
    queryFn: () => getQueueDetail(id!),
    enabled: !!id,
  })

  const { data: emittersData } = useQuery({
    queryKey: ['queue-emitters', id, emittersPage],
    queryFn: () => getQueueEmitters(id!, { page: emittersPage }),
    enabled: !!id,
    refetchInterval: 30000,
  })

  const { data: inFlightData } = useQuery({
    queryKey: ['queue-in-flight', id],
    queryFn: () => getQueueInFlightSignals(id!),
    enabled: !!id,
    refetchInterval: 5000,
  })

  const restartMutation = useMutation({
    mutationFn: () => restartQueue(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['queue', id] }),
  })

  const forceRestartMutation = useMutation({
    mutationFn: () => forceRestartQueue(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['queue', id] }),
  })

  const clearMutation = useMutation({
    mutationFn: () => clearQueue(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['queue', id] }),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load queue'} />
  if (!data) return null

  const { queue, status, signals, in_flight_signals, temporal_ui_url } = data

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <nav className="text-sm text-gray-500 dark:text-gray-400">
        <Link to="/queues" className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200">Queues</Link>
        <span className="mx-1">/</span>
        <span className="font-mono">{truncateId(queue.id)}</span>
      </nav>

      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <div className="flex items-center gap-3 flex-wrap">
            <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">{queue.name}</h1>
            {queue.status_v2?.status && (
              <Badge variant="status" status={getStatus(queue.status_v2)}>
                {getStatus(queue.status_v2)}
              </Badge>
            )}
            <Badge variant="status" status={queueStatusBadgeStatus(status)}>
              {queueStatusLabel(status)}
            </Badge>
          </div>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400 font-mono">{queue.id}</p>
          <div className="mt-1 flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
            <span>Owner:</span>
            <span className="font-mono text-xs">{truncateId(queue.owner_id)}</span>
            <Badge variant="default">{queue.owner_type}</Badge>
          </div>
        </div>
        {temporal_ui_url && queue.workflow?.id && (
          <a
            href={`${temporal_ui_url}/namespaces/components/workflows/${queue.workflow.id}`}
            target="_blank"
            rel="noopener noreferrer"
            className="rounded-md bg-gray-100 dark:bg-gray-800 px-3 py-1.5 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700"
          >
            View in Temporal
          </a>
        )}
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
          <div className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Max Depth</div>
          <div className="mt-1 text-2xl font-semibold text-gray-900 dark:text-gray-100">{queue.max_depth}</div>
        </div>
        <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
          <div className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Max In-Flight</div>
          <div className="mt-1 text-2xl font-semibold text-gray-900 dark:text-gray-100">{queue.max_in_flight}</div>
        </div>
        <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
          <div className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Queue Depth</div>
          <div className="mt-1 text-2xl font-semibold text-gray-900 dark:text-gray-100">{status?.QueueDepthCount ?? '-'}</div>
        </div>
        <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
          <div className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">In-Flight</div>
          <div className="mt-1 text-2xl font-semibold text-gray-900 dark:text-gray-100">{status?.InFlightCount ?? '-'}</div>
        </div>
      </div>

      {/* Status Timestamps */}
      {queue.metadata && (
        <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Status Timestamps</h2>
          <div className="mt-2 grid grid-cols-2 gap-x-8 gap-y-2 sm:grid-cols-3 lg:grid-cols-5">
            {[
              { label: 'Ready At', key: 'ready_at' },
              { label: 'Restarted At', key: 'restarted_at' },
              { label: 'Stopped At', key: 'stopped_at' },
              { label: 'Idled At', key: 'idled_at' },
              { label: 'Finished At', key: 'finished_at' },
            ].map(({ label, key }) => (
              <div key={key}>
                <div className="text-xs text-gray-500 dark:text-gray-400">{label}</div>
                <div className="text-sm text-gray-900 dark:text-gray-100 font-mono">
                  {queue.metadata?.[key] ? formatRelativeDate(queue.metadata[key]) : '-'}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="flex gap-2">
        <button
          onClick={() => restartMutation.mutate()}
          disabled={restartMutation.isPending}
          className="rounded-md bg-yellow-600 dark:bg-yellow-500 px-3 py-1.5 text-sm font-medium text-white hover:bg-yellow-700 dark:hover:bg-yellow-600 disabled:opacity-50"
        >
          {restartMutation.isPending ? 'Restarting...' : 'Restart Queue'}
        </button>
        <button
          onClick={() => {
            if (confirm('Are you sure you want to FORCE restart this queue? This skips waiting for active workers to finish.')) {
              forceRestartMutation.mutate()
            }
          }}
          disabled={forceRestartMutation.isPending}
          className="rounded-md bg-red-600 dark:bg-red-500 px-3 py-1.5 text-sm font-medium text-white hover:bg-red-700 dark:hover:bg-red-600 disabled:opacity-50"
        >
          {forceRestartMutation.isPending ? 'Force restarting...' : 'Force Restart'}
        </button>
        <button
          onClick={() => {
            if (confirm('Are you sure you want to cancel all in-flight signals in this queue?')) {
              clearMutation.mutate()
            }
          }}
          disabled={clearMutation.isPending}
          className="rounded-md bg-red-600 dark:bg-red-500 px-3 py-1.5 text-sm font-medium text-white hover:bg-red-700 dark:hover:bg-red-600 disabled:opacity-50"
        >
          {clearMutation.isPending ? 'Clearing...' : 'Clear Queue'}
        </button>
      </div>

      {/* Queue Status (StatusV2) */}
      {queue.status_v2?.status && (
        <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Queue Status</h2>
          <p className="mt-0.5 text-xs text-gray-500 dark:text-gray-400">Current status and history</p>
          <div className="mt-3">
            <StatusHistory status={queue.status_v2} defaultExpanded maxCollapsed={5} />
          </div>
        </div>
      )}

      {/* Emitters */}
      <div className="table-card rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Emitters</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
            <thead className="">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Mode</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Schedule</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Signal type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Emit count</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Last emitted</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Owner</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {(queue.emitters || emittersData?.emitters || []).map((emitter: any) => (
                <tr key={emitter.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Link to={`/queues/${id}/emitters/${emitter.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200 font-mono">
                      {truncateId(emitter.id)}
                    </Link>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{emitter.name}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{emitter.mode || '-'}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400 font-mono">{emitter.cron_schedule || '-'}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{emitter.signal_type || '-'}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    {emitter.status?.status ? <Badge variant="status" status={getStatus(emitter.status)}>{getStatus(emitter.status)}</Badge> : <span className="text-gray-400 dark:text-gray-500">-</span>}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm font-mono">{emitter.emit_count ?? 0}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{emitter.last_emitted_at ? formatRelativeDate(emitter.last_emitted_at) : '-'}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
                    <span className="font-mono text-xs">{truncateId(emitter.owner_id)}</span>
                    <span className="ml-1 text-xs text-gray-400 dark:text-gray-500">({emitter.owner_type})</span>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDate(emitter.created_at)}</td>
                </tr>
              ))}
              {(!(queue.emitters || emittersData?.emitters) || (queue.emitters || emittersData?.emitters || []).length === 0) && (
                <tr>
                  <td colSpan={10} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No emitters</td>
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
      <div className="table-card rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Recent Signals</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
            <thead className="">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Owner Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Owner ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {(signals || []).map((signal: any) => (
                <tr key={signal.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Link to={`/queues/${id}/signals/${signal.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200 font-mono">
                      {truncateId(signal.id)}
                    </Link>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{signal.type}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Badge variant="status" status={getStatus(signal.status)}>{getStatus(signal.status)}</Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Badge variant="default">{signal.owner_type}</Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm font-mono text-gray-500 dark:text-gray-400">
                    {truncateId(signal.owner_id)}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDate(signal.created_at)}</td>
                </tr>
              ))}
              {(!signals || signals.length === 0) && (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No signals</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* In-Flight Signals */}
      <div className="table-card rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">
          In-Flight Signals
          {(in_flight_signals || inFlightData?.signals || []).length > 0 && (
            <span className="ml-2 text-xs font-normal text-gray-500 dark:text-gray-400">
              ({(in_flight_signals || inFlightData?.signals || []).length})
            </span>
          )}
        </h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
            <thead className="">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Owner Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Owner ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Updated</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Duration</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {(in_flight_signals || inFlightData?.signals || []).map((signal: any) => {
                const createdMs = signal.created_at ? new Date(signal.created_at).getTime() : 0
                const nowMs = Date.now()
                const durationMs = createdMs ? nowMs - createdMs : 0
                const durationStr = durationMs > 0
                  ? durationMs < 60000
                    ? `${Math.round(durationMs / 1000)}s`
                    : durationMs < 3600000
                      ? `${Math.floor(durationMs / 60000)}m ${Math.round((durationMs % 60000) / 1000)}s`
                      : `${Math.floor(durationMs / 3600000)}h ${Math.floor((durationMs % 3600000) / 60000)}m`
                  : '-'
                return (
                  <tr key={signal.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                    <td className="whitespace-nowrap px-4 py-3 text-sm">
                      <Link to={`/queues/${id}/signals/${signal.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200 font-mono">
                        {truncateId(signal.id)}
                      </Link>
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{signal.type}</td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm">
                      <Badge variant="status" status={getStatus(signal.status)}>{getStatus(signal.status)}</Badge>
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm">
                      <Badge variant="default">{signal.owner_type}</Badge>
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm font-mono text-gray-500 dark:text-gray-400">
                      {truncateId(signal.owner_id)}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
                      {formatRelativeDate(signal.updated_at)}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400 font-mono">
                      {durationStr}
                    </td>
                  </tr>
                )
              })}
              {(!(in_flight_signals || inFlightData?.signals) || (in_flight_signals || inFlightData?.signals || []).length === 0) && (
                <tr>
                  <td colSpan={7} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No in-flight signals</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
