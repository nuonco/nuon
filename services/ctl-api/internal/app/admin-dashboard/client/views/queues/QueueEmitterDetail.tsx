import { useQuery } from '@tanstack/react-query'
import { Link, useParams } from 'react-router'
import { getQueueEmitterDetail } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { JsonViewer } from '@/components/common/JsonViewer'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

function getStatus(s: any): string {
  if (!s) return ''
  if (typeof s === 'string') return s
  if (typeof s === 'object' && s.status) return String(s.status)
  return String(s)
}

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

  const { emitter, queue, signals = [], temporal_ui_url: temporalUIUrl } = data
  const status = getStatus(emitter?.status)
  const isFireOnce = emitter?.mode === 'fire_once' || emitter?.mode === 'scheduled'

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <div className="flex gap-2 text-xs text-gray-500">
        <Link to="/queues" className="text-primary-600 hover:text-primary-700">Queues</Link>
        <span>&rarr;</span>
        <Link to={`/queues/${queue?.id}`} className="text-primary-600 hover:text-primary-700">{truncateId(queue?.id)}</Link>
        <span>&rarr;</span>
        <span>Emitter</span>
      </div>

      {/* Header */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <div className="flex flex-wrap items-center gap-2 mb-2">
          <h1 className="text-lg font-semibold">{emitter.name || 'Emitter'}</h1>
          <Badge>{emitter.mode}</Badge>
          <Badge variant="status" status={status}>{status || 'unknown'}</Badge>
          {temporalUIUrl && emitter.workflow?.id && emitter.workflow?.namespace && (
            <a
              href={`${temporalUIUrl}/namespaces/${emitter.workflow.namespace}/workflows/${emitter.workflow.id}`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-xs text-primary-600 hover:text-primary-700"
            >
              View in Temporal &rarr;
            </a>
          )}
        </div>
        <div className="space-y-1 text-xs">
          <div><span className="text-gray-500 uppercase">Emitter ID:</span> <span className="font-mono select-all">{emitter.id}</span></div>
          <div><span className="text-gray-500 uppercase">Queue ID:</span> <Link to={`/queues/${emitter.queue_id}`} className="font-mono text-primary-600 hover:text-primary-700">{emitter.queue_id}</Link></div>
        </div>
      </div>

      {/* Configuration + Runtime state */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {/* Configuration */}
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h2 className="text-sm font-semibold text-gray-900 mb-2">Configuration</h2>
          <div className="space-y-2 text-xs">
            <InfoRow label="Signal type" value={emitter.signal_type} />
            <InfoRow label="Mode" value={emitter.mode} />
            {emitter.mode === 'cron' && (
              <InfoRow label="Cron schedule" value={emitter.cron_schedule || '-'} />
            )}
            {isFireOnce && (
              <InfoRow label="Scheduled at" value={formatDate(emitter.scheduled_at)} />
            )}
            {emitter.description && (
              <InfoRow label="Description" value={emitter.description} />
            )}
            <InfoRow label="Owner" value={`${emitter.owner_type} / ${truncateId(emitter.owner_id)}`} />
            <InfoRow label="Created" value={formatDate(emitter.created_at)} />
          </div>
        </div>

        {/* Runtime state */}
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h2 className="text-sm font-semibold text-gray-900 mb-2">Runtime State</h2>
          <div className="space-y-2 text-xs">
            <div className="flex items-start gap-3">
              <span className="text-gray-500 uppercase w-28 shrink-0">Status</span>
              <Badge variant="status" status={status}>{status || 'unknown'}</Badge>
            </div>
            {emitter.status?.status_human_description && (
              <InfoRow label="Description" value={emitter.status.status_human_description} />
            )}
            <InfoRow label="Emit count" value={String(emitter.emit_count ?? 0)} />
            <InfoRow label="Last emitted" value={formatDate(emitter.last_emitted_at)} />
            <InfoRow label="Next emit" value={formatDate(emitter.next_emit_at)} />
            {isFireOnce && (
              <div className="flex items-start gap-3">
                <span className="text-gray-500 uppercase w-28 shrink-0">Fired</span>
                <Badge variant="status" status={emitter.fired ? 'yes' : 'no'}>{emitter.fired ? 'Yes' : 'No'}</Badge>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Signal template */}
      {emitter.signal_template && (
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h2 className="text-sm font-semibold text-gray-900 mb-2">Signal Template</h2>
          <JsonViewer data={emitter.signal_template} />
        </div>
      )}

      {/* Emitted signals */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Emitted Signals ({signals.length})</h2>
        {signals.length > 0 ? (
          <div className="mt-2 table-card">
            <table>
              <thead>
                <tr>
                  <th>Signal ID</th>
                  <th>Type</th>
                  <th>Status</th>
                  <th>Owner</th>
                  <th>Created</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {signals.map((sig: any) => (
                  <tr key={sig.id}>
                    <td className="font-mono text-xs">
                      <Link
                        to={`/queues/${queueId}/signals/${sig.id}`}
                        className="text-primary-600 hover:text-primary-700"
                      >
                        {truncateId(sig.id)}
                      </Link>
                    </td>
                    <td className="text-xs"><Badge>{sig.type}</Badge></td>
                    <td><Badge variant="status" status={getStatus(sig.status)}>{getStatus(sig.status)}</Badge></td>
                    <td className="text-xs text-gray-500">
                      <span className="font-mono">{truncateId(sig.owner_id)}</span>
                      {sig.owner_type && <span className="text-gray-400 ml-1">({sig.owner_type})</span>}
                    </td>
                    <td className="text-xs text-gray-500 whitespace-nowrap">{formatDate(sig.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="mt-2 text-sm text-gray-500">No signals emitted yet</p>
        )}
      </div>
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value?: string }) {
  return (
    <div className="flex items-start gap-3">
      <span className="text-gray-500 uppercase w-28 shrink-0">{label}</span>
      <span className="font-mono break-all">{value || '-'}</span>
    </div>
  )
}
