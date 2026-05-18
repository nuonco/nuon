import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Link, useParams } from 'react-router'
import { useState } from 'react'
import { getQueueSignalDetail, getSignalGraph, directExecuteSignal, reEnqueueSignal } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { JsonViewer } from '@/components/common/JsonViewer'
import { StatusHistory } from '@/components/common/StatusHistory'
import { SignalFlowGraph } from '@/components/common/SignalFlowGraph'
import { SignalTreeView } from '@/components/common/SignalTreeView'
import { TemporalWorkflowCard } from '@/components/common/TemporalWorkflowCard'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId, formatDuration } from '@/utils/format'

function getStatus(s: any): string {
  if (!s) return ''
  if (typeof s === 'string') return s
  if (typeof s === 'object' && s.status) return String(s.status)
  return String(s)
}

function getMeta(status: any, key: string): string {
  if (!status) return ''
  if (status.metadata?.[key]) return String(status.metadata[key])
  for (const h of status.history || []) {
    if (h.metadata?.[key]) return String(h.metadata[key])
  }
  return ''
}

function timeBetween(a: string, b: string): string {
  if (!a || !b) return ''
  const da = new Date(a).getTime()
  const db = new Date(b).getTime()
  if (isNaN(da) || isNaN(db)) return ''
  const ms = db - da
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
  return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`
}

export const QueueSignalDetail = () => {
  const { id: queueId, signalId } = useParams<{ id: string; signalId: string }>()
  const queryClient = useQueryClient()
  const [hideReady, setHideReady] = useState(true)
  const [sortOrder, setSortOrder] = useState<'newest' | 'oldest'>('newest')
  const [showGraph, setShowGraph] = useState(false)
  const [graphViewMode, setGraphViewMode] = useState<'tree' | 'graph'>('tree')

  const { data, isLoading, error } = useQuery({
    queryKey: ['queue-signal', queueId, signalId],
    queryFn: () => getQueueSignalDetail(queueId!, signalId!),
    enabled: !!queueId && !!signalId,
  })

  const { data: graphData } = useQuery({
    queryKey: ['signal-graph', queueId, signalId],
    queryFn: () => getSignalGraph(queueId!, signalId!, 2),
    enabled: !!queueId && !!signalId && showGraph,
  })

  const directExecuteMutation = useMutation({
    mutationFn: () => directExecuteSignal(queueId!, signalId!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['queue-signal', queueId, signalId] }),
  })

  const reEnqueueMutation = useMutation({
    mutationFn: () => reEnqueueSignal(queueId!, signalId!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['queue-signal', queueId, signalId] }),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load signal'} />
  if (!data) return null

  const { signal, queue, workflow_info: wfInfo, signal_attrs: attrs, signals_ahead: signalsAhead = [], temporal_ui_url: temporalUIUrl } = data
  const status = getStatus(signal?.status)
  const statusHistory = signal?.status?.history || []
  const enqueuedAt = signal?.created_at
  const enqueueStartedAt = getMeta(signal?.status, 'enqueue_started_at')
  const enqueueFinishedAt = getMeta(signal?.status, 'enqueue_finished_at')
  const enqueueError = getMeta(signal?.status, 'enqueue_error')
  const dequeuedAt = getMeta(signal?.status, 'dequeued_at')
  const executeStartedAt = getMeta(signal?.status, 'execute_started_at')
  const executeFinishedAt = getMeta(signal?.status, 'execute_finished_at')
  const enqueueTimingMissing = !enqueueStartedAt && !enqueueFinishedAt

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <div className="flex gap-2 text-xs text-gray-500 dark:text-gray-400">
        <Link to="/queues" className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">Queues</Link>
        <span>&rarr;</span>
        <Link to={`/queues/${queue?.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">{truncateId(queue?.id)}</Link>
        <span>&rarr;</span>
        <span>Signal</span>
      </div>

      {/* Header */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <div className="flex flex-wrap items-center gap-2 mb-2">
          <h1 className="text-lg font-semibold">Signal</h1>
          <Badge>{signal?.type}</Badge>
          <Badge variant="status" status={status}>{status}</Badge>
          <Link to={`/queues/${queueId}/signals/${signalId}/graph`} className="inline-flex items-center rounded-md bg-primary-50 dark:bg-primary-950 border border-primary-200 dark:border-primary-800 px-2 py-1 text-xs font-medium text-primary-700 dark:text-primary-300 hover:bg-primary-100 dark:hover:bg-primary-900">
            View as graph
          </Link>
          <Link to="/signal-catalog" className="text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">View catalog &rarr;</Link>
          {signal?.enqueued === false && (
            <Badge variant="status" status="failed">NOT ENQUEUED</Badge>
          )}
          <div className="ml-auto flex items-center gap-2">
            {signal?.enqueued === false && (
              <button
                onClick={() => {
                  if (confirm('Re-enqueue this signal? This will send it to the enqueuer for processing.')) {
                    reEnqueueMutation.mutate()
                  }
                }}
                disabled={reEnqueueMutation.isPending}
                className="rounded-md bg-orange-600 dark:bg-orange-500 px-2 py-1 text-xs font-medium text-white hover:bg-orange-700 dark:hover:bg-orange-600 disabled:opacity-50"
              >
                {reEnqueueMutation.isPending ? 'Re-enqueueing...' : 'Re-enqueue'}
              </button>
            )}
            <button
              onClick={() => {
                if (confirm('Are you sure you want to directly execute this signal?')) {
                  directExecuteMutation.mutate()
                }
              }}
              disabled={directExecuteMutation.isPending}
              className="rounded-md bg-red-600 dark:bg-red-500 px-2 py-1 text-xs font-medium text-white hover:bg-red-700 dark:hover:bg-red-600 disabled:opacity-50"
            >
              {directExecuteMutation.isPending ? 'Executing...' : 'Direct Execute'}
            </button>
          </div>
        </div>
        <div className="space-y-1 text-xs">
          <div><span className="text-gray-500 dark:text-gray-400 uppercase">Signal ID:</span> <span className="font-mono">{signal?.id}</span></div>
          <div><span className="text-gray-500 dark:text-gray-400 uppercase">Queue ID:</span> <Link to={`/queues/${signal?.queue_id}`} className="font-mono text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">{signal?.queue_id}</Link></div>
        </div>
      </div>

      {/* Temporal workflow stats */}
      {temporalUIUrl && signal?.workflow?.id && signal?.workflow?.namespace && (
        <TemporalWorkflowCard
          temporalUIUrl={temporalUIUrl}
          namespace={signal.workflow.namespace}
          workflowId={signal.workflow.id}
        />
      )}

      {/* Signal attributes */}
      {attrs && (
        <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Signal attributes</h2>
          <div className="mt-2 flex flex-wrap gap-2 text-xs">
            {attrs.Namespace && <Badge>ns: {attrs.Namespace}</Badge>}
            {attrs.AutoRetry && <Badge variant="status" status="online">auto-retry</Badge>}
            {attrs.HasMaxAutoRetries && <Badge>max retries: {attrs.MaxRetries}</Badge>}
            {attrs.HasCloneSteps && <Badge>clone-steps</Badge>}
            {attrs.HasNoOpCheck && <Badge>no-op-check</Badge>}
            {attrs.HasPolicyEval && <Badge>policy-eval</Badge>}
            {attrs.HasSkipCleanup && <Badge>skip-cleanup</Badge>}
            {attrs.HasOnApprove && <Badge>on-approve</Badge>}
            {attrs.HasOnRetry && <Badge>on-retry</Badge>}
            {attrs.HasOnSkip && <Badge>on-skip</Badge>}
            {attrs.HasOnDeny && <Badge>on-deny</Badge>}
            {attrs.SkipGroup && <Badge>skip-group</Badge>}
            {attrs.HasFetchSteps && <Badge>fetch-steps</Badge>}
          </div>
        </div>
      )}

      {/* Timeline */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">Timeline</h2>
        <div className="flex items-start justify-between">
          <TimelineStep label="Enqueued" value={enqueuedAt} active />
          <TimelineConnector duration={timeBetween(enqueueStartedAt, enqueueFinishedAt)} />
          {enqueueError ? (
            <TimelineStep label="Enqueue failed" value={enqueueError} active={false} />
          ) : (
            <TimelineStep label="Enqueue finished" value={enqueueFinishedAt} active={!!enqueueFinishedAt} />
          )}
          <TimelineConnector duration={timeBetween(enqueueFinishedAt, dequeuedAt)} />
          <TimelineStep label="Dequeued" value={dequeuedAt} active={!!dequeuedAt} />
          <TimelineConnector duration={timeBetween(dequeuedAt, executeStartedAt)} />
          <TimelineStep label="Execute started" value={executeStartedAt} active={!!executeStartedAt} />
          <TimelineConnector duration={timeBetween(executeStartedAt, executeFinishedAt)} />
          <TimelineStep label="Execute finished" value={executeFinishedAt} active={!!executeFinishedAt} />
        </div>
        {enqueueTimingMissing && (
          <div className="mt-3 text-xs text-gray-500 dark:text-gray-400">
            Enqueue timing metadata not available (signal created before tracking was added).
          </div>
        )}
      </div>

      {/* Signals ahead */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Signals ahead</h2>
        {signalsAhead.length > 0 ? (
          <div className="mt-2 space-y-1">
            {signalsAhead.map((sig: any) => (
              <div key={sig.id} className="flex items-center gap-3 p-2 border border-gray-100 dark:border-gray-800 rounded text-xs">
                <Link to={`/queues/${queue?.id}/signals/${sig.id}`} className="font-mono text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 truncate flex-1">{sig.id}</Link>
                <Badge>{sig.type}</Badge>
                <Badge variant="status" status={getStatus(sig.status)}>{getStatus(sig.status)}</Badge>
                <span className="text-gray-400 dark:text-gray-500 whitespace-nowrap">{formatDate(sig.created_at)}</span>
              </div>
            ))}
          </div>
        ) : (
          <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">No signals ahead</p>
        )}
      </div>

      {/* Executions */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Executions</h2>
        <div className="mt-2 flex items-center gap-4">
          <div>
            <p className="text-xs text-gray-500 dark:text-gray-400 uppercase">Execution count</p>
            <p className={`text-2xl font-bold font-mono ${signal?.execution_count > 1 ? 'text-orange-500' : ''}`}>
              {signal?.execution_count ?? 0}
            </p>
          </div>
        </div>
        {wfInfo && (
          <div className="mt-3 border-t border-gray-200 dark:border-gray-800 pt-3 space-y-2 text-xs">
            <div><span className="text-gray-500 dark:text-gray-400 uppercase w-32 inline-block">Status</span> <Badge variant="status" status={wfInfo.status}>{wfInfo.status}</Badge></div>
            {wfInfo.update_executions?.length > 0 && (
              <div><span className="text-gray-500 dark:text-gray-400 uppercase w-32 inline-block">Updates</span> <span className="font-mono">{wfInfo.update_executions.length}</span></div>
            )}
            {wfInfo.activities?.length > 0 && (
              <div><span className="text-gray-500 dark:text-gray-400 uppercase w-32 inline-block">Activities</span> <span className="font-mono">{wfInfo.activities.length}</span></div>
            )}
            {wfInfo.awaited_signals?.length > 0 && (
              <div><span className="text-gray-500 dark:text-gray-400 uppercase w-32 inline-block">Awaited</span> <span className="font-mono text-orange-500">{wfInfo.awaited_signals.length} signal{wfInfo.awaited_signals.length !== 1 ? 's' : ''}</span></div>
            )}
            {wfInfo.enqueued_signals?.length > 0 && (
              <div><span className="text-gray-500 dark:text-gray-400 uppercase w-32 inline-block">Enqueued</span> <span className="font-mono text-green-500">{wfInfo.enqueued_signals.length} signal{wfInfo.enqueued_signals.length !== 1 ? 's' : ''}</span></div>
            )}
            {/* Failures */}
            {(wfInfo.status === 'Failed' || wfInfo.status === 'Timed Out') && (
              <>
                {wfInfo.update_executions?.filter((ue: any) => ue.failure).map((ue: any, i: number) => (
                  <div key={i} className="mt-1">
                    <Badge variant="status" status="failed">{ue.name}</Badge>
                    <pre className="mt-1 text-xs text-red-600 dark:text-red-400 font-mono whitespace-pre-wrap bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded p-2">{ue.failure}</pre>
                  </div>
                ))}
                {wfInfo.orphan_activities?.filter((a: any) => a.failure).map((a: any, i: number) => (
                  <div key={i} className="mt-1">
                    <Badge variant="status" status="failed">{a.name}</Badge>
                    <pre className="mt-1 text-xs text-red-600 dark:text-red-400 font-mono whitespace-pre-wrap bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded p-2">{a.failure}</pre>
                  </div>
                ))}
              </>
            )}
          </div>
        )}
      </div>

      {/* Dependencies: awaited signals, enqueued signals, child workflows */}
      {wfInfo && (
        <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">
            Dependencies
            <span className="ml-2 text-xs font-normal text-gray-500 dark:text-gray-400">
              {(wfInfo.awaited_signals?.length || 0) + (wfInfo.enqueued_signals?.length || 0) + (wfInfo.child_workflows?.length || 0)} total
            </span>
          </h2>

          {wfInfo.enqueued_signals?.length > 0 && (
            <div className="mt-3">
              <h3 className="text-xs font-semibold text-green-700 dark:text-green-300 uppercase mb-2">
                ▸ Enqueued signals ({wfInfo.enqueued_signals.length})
              </h3>
              <div className="space-y-1">
                {wfInfo.enqueued_signals.map((es: any, i: number) => {
                  const sigStatus = es.signal ? getStatus(es.signal.status) : ''
                  const signalLink = es.signal?.queue_id
                    ? `/queues/${es.signal.queue_id}/signals/${es.queue_signal_id}`
                    : `/queue-signals?search=${es.queue_signal_id}`
                  return (
                    <div key={i} className="flex items-center gap-3 p-2 border border-green-100 dark:border-green-900 bg-green-50/50 dark:bg-green-900/20 rounded text-xs">
                      <span className="text-green-500 text-[8px]">▸</span>
                      <Link to={signalLink} className="font-mono text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 truncate">{es.queue_signal_id}</Link>
                      {es.signal?.type && <Badge>{es.signal.type}</Badge>}
                      {sigStatus && <Badge variant="status" status={sigStatus}>{sigStatus}</Badge>}
                      {es.activity_name && <span className="text-gray-400 dark:text-gray-500 font-mono text-[10px] ml-auto">{es.activity_name}</span>}
                    </div>
                  )
                })}
              </div>
            </div>
          )}

          {wfInfo.awaited_signals?.length > 0 && (
            <div className="mt-3">
              <h3 className="text-xs font-semibold text-orange-700 dark:text-orange-300 uppercase mb-2">
                ◇ Awaited signals ({wfInfo.awaited_signals.length})
              </h3>
              <div className="space-y-1">
                {wfInfo.awaited_signals.map((as: any, i: number) => {
                  const sigStatus = as.signal ? getStatus(as.signal.status) : as.status
                  const signalLink = as.signal?.queue_id
                    ? `/queues/${as.signal.queue_id}/signals/${as.queue_signal_id}`
                    : `/queue-signals?search=${as.queue_signal_id}`
                  return (
                    <div key={i} className="flex items-center gap-3 p-2 border border-orange-100 dark:border-orange-900 bg-orange-50/50 dark:bg-orange-900/20 rounded text-xs">
                      <span className="text-orange-500 text-[8px]">◇</span>
                      <Link to={signalLink} className="font-mono text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 truncate">{as.queue_signal_id}</Link>
                      {as.signal?.type && <Badge>{as.signal.type}</Badge>}
                      <Badge variant="status" status={sigStatus}>{sigStatus}</Badge>
                      <span className="text-gray-400 dark:text-gray-500 font-mono ml-auto">{formatDuration(as.duration)}</span>
                    </div>
                  )
                })}
              </div>
            </div>
          )}

          {wfInfo.child_workflows?.length > 0 && (
            <div className="mt-3">
              <h3 className="text-xs font-semibold text-blue-700 dark:text-blue-300 uppercase mb-2">
                ⑂ Child workflows ({wfInfo.child_workflows.length})
              </h3>
              <div className="space-y-1">
                {wfInfo.child_workflows.map((cw: any, i: number) => (
                  <div key={i} className="flex items-center gap-3 p-2 border border-blue-100 dark:border-blue-900 bg-blue-50/50 dark:bg-blue-900/20 rounded text-xs">
                    <span className="text-blue-500 text-[8px]">⑂</span>
                    <Link
                      to={`/temporal-workflows?namespace=${cw.namespace}&workflow_id=${cw.workflow_id}`}
                      className="font-mono text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 truncate"
                    >
                      {cw.workflow_type || cw.workflow_id}
                    </Link>
                    <Badge variant="status" status={cw.status}>{cw.status}</Badge>
                    {cw.namespace && <span className="text-gray-400 dark:text-gray-500 font-mono text-[10px]">{cw.namespace}</span>}
                    <span className="text-gray-400 dark:text-gray-500 font-mono ml-auto">{formatDuration(cw.duration)}</span>
                    {temporalUIUrl && cw.workflow_id && cw.namespace && (
                      <a
                        href={`${temporalUIUrl}/namespaces/${cw.namespace}/workflows/${cw.workflow_id}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 whitespace-nowrap"
                        onClick={(e) => e.stopPropagation()}
                      >
                        Temporal &rarr;
                      </a>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {!wfInfo.awaited_signals?.length && !wfInfo.enqueued_signals?.length && !wfInfo.child_workflows?.length && (
            <p className="mt-2 text-xs text-gray-500 dark:text-gray-400">No dependencies</p>
          )}
        </div>
      )}

      {/* Signal tree / graph */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Signal tree</h2>
          <div className="flex items-center gap-2">
            {showGraph && (
              <div className="flex rounded-md border border-gray-300 dark:border-gray-700 text-xs overflow-hidden">
                <button
                  onClick={() => setGraphViewMode('tree')}
                  className={`px-2.5 py-1 font-medium ${graphViewMode === 'tree' ? 'bg-primary-600 dark:bg-primary-500 text-white' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'}`}
                >
                  Tree
                </button>
                <button
                  onClick={() => setGraphViewMode('graph')}
                  className={`px-2.5 py-1 font-medium ${graphViewMode === 'graph' ? 'bg-primary-600 dark:bg-primary-500 text-white' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'}`}
                >
                  Graph
                </button>
              </div>
            )}
            <button
              onClick={() => setShowGraph(!showGraph)}
              className={`rounded-md px-3 py-1.5 text-xs font-medium ${showGraph ? 'bg-primary-600 dark:bg-primary-500 text-white' : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'}`}
            >
              {showGraph ? 'Hide' : 'Load tree'}
            </button>
          </div>
        </div>
        {showGraph && graphData?.graph && (
          <div className="mt-3">
            {graphViewMode === 'tree' ? (
              <SignalTreeView graphData={graphData.graph} temporalUIUrl={temporalUIUrl} height="32rem" />
            ) : (
              <SignalFlowGraph graphData={graphData.graph} height="32rem" />
            )}
          </div>
        )}
        {showGraph && !graphData?.graph && (
          <p className="mt-2 text-xs text-gray-500 dark:text-gray-400">Loading graph...</p>
        )}
      </div>

      {/* Execution waterfall */}
      {wfInfo && (
        <ExecutionWaterfall wfInfo={wfInfo} temporalUIUrl={temporalUIUrl} hideReady={hideReady} setHideReady={setHideReady} sortOrder={sortOrder} setSortOrder={setSortOrder} />
      )}

      {/* Update handlers */}
      {wfInfo?.update_handlers?.length > 0 && (
        <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Update handlers</h2>
          <div className="mt-2 flex flex-wrap gap-2">
            {wfInfo.update_handlers.map((h: string) => <Badge key={h}>{h}</Badge>)}
          </div>
        </div>
      )}

      {/* Signal info + Handler */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-2">Signal info</h2>
          <table className="w-full text-xs table-fixed">
            <colgroup>
              <col className="w-36" />
              <col />
            </colgroup>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
              <InfoTableRow label="Type" value={signal?.type} badge />
              <InfoTableRow label="Status" value={status} statusBadge />
              <InfoTableRow label="Exec count" value={String(signal?.execution_count ?? 0)} highlight={signal?.execution_count > 1} />
              {signal?.status?.status_human_description && <InfoTableRow label="Description" value={signal.status.status_human_description} />}
              <InfoTableRow label="Enqueued" value={signal?.enqueued ? 'Yes' : 'No'} statusBadge />
              <InfoTableRow label="Enqueue source" value={getMeta(signal?.status, 'enqueue_source') || '-'} badge={!!getMeta(signal?.status, 'enqueue_source')} />
              <InfoTableRow label="Enqueue duration" value={timeBetween(enqueueStartedAt, enqueueFinishedAt) || '-'} mono />
              <InfoTableRow label="Execute duration" value={timeBetween(executeStartedAt, executeFinishedAt) || '-'} mono />
              <InfoTableRow label="Created" value={formatDate(signal?.created_at)} />
              <InfoTableRow label="Updated" value={formatDate(signal?.updated_at)} />
              {enqueueError && <InfoTableRow label="Enqueue error" value={enqueueError} error />}
            </tbody>
          </table>
        </div>
        <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-2">Handler</h2>
          <table className="w-full text-xs table-fixed">
            <colgroup>
              <col className="w-36" />
              <col />
            </colgroup>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
              <InfoTableRow label="Owner type" value={signal?.owner_type} />
              <InfoTableRow label="Owner ID" value={signal?.owner_id} mono />
              {signal?.emitter_id && (
                <tr>
                  <td className="py-1.5 pr-3 text-gray-500 dark:text-gray-400 uppercase whitespace-nowrap align-top font-medium">Emitter</td>
                  <td className="py-1.5 overflow-hidden text-ellipsis">
                    <Link to={`/queues/${queue?.id}/emitters/${signal.emitter_id}`} className="font-mono text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 break-all">{signal.emitter_id}</Link>
                  </td>
                </tr>
              )}
              {signal?.workflow?.id && <InfoTableRow label="Workflow ID" value={signal.workflow.id} mono />}
              {signal?.workflow?.namespace && <InfoTableRow label="Namespace" value={signal.workflow.namespace} />}
            </tbody>
          </table>
          <div className="mt-3 pt-3 border-t border-gray-200 dark:border-gray-800 flex flex-wrap gap-3 text-xs">
            {signal?.workflow?.id && signal?.workflow?.namespace && (
              <Link to={`/temporal-workflows?namespace=${signal.workflow.namespace}&workflow_id=${signal.workflow.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
                View handler workflow &rarr;
              </Link>
            )}
            {temporalUIUrl && signal?.workflow?.id && signal?.workflow?.namespace && (
              <a href={`${temporalUIUrl}/namespaces/${signal.workflow.namespace}/workflows/${signal.workflow.id}`} target="_blank" rel="noopener noreferrer" className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
                View in Temporal UI &rarr;
              </a>
            )}
            <Link to={`/queue-signals?search=${signal?.owner_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
              Owner signals &rarr;
            </Link>
            {signal?.owner_type === 'install_workflow_steps' && (
              <Link to={`/workflows?search=${signal?.owner_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">Step's workflow &rarr;</Link>
            )}
            {signal?.owner_type === 'install_workflows' && (
              <Link to={`/workflows?search=${signal?.owner_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">Install workflow &rarr;</Link>
            )}
          </div>
        </div>
      </div>

      {/* Signal data */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Signal data</h2>
        <div className="mt-2">
          <JsonViewer data={signal?.signal || signal} />
        </div>
      </div>

      {/* Status history */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Status history</h2>
        <div className="mt-2">
          <StatusHistory status={signal?.status} maxCollapsed={5} />
        </div>
      </div>
    </div>
  )
}

function TimelineStep({ label, value, active }: { label: string; value?: string; active: boolean }) {
  return (
    <div className="flex flex-col items-center text-center min-w-0 shrink-0">
      <div className={`w-3 h-3 rounded-full mb-2 ${active ? 'bg-primary-500' : 'bg-gray-300 dark:bg-gray-600'}`} />
      <p className="text-[10px] text-gray-500 dark:text-gray-400 uppercase font-medium">{label}</p>
      {value ? (
        <p className="text-[10px] font-mono mt-0.5 max-w-[140px] break-all">{formatDate(value)}</p>
      ) : (
        <p className="text-[10px] font-mono mt-0.5 text-gray-300 dark:text-gray-400">&mdash;</p>
      )}
    </div>
  )
}

function TimelineConnector({ duration }: { duration: string }) {
  return (
    <div className="flex flex-col items-center flex-1 min-w-[40px] pt-1">
      <div className="h-px w-full bg-gray-200 dark:bg-gray-700 mt-1" />
      {duration && <p className="text-[10px] font-mono text-primary-500 mt-1">{duration}</p>}
    </div>
  )
}

function InfoRow({ label, value, highlight }: { label: string; value?: string; highlight?: boolean }) {
  return (
    <div className="flex items-start gap-3">
      <span className="text-gray-500 dark:text-gray-400 uppercase w-28 shrink-0">{label}</span>
      <span className={`font-mono break-all ${highlight ? 'text-orange-500 font-bold' : ''}`}>{value || '-'}</span>
    </div>
  )
}

function InfoTableRow({ label, value, highlight, badge, statusBadge, mono, error: isError }: {
  label: string; value?: string; highlight?: boolean; badge?: boolean; statusBadge?: boolean; mono?: boolean; error?: boolean;
}) {
  return (
    <tr>
      <td className="py-1.5 pr-3 text-gray-500 dark:text-gray-400 uppercase whitespace-nowrap align-top font-medium">{label}</td>
      <td className={`py-1.5 overflow-hidden text-ellipsis ${highlight ? 'text-orange-500 font-bold font-mono' : ''} ${mono ? 'font-mono' : ''} ${isError ? 'text-red-600 dark:text-red-400 font-mono' : ''}`}>
        {statusBadge && value
          ? <Badge variant="status" status={value}>{value}</Badge>
          : badge && value
            ? <Badge>{value}</Badge>
            : <span className="break-all">{value || '-'}</span>}
      </td>
    </tr>
  )
}

function StatusHistoryEntry({ h, isCurrent }: { h: any; isCurrent?: boolean }) {
  if (!h) return null
  const status = getStatus(h)
  return (
    <div className="flex items-start gap-3 text-xs border-b border-gray-100 dark:border-gray-800 pb-2 last:border-0">
      <Badge variant="status" status={status}>{status}</Badge>
      <div className="flex-1 space-y-0.5">
        <div>
          {status}
          {isCurrent && <span className="text-primary-600 dark:text-primary-400 font-medium ml-1">(current)</span>}
          {h.status_human_description && <span className="text-gray-500 dark:text-gray-400 ml-1">— {h.status_human_description}</span>}
        </div>
        {h.created_at_ts > 0 && (
          <div className="text-gray-400 dark:text-gray-500 font-mono">{new Date(h.created_at_ts / 1000000).toISOString().replace('T', ' ').slice(0, 19)} UTC</div>
        )}
        {h.metadata && Object.keys(h.metadata).length > 0 && (
          <div className="flex flex-wrap gap-x-4 gap-y-0.5 text-gray-400 dark:text-gray-500">
            {Object.entries(h.metadata).map(([k, v]) => (
              <span key={k}><span className="text-gray-500 dark:text-gray-400">{k}:</span> {String(v)}</span>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

// -- Execution Waterfall --

function ExecutionWaterfall({ wfInfo, temporalUIUrl, hideReady, setHideReady, sortOrder, setSortOrder }: {
  wfInfo: any; temporalUIUrl: string;
  hideReady: boolean; setHideReady: (v: boolean) => void;
  sortOrder: 'newest' | 'oldest'; setSortOrder: (v: 'newest' | 'oldest') => void;
}) {
  const allUpdates: any[] = wfInfo.update_executions || []
  const orphans = wfInfo.orphan_activities || []
  const childWfs = wfInfo.child_workflows || []
  const awaited = wfInfo.awaited_signals || []
  const enqueued = wfInfo.enqueued_signals || []

  // Filter
  const HIDE_NAMES = new Set(['ready', 'Ready'])
  const updates = hideReady ? allUpdates.filter((ue: any) => !HIDE_NAMES.has(ue.name)) : allUpdates

  // Sort
  const sortedUpdates = [...updates].sort((a, b) => {
    const ta = new Date(a.started_at).getTime() || 0
    const tb = new Date(b.started_at).getTime() || 0
    return sortOrder === 'newest' ? tb - ta : ta - tb
  })

  const totalItems = allUpdates.length + (orphans.length > 0 ? 1 : 0) + childWfs.length + awaited.length + enqueued.length
  if (totalItems === 0) return null

  return (
    <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <div>
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Execution waterfall</h2>
          <p className="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
            {updates.length}/{allUpdates.length} update{allUpdates.length !== 1 ? 's' : ''}, {childWfs.length} child wf{childWfs.length !== 1 ? 's' : ''}, {awaited.length} awaited, {enqueued.length} enqueued
          </p>
        </div>
        <div className="flex items-center gap-2 text-xs">
          <label className="flex items-center gap-1 text-gray-600 dark:text-gray-400">
            <input type="checkbox" checked={hideReady} onChange={(e) => setHideReady(e.target.checked)} />
            Hide "ready" updates
          </label>
          <select value={sortOrder} onChange={(e) => setSortOrder(e.target.value as any)} className="rounded border-gray-300 dark:border-gray-700 text-xs py-1 px-2">
            <option value="newest">Newest first</option>
            <option value="oldest">Oldest first</option>
          </select>
        </div>
      </div>

      <div className="mt-3 relative">
        {/* Vertical line */}
        <div className="absolute left-3 top-0 bottom-0 w-px bg-gray-200 dark:bg-gray-700" />

        <div className="space-y-1">
          {/* Update executions */}
          {sortedUpdates.map((ue: any, i: number) => (
            <WaterfallUpdateNode key={`ue-${ue.update_id || i}`} ue={ue} awaited={awaited} childWfs={childWfs} temporalUIUrl={temporalUIUrl} />
          ))}

          {/* Orphan activities */}
          {orphans.length > 0 && (
            <WaterfallSection icon="○" label={`${orphans.length} other activit${orphans.length !== 1 ? 'ies' : 'y'}`} status="">
              {orphans.map((act: any, i: number) => (
                <WaterfallActivityRow key={i} act={act} />
              ))}
            </WaterfallSection>
          )}

          {/* Standalone child workflows (not nested under an update) */}
          {childWfs.length > 0 && updates.length === 0 && (
            <WaterfallSection icon="⑂" label={`${childWfs.length} child workflow${childWfs.length !== 1 ? 's' : ''}`} status="">
              {childWfs.map((cw: any, i: number) => (
                <WaterfallChildWorkflow key={i} cw={cw} temporalUIUrl={temporalUIUrl} />
              ))}
            </WaterfallSection>
          )}

          {/* Standalone awaited signals (not nested under an update) */}
          {awaited.length > 0 && updates.length === 0 && (
            <WaterfallSection icon="◇" label={`${awaited.length} awaited signal${awaited.length !== 1 ? 's' : ''}`} status="">
              {awaited.map((as: any, i: number) => (
                <WaterfallAwaitedSignal key={i} as={as} />
              ))}
            </WaterfallSection>
          )}

          {/* Standalone enqueued signals (not nested under an update) */}
          {enqueued.length > 0 && updates.length === 0 && (
            <WaterfallSection icon="▸" label={`${enqueued.length} enqueued signal${enqueued.length !== 1 ? 's' : ''}`} status="">
              {enqueued.map((es: any, i: number) => (
                <WaterfallEnqueuedSignal key={i} es={es} />
              ))}
            </WaterfallSection>
          )}
        </div>
      </div>
    </div>
  )
}

function WaterfallUpdateNode({ ue, awaited, childWfs, temporalUIUrl }: { ue: any; awaited: any[]; childWfs: any[]; temporalUIUrl: string }) {
  const [expanded, setExpanded] = useState(true)
  const activities = ue.activities || []

  // Use update-level signals if available, otherwise fall back to time-range overlap
  const relatedAwaited = ue.awaited_signals?.length > 0
    ? ue.awaited_signals
    : awaited.filter((as: any) => {
        const ueStart = new Date(ue.started_at).getTime()
        const ueEnd = ue.finished_at ? new Date(ue.finished_at).getTime() : Date.now()
        const asStart = new Date(as.started_at).getTime()
        return asStart >= ueStart && asStart <= ueEnd
      })
  const relatedEnqueued: any[] = ue.enqueued_signals || []
  const ueStart = new Date(ue.started_at).getTime()
  const ueEnd = ue.finished_at ? new Date(ue.finished_at).getTime() : Date.now()
  const relatedChildWfs = childWfs.filter((cw: any) => {
    const cwStart = new Date(cw.started_at).getTime()
    return cwStart >= ueStart && cwStart <= ueEnd
  })

  const statusColor = ue.status === 'Completed' ? 'bg-green-500' : ue.status === 'Failed' ? 'bg-red-500' : ue.status === 'Running' ? 'bg-primary-500 animate-pulse' : 'bg-gray-400 dark:bg-gray-500'

  return (
    <div className="relative pl-7">
      {/* Dot on the vertical line */}
      <div className={`absolute left-1.5 top-2.5 w-3 h-3 rounded-full ${statusColor} ring-2 ring-white dark:ring-gray-950`} />

      <div className="border border-gray-200 dark:border-gray-800 rounded-lg overflow-hidden">
        <button
          onClick={() => setExpanded(!expanded)}
          className="w-full flex items-center justify-between px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-800 text-left"
        >
          <div className="flex items-center gap-2 text-xs flex-wrap">
            <Badge variant="status" status={ue.status}>{ue.status}</Badge>
            <span className="font-mono font-semibold">{ue.name}</span>
            <span className="text-gray-400 dark:text-gray-500">{formatDuration(ue.duration)}</span>
            <span className="text-gray-400 dark:text-gray-500">{activities.length} activit{activities.length !== 1 ? 'ies' : 'y'}</span>
            {relatedAwaited.length > 0 && <span className="text-orange-500">{relatedAwaited.length} awaited</span>}
            {relatedEnqueued.length > 0 && <span className="text-green-500">{relatedEnqueued.length} enqueued</span>}
            {relatedChildWfs.length > 0 && <span className="text-blue-500">{relatedChildWfs.length} child wf</span>}
            <span className="text-gray-300 dark:text-gray-400 font-mono ml-auto">{ue.started_at ? formatDate(ue.started_at) : ''}</span>
          </div>
          <span className="text-gray-400 dark:text-gray-500 text-xs">{expanded ? '▾' : '▸'}</span>
        </button>

        {expanded && (
          <div className="border-t border-gray-200 dark:border-gray-800">
            {/* Update metadata */}
            {(ue.input || ue.result || ue.failure) && (
              <div className="px-3 py-2 border-b border-gray-100 dark:border-gray-800 text-xs space-y-1">
                {ue.failure && (
                  <pre className="font-mono text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded p-2 overflow-x-auto max-h-24 whitespace-pre-wrap">{ue.failure}</pre>
                )}
                {ue.input && (
                  <details className="group">
                    <summary className="text-gray-500 dark:text-gray-400 cursor-pointer hover:text-gray-700 dark:hover:text-gray-200">Input</summary>
                    <pre className="mt-1 font-mono rounded p-2 overflow-x-auto max-h-24 text-[10px]">{ue.input}</pre>
                  </details>
                )}
                {ue.result && (
                  <details className="group">
                    <summary className="text-gray-500 dark:text-gray-400 cursor-pointer hover:text-gray-700 dark:hover:text-gray-200">Result</summary>
                    <pre className="mt-1 font-mono rounded p-2 overflow-x-auto max-h-24 text-[10px]">{ue.result}</pre>
                  </details>
                )}
              </div>
            )}

            {/* Activities + awaited signals + child workflows as waterfall items */}
            <div className="divide-y divide-gray-100 dark:divide-gray-800">
              {activities.map((act: any, i: number) => (
                <WaterfallActivityRow key={`act-${i}`} act={act} />
              ))}
              {relatedAwaited.map((as: any, i: number) => (
                <WaterfallAwaitedSignal key={`as-${i}`} as={as} />
              ))}
              {relatedEnqueued.map((es: any, i: number) => (
                <WaterfallEnqueuedSignal key={`es-${i}`} es={es} />
              ))}
              {relatedChildWfs.map((cw: any, i: number) => (
                <WaterfallChildWorkflow key={`cw-${i}`} cw={cw} temporalUIUrl={temporalUIUrl} />
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function extractSignalIdFromActivity(act: any): string | null {
  const name = act.name || ''
  if (name.includes('AwaitSignal')) {
    return extractQsiFromJson(act.input)
  }
  if (name.includes('EnqueueSignal')) {
    return extractQsiFromJson(act.result) || extractQsiFromJson(act.input)
  }
  return null
}

function extractQsiFromJson(raw: string | undefined): string | null {
  if (!raw) return null
  try {
    const parsed = JSON.parse(raw)
    if (typeof parsed === 'string' && parsed.startsWith('qsi')) return parsed
    if (typeof parsed === 'object' && parsed !== null) {
      for (const key of ['id', 'ID', 'queue_signal_id']) {
        if (typeof parsed[key] === 'string' && parsed[key].startsWith('qsi')) return parsed[key]
      }
    }
  } catch { /* ignore */ }
  return null
}

function WaterfallActivityRow({ act }: { act: any }) {
  const [showDetail, setShowDetail] = useState(false)
  const dot = act.status === 'Completed' ? 'text-green-500' : act.status === 'Failed' ? 'text-red-500' : act.status === 'Running' ? 'text-primary-500' : 'text-gray-400 dark:text-gray-500'
  const signalId = extractSignalIdFromActivity(act)
  const isSignalActivity = !!signalId

  return (
    <>
      <div className="flex items-center gap-2 px-3 py-1.5 text-xs hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer" onClick={() => setShowDetail(!showDetail)}>
        <span className={`${dot} text-[8px]`}>●</span>
        <span className="font-mono text-gray-900 dark:text-gray-100 flex-1 truncate">{act.name}</span>
        {isSignalActivity && (
          <Link to={`/queue-signals?search=${signalId}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 font-mono text-[10px]" onClick={(e) => e.stopPropagation()}>
            {truncateId(signalId!)} →
          </Link>
        )}
        <Badge variant="status" status={act.status}>{act.status}</Badge>
        <span className="text-gray-400 dark:text-gray-500 font-mono w-16 text-right">{formatDuration(act.duration)}</span>
        {act.attempt > 1 && <span className="text-orange-500">×{act.attempt}</span>}
      </div>
      {showDetail && (
        <div className="px-3 py-2 text-[10px] space-y-1 border-t border-gray-100 dark:border-gray-800">
          {act.failure && <pre className="font-mono text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/30 rounded p-1.5 whitespace-pre-wrap">{act.failure}</pre>}
          {act.input && <details><summary className="text-gray-500 dark:text-gray-400 cursor-pointer">Input</summary><pre className="mt-0.5 font-mono rounded p-1.5 overflow-x-auto max-h-20">{act.input}</pre></details>}
          {act.result && <details><summary className="text-gray-500 dark:text-gray-400 cursor-pointer">Result</summary><pre className="mt-0.5 font-mono rounded p-1.5 overflow-x-auto max-h-20">{act.result}</pre></details>}
        </div>
      )}
    </>
  )
}

function WaterfallAwaitedSignal({ as: asig }: { as: any }) {
  const [showDetail, setShowDetail] = useState(false)
  const signalStatus = asig.signal ? getStatus(asig.signal.status) : asig.status

  return (
    <>
      <div className="flex items-center gap-2 px-3 py-1.5 text-xs hover:bg-orange-50 dark:hover:bg-orange-900/30 cursor-pointer" onClick={() => setShowDetail(!showDetail)}>
        <span className="text-orange-500 text-[8px]">◇</span>
        <span className="text-orange-700 dark:text-orange-300 font-medium">await signal</span>
        {asig.queue_signal_id && (
          <Link to={`/queue-signals?search=${asig.queue_signal_id}`} className="font-mono text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300" onClick={(e) => e.stopPropagation()}>
            {truncateId(asig.queue_signal_id)}
          </Link>
        )}
        <Badge variant="status" status={asig.status}>{asig.status}</Badge>
        <span className="text-gray-400 dark:text-gray-500 font-mono ml-auto w-16 text-right">{formatDuration(asig.duration)}</span>
      </div>
      {showDetail && asig.signal && (
        <div className="px-3 py-2 bg-orange-50/50 dark:bg-orange-900/30 text-[10px] space-y-1 border-t border-orange-100 dark:border-orange-900">
          <div className="grid grid-cols-2 gap-1 sm:grid-cols-4">
            <div><span className="text-gray-500 dark:text-gray-400">Type:</span> <span className="font-mono">{asig.signal.type}</span></div>
            <div><span className="text-gray-500 dark:text-gray-400">Signal status:</span> <Badge variant="status" status={signalStatus}>{signalStatus}</Badge></div>
            <div><span className="text-gray-500 dark:text-gray-400">Queue:</span> <Link to={`/queues/${asig.signal.queue_id}`} className="font-mono text-primary-600 dark:text-primary-400">{truncateId(asig.signal.queue_id)}</Link></div>
            <div><span className="text-gray-500 dark:text-gray-400">Owner:</span> <span className="font-mono">{truncateId(asig.signal.owner_id)}</span></div>
          </div>
          {asig.failure && <pre className="font-mono text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/30 rounded p-1.5 whitespace-pre-wrap">{asig.failure}</pre>}
        </div>
      )}
    </>
  )
}

function WaterfallEnqueuedSignal({ es }: { es: any }) {
  const [showDetail, setShowDetail] = useState(false)
  const signalStatus = es.signal ? getStatus(es.signal.status) : ''

  return (
    <>
      <div className="flex items-center gap-2 px-3 py-1.5 text-xs hover:bg-green-50 dark:hover:bg-green-900/30 cursor-pointer" onClick={() => setShowDetail(!showDetail)}>
        <span className="text-green-500 text-[8px]">▸</span>
        <span className="text-green-700 dark:text-green-300 font-medium">enqueue signal</span>
        {es.queue_signal_id && (
          <Link to={`/queue-signals?search=${es.queue_signal_id}`} className="font-mono text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300" onClick={(e) => e.stopPropagation()}>
            {truncateId(es.queue_signal_id)}
          </Link>
        )}
        {es.activity_name && <span className="text-gray-400 dark:text-gray-500 font-mono">{es.activity_name}</span>}
        {signalStatus && <Badge variant="status" status={signalStatus}>{signalStatus}</Badge>}
      </div>
      {showDetail && es.signal && (
        <div className="px-3 py-2 bg-green-50/50 dark:bg-green-900/30 text-[10px] space-y-1 border-t border-green-100 dark:border-green-900">
          <div className="grid grid-cols-2 gap-1 sm:grid-cols-4">
            <div><span className="text-gray-500 dark:text-gray-400">Type:</span> <span className="font-mono">{es.signal.type}</span></div>
            <div><span className="text-gray-500 dark:text-gray-400">Signal status:</span> <Badge variant="status" status={signalStatus}>{signalStatus}</Badge></div>
            <div><span className="text-gray-500 dark:text-gray-400">Queue:</span> <Link to={`/queues/${es.signal.queue_id}`} className="font-mono text-primary-600 dark:text-primary-400">{truncateId(es.signal.queue_id)}</Link></div>
            <div><span className="text-gray-500 dark:text-gray-400">Owner:</span> <span className="font-mono">{truncateId(es.signal.owner_id)}</span></div>
          </div>
        </div>
      )}
    </>
  )
}

function WaterfallChildWorkflow({ cw, temporalUIUrl }: { cw: any; temporalUIUrl: string }) {
  const [showDetail, setShowDetail] = useState(false)

  return (
    <>
      <div className="flex items-center gap-2 px-3 py-1.5 text-xs hover:bg-blue-50 dark:hover:bg-blue-900/30 cursor-pointer" onClick={() => setShowDetail(!showDetail)}>
        <span className="text-blue-500 text-[8px]">⑂</span>
        <span className="text-blue-700 dark:text-blue-300 font-medium">child workflow</span>
        <span className="font-mono text-gray-900 dark:text-gray-100">{cw.workflow_type}</span>
        <Badge variant="status" status={cw.status}>{cw.status}</Badge>
        <span className="text-gray-400 dark:text-gray-500 font-mono ml-auto w-16 text-right">{formatDuration(cw.duration)}</span>
      </div>
      {showDetail && (
        <div className="px-3 py-2 bg-blue-50/50 dark:bg-blue-900/30 text-[10px] space-y-1 border-t border-blue-100 dark:border-blue-900">
          <div className="grid grid-cols-2 gap-1 sm:grid-cols-4">
            <div><span className="text-gray-500 dark:text-gray-400">Namespace:</span> {cw.namespace}</div>
            <div><span className="text-gray-500 dark:text-gray-400">Workflow ID:</span> <span className="font-mono">{cw.workflow_id}</span></div>
            <div><span className="text-gray-500 dark:text-gray-400">Run ID:</span> <span className="font-mono">{truncateId(cw.run_id)}</span></div>
            {temporalUIUrl && (
              <div>
                <a href={`${temporalUIUrl}/namespaces/${cw.namespace}/workflows/${cw.workflow_id}`} target="_blank" rel="noopener noreferrer" className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
                  View in Temporal &rarr;
                </a>
              </div>
            )}
          </div>
          {cw.failure && <pre className="font-mono text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/30 rounded p-1.5 whitespace-pre-wrap">{cw.failure}</pre>}
        </div>
      )}
    </>
  )
}

function WaterfallSection({ icon, label, status, children }: { icon: string; label: string; status: string; children: React.ReactNode }) {
  const [expanded, setExpanded] = useState(true)
  return (
    <div className="relative pl-7">
      <div className="absolute left-1.5 top-2.5 w-3 h-3 rounded-full bg-gray-300 dark:bg-gray-600 ring-2 ring-white dark:ring-gray-950" />
      <div className="border border-gray-200 dark:border-gray-800 rounded-lg overflow-hidden">
        <button onClick={() => setExpanded(!expanded)} className="w-full flex items-center justify-between px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-800 text-left">
          <div className="flex items-center gap-2 text-xs">
            <span>{icon}</span>
            <span className="font-medium text-gray-700 dark:text-gray-300">{label}</span>
            {status && <Badge variant="status" status={status}>{status}</Badge>}
          </div>
          <span className="text-gray-400 dark:text-gray-500 text-xs">{expanded ? '▾' : '▸'}</span>
        </button>
        {expanded && <div className="border-t border-gray-200 dark:border-gray-800 divide-y divide-gray-100 dark:divide-gray-800">{children}</div>}
      </div>
    </div>
  )
}
