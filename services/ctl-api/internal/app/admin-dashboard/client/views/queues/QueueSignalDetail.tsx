import { useQuery } from '@tanstack/react-query'
import { Link, useParams } from 'react-router'
import { useState } from 'react'
import { getQueueSignalDetail } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { JsonViewer } from '@/components/common/JsonViewer'
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

  const { data, isLoading, error } = useQuery({
    queryKey: ['queue-signal', queueId, signalId],
    queryFn: () => getQueueSignalDetail(queueId!, signalId!),
    enabled: !!queueId && !!signalId,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load signal'} />
  if (!data) return null

  const { signal, queue, workflow_info: wfInfo, signal_attrs: attrs, signals_ahead: signalsAhead = [], temporal_ui_url: temporalUIUrl } = data
  const status = getStatus(signal?.status)
  const statusHistory = signal?.status?.history || []
  const enqueuedAt = signal?.created_at
  const dequeuedAt = getMeta(signal?.status, 'dequeued_at')
  const executeStartedAt = getMeta(signal?.status, 'execute_started_at')
  const executeFinishedAt = getMeta(signal?.status, 'execute_finished_at')

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <div className="flex gap-2 text-xs text-gray-500">
        <Link to="/queues" className="text-primary-600 hover:text-primary-700">Queues</Link>
        <span>&rarr;</span>
        <Link to={`/queues/${queue?.id}`} className="text-primary-600 hover:text-primary-700">{truncateId(queue?.id)}</Link>
        <span>&rarr;</span>
        <span>Signal</span>
      </div>

      {/* Header */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <div className="flex flex-wrap items-center gap-2 mb-2">
          <h1 className="text-lg font-semibold">Signal</h1>
          <Badge>{signal?.type}</Badge>
          <Badge variant="status" status={status}>{status}</Badge>
          {temporalUIUrl && signal?.workflow?.id && signal?.workflow?.namespace && (
            <a href={`${temporalUIUrl}/namespaces/${signal.workflow.namespace}/workflows/${signal.workflow.id}`} target="_blank" rel="noopener noreferrer" className="text-xs text-primary-600 hover:text-primary-700">
              View in Temporal &rarr;
            </a>
          )}
          <Link to="/signal-catalog" className="text-xs text-primary-600 hover:text-primary-700">View catalog &rarr;</Link>
        </div>
        <div className="space-y-1 text-xs">
          <div><span className="text-gray-500 uppercase">Signal ID:</span> <span className="font-mono">{signal?.id}</span></div>
          <div><span className="text-gray-500 uppercase">Queue ID:</span> <Link to={`/queues/${signal?.queue_id}`} className="font-mono text-primary-600 hover:text-primary-700">{signal?.queue_id}</Link></div>
        </div>
      </div>

      {/* Signal attributes */}
      {attrs && (
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h2 className="text-sm font-semibold text-gray-900">Signal attributes</h2>
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
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900 mb-3">Timeline</h2>
        <div className="flex items-start justify-between">
          <TimelineStep label="Enqueued" value={enqueuedAt} active />
          <TimelineConnector duration={timeBetween(enqueuedAt, dequeuedAt)} />
          <TimelineStep label="Dequeued" value={dequeuedAt} active={!!dequeuedAt} />
          <TimelineConnector duration={timeBetween(dequeuedAt, executeStartedAt)} />
          <TimelineStep label="Execute started" value={executeStartedAt} active={!!executeStartedAt} />
          <TimelineConnector duration={timeBetween(executeStartedAt, executeFinishedAt)} />
          <TimelineStep label="Execute finished" value={executeFinishedAt} active={!!executeFinishedAt} />
        </div>
      </div>

      {/* Signals ahead */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Signals ahead</h2>
        {signalsAhead.length > 0 ? (
          <div className="mt-2 space-y-1">
            {signalsAhead.map((sig: any) => (
              <div key={sig.id} className="flex items-center gap-3 p-2 border border-gray-100 rounded text-xs">
                <Link to={`/queues/${queue?.id}/signals/${sig.id}`} className="font-mono text-primary-600 hover:text-primary-700 truncate flex-1">{sig.id}</Link>
                <Badge>{sig.type}</Badge>
                <Badge variant="status" status={getStatus(sig.status)}>{getStatus(sig.status)}</Badge>
                <span className="text-gray-400 whitespace-nowrap">{formatDate(sig.created_at)}</span>
              </div>
            ))}
          </div>
        ) : (
          <p className="mt-2 text-sm text-gray-500">No signals ahead</p>
        )}
      </div>

      {/* Executions */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Executions</h2>
        <div className="mt-2 flex items-center gap-4">
          <div>
            <p className="text-xs text-gray-500 uppercase">Execution count</p>
            <p className={`text-2xl font-bold font-mono ${signal?.execution_count > 1 ? 'text-orange-500' : ''}`}>
              {signal?.execution_count ?? 0}
            </p>
          </div>
        </div>
        {wfInfo && (
          <div className="mt-3 border-t border-gray-200 pt-3 space-y-2 text-xs">
            <div><span className="text-gray-500 uppercase w-32 inline-block">Status</span> <Badge variant="status" status={wfInfo.Status}>{wfInfo.Status}</Badge></div>
            {wfInfo.UpdateExecutions?.length > 0 && (
              <div><span className="text-gray-500 uppercase w-32 inline-block">Updates</span> <span className="font-mono">{wfInfo.UpdateExecutions.length}</span></div>
            )}
            {wfInfo.Activities?.length > 0 && (
              <div><span className="text-gray-500 uppercase w-32 inline-block">Activities</span> <span className="font-mono">{wfInfo.Activities.length}</span></div>
            )}
            {/* Failures */}
            {(wfInfo.Status === 'Failed' || wfInfo.Status === 'Timed Out') && (
              <>
                {wfInfo.UpdateExecutions?.filter((ue: any) => ue.Failure).map((ue: any, i: number) => (
                  <div key={i} className="mt-1">
                    <Badge variant="status" status="failed">{ue.Name}</Badge>
                    <pre className="mt-1 text-xs text-red-600 font-mono whitespace-pre-wrap bg-red-50 border border-red-200 rounded p-2">{ue.Failure}</pre>
                  </div>
                ))}
                {wfInfo.OrphanActivities?.filter((a: any) => a.Failure).map((a: any, i: number) => (
                  <div key={i} className="mt-1">
                    <Badge variant="status" status="failed">{a.Name}</Badge>
                    <pre className="mt-1 text-xs text-red-600 font-mono whitespace-pre-wrap bg-red-50 border border-red-200 rounded p-2">{a.Failure}</pre>
                  </div>
                ))}
              </>
            )}
          </div>
        )}
      </div>

      {/* Workflow activities */}
      {wfInfo?.Activities?.length > 0 && (
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h2 className="text-sm font-semibold text-gray-900">Activities ({wfInfo.Activities.length})</h2>
          <div className="mt-2 table-card">
            <table>
              <thead>
                <tr>
                  <th>Name</th><th>Status</th><th>Duration</th><th>Attempt</th><th>Failure</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {wfInfo.Activities.map((act: any, i: number) => (
                  <tr key={i}>
                    <td className="font-mono text-xs">{act.Name}</td>
                    <td><Badge variant="status" status={act.Status}>{act.Status}</Badge></td>
                    <td className="font-mono text-xs text-gray-500">{formatDuration(act.Duration)}</td>
                    <td className="text-xs text-gray-500">{act.Attempt}</td>
                    <td className="text-xs text-red-500 max-w-[200px] truncate">{act.Failure || '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Child workflows */}
      {wfInfo?.ChildWorkflows?.length > 0 && (
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h2 className="text-sm font-semibold text-gray-900">Child workflows ({wfInfo.ChildWorkflows.length})</h2>
          <div className="mt-2 table-card">
            <table>
              <thead>
                <tr>
                  <th>Type</th><th>Status</th><th>Namespace</th><th>Duration</th><th>Workflow ID</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {wfInfo.ChildWorkflows.map((cw: any, i: number) => (
                  <tr key={i}>
                    <td className="font-mono text-xs">{cw.WorkflowType}</td>
                    <td><Badge variant="status" status={cw.Status}>{cw.Status}</Badge></td>
                    <td className="text-xs text-gray-500">{cw.Namespace}</td>
                    <td className="font-mono text-xs text-gray-500">{formatDuration(cw.Duration)}</td>
                    <td className="font-mono text-xs">
                      {temporalUIUrl ? (
                        <a href={`${temporalUIUrl}/namespaces/${cw.Namespace}/workflows/${cw.WorkflowID}`} target="_blank" rel="noopener noreferrer" className="text-primary-600 hover:text-primary-700">{truncateId(cw.WorkflowID)}</a>
                      ) : truncateId(cw.WorkflowID)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Awaited signals */}
      {wfInfo?.AwaitedSignals?.length > 0 && (
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h2 className="text-sm font-semibold text-gray-900">Awaited signals ({wfInfo.AwaitedSignals.length})</h2>
          <div className="mt-2 table-card">
            <table>
              <thead>
                <tr>
                  <th>Signal ID</th><th>Status</th><th>Duration</th><th>Failure</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {wfInfo.AwaitedSignals.map((as: any, i: number) => (
                  <tr key={i}>
                    <td className="font-mono text-xs">
                      {as.QueueSignalID ? (
                        <Link to={`/queue-signals?search=${as.QueueSignalID}`} className="text-primary-600 hover:text-primary-700">{truncateId(as.QueueSignalID)}</Link>
                      ) : '-'}
                    </td>
                    <td><Badge variant="status" status={as.Status}>{as.Status}</Badge></td>
                    <td className="font-mono text-xs text-gray-500">{formatDuration(as.Duration)}</td>
                    <td className="text-xs text-red-500 max-w-[200px] truncate">{as.Failure || '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Update executions */}
      {wfInfo?.UpdateExecutions?.length > 0 && (
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h2 className="text-sm font-semibold text-gray-900">Update executions ({wfInfo.UpdateExecutions.length})</h2>
          <div className="mt-2 space-y-2">
            {wfInfo.UpdateExecutions.map((ue: any, i: number) => (
              <UpdateExecutionCard key={i} ue={ue} />
            ))}
          </div>
        </div>
      )}

      {/* Update handlers */}
      {wfInfo?.UpdateHandlers?.length > 0 && (
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h2 className="text-sm font-semibold text-gray-900">Update handlers</h2>
          <div className="mt-2 flex flex-wrap gap-2">
            {wfInfo.UpdateHandlers.map((h: string) => <Badge key={h}>{h}</Badge>)}
          </div>
        </div>
      )}

      {/* Signal info + Handler */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h2 className="text-sm font-semibold text-gray-900 mb-2">Signal info</h2>
          <div className="space-y-2 text-xs">
            <InfoRow label="Type" value={signal?.type} />
            <InfoRow label="Status" value={status} />
            <InfoRow label="Exec count" value={String(signal?.execution_count ?? 0)} highlight={signal?.execution_count > 1} />
            {signal?.status?.status_human_description && <InfoRow label="Description" value={signal.status.status_human_description} />}
            <InfoRow label="Created" value={formatDate(signal?.created_at)} />
            <InfoRow label="Updated" value={formatDate(signal?.updated_at)} />
            {signal?.status?.metadata && Object.entries(signal.status.metadata).map(([k, v]) => (
              <InfoRow key={k} label={k} value={String(v)} />
            ))}
          </div>
        </div>
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h2 className="text-sm font-semibold text-gray-900 mb-2">Handler</h2>
          <div className="space-y-2 text-xs">
            <InfoRow label="Owner type" value={signal?.owner_type} />
            <div className="flex items-start gap-3">
              <span className="text-gray-500 uppercase w-28 shrink-0">Owner ID</span>
              <span className="font-mono break-all">{signal?.owner_id}</span>
            </div>
            {signal?.emitter_id && (
              <div className="flex items-start gap-3">
                <span className="text-gray-500 uppercase w-28 shrink-0">Emitter</span>
                <Link to={`/queues/${queue?.id}/emitters/${signal.emitter_id}`} className="font-mono text-primary-600 hover:text-primary-700">{signal.emitter_id}</Link>
              </div>
            )}
            {signal?.workflow?.id && <InfoRow label="Workflow ID" value={signal.workflow.id} />}
            {signal?.workflow?.namespace && <InfoRow label="Namespace" value={signal.workflow.namespace} />}
          </div>
          <div className="mt-3 pt-3 border-t border-gray-200 flex flex-wrap gap-3 text-xs">
            {signal?.workflow?.id && signal?.workflow?.namespace && (
              <Link to={`/temporal-workflows?namespace=${signal.workflow.namespace}&workflow_id=${signal.workflow.id}`} className="text-primary-600 hover:text-primary-700">
                View handler workflow &rarr;
              </Link>
            )}
            {temporalUIUrl && signal?.workflow?.id && signal?.workflow?.namespace && (
              <a href={`${temporalUIUrl}/namespaces/${signal.workflow.namespace}/workflows/${signal.workflow.id}`} target="_blank" rel="noopener noreferrer" className="text-primary-600 hover:text-primary-700">
                View in Temporal UI &rarr;
              </a>
            )}
            <Link to={`/queue-signals?search=${signal?.owner_id}`} className="text-primary-600 hover:text-primary-700">
              Owner signals &rarr;
            </Link>
            {signal?.owner_type === 'install_workflow_steps' && (
              <Link to={`/workflows?search=${signal?.owner_id}`} className="text-primary-600 hover:text-primary-700">Step's workflow &rarr;</Link>
            )}
            {signal?.owner_type === 'install_workflows' && (
              <Link to={`/workflows?search=${signal?.owner_id}`} className="text-primary-600 hover:text-primary-700">Install workflow &rarr;</Link>
            )}
          </div>
        </div>
      </div>

      {/* Signal data */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Signal data</h2>
        <div className="mt-2">
          <JsonViewer data={signal?.signal || signal} />
        </div>
      </div>

      {/* Status history */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Status history</h2>
        <div className="mt-2 space-y-2">
          {/* Current */}
          <StatusHistoryEntry h={signal?.status} isCurrent />
          {/* History */}
          {statusHistory.map((h: any, i: number) => (
            <StatusHistoryEntry key={i} h={h} />
          ))}
        </div>
      </div>
    </div>
  )
}

function TimelineStep({ label, value, active }: { label: string; value?: string; active: boolean }) {
  return (
    <div className="flex flex-col items-center text-center min-w-0 shrink-0">
      <div className={`w-3 h-3 rounded-full mb-2 ${active ? 'bg-primary-500' : 'bg-gray-300'}`} />
      <p className="text-[10px] text-gray-500 uppercase font-medium">{label}</p>
      {value ? (
        <p className="text-[10px] font-mono mt-0.5 max-w-[140px] break-all">{formatDate(value)}</p>
      ) : (
        <p className="text-[10px] font-mono mt-0.5 text-gray-300">&mdash;</p>
      )}
    </div>
  )
}

function TimelineConnector({ duration }: { duration: string }) {
  return (
    <div className="flex flex-col items-center flex-1 min-w-[40px] pt-1">
      <div className="h-px w-full bg-gray-200 mt-1" />
      {duration && <p className="text-[10px] font-mono text-primary-500 mt-1">{duration}</p>}
    </div>
  )
}

function InfoRow({ label, value, highlight }: { label: string; value?: string; highlight?: boolean }) {
  return (
    <div className="flex items-start gap-3">
      <span className="text-gray-500 uppercase w-28 shrink-0">{label}</span>
      <span className={`font-mono break-all ${highlight ? 'text-orange-500 font-bold' : ''}`}>{value || '-'}</span>
    </div>
  )
}

function StatusHistoryEntry({ h, isCurrent }: { h: any; isCurrent?: boolean }) {
  if (!h) return null
  const status = getStatus(h)
  return (
    <div className="flex items-start gap-3 text-xs border-b border-gray-100 pb-2 last:border-0">
      <Badge variant="status" status={status}>{status}</Badge>
      <div className="flex-1 space-y-0.5">
        <div>
          {status}
          {isCurrent && <span className="text-primary-600 font-medium ml-1">(current)</span>}
          {h.status_human_description && <span className="text-gray-500 ml-1">— {h.status_human_description}</span>}
        </div>
        {h.created_at_ts > 0 && (
          <div className="text-gray-400 font-mono">{new Date(h.created_at_ts / 1000000).toISOString().replace('T', ' ').slice(0, 19)} UTC</div>
        )}
        {h.metadata && Object.keys(h.metadata).length > 0 && (
          <div className="flex flex-wrap gap-x-4 gap-y-0.5 text-gray-400">
            {Object.entries(h.metadata).map(([k, v]) => (
              <span key={k}><span className="text-gray-500">{k}:</span> {String(v)}</span>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

function UpdateExecutionCard({ ue }: { ue: any }) {
  const [expanded, setExpanded] = useState(false)
  return (
    <div className="border border-gray-200 rounded-md">
      <button onClick={() => setExpanded(!expanded)} className="w-full flex items-center justify-between px-3 py-2 text-left hover:bg-gray-50">
        <div className="flex items-center gap-2 text-xs">
          <Badge variant="status" status={ue.Status}>{ue.Status}</Badge>
          <span className="font-mono font-medium">{ue.Name}</span>
          <span className="text-gray-400">{formatDuration(ue.Duration)}</span>
        </div>
        <span className="text-gray-400 text-xs">{expanded ? '▾' : '▸'}</span>
      </button>
      {expanded && (
        <div className="border-t border-gray-200 px-3 py-2 text-xs space-y-2">
          <div><span className="text-gray-500">Update ID:</span> <span className="font-mono">{ue.UpdateID}</span></div>
          {ue.Input && <div><span className="text-gray-500">Input:</span> <pre className="mt-0.5 font-mono bg-gray-50 rounded p-2 overflow-x-auto max-h-32">{ue.Input}</pre></div>}
          {ue.Result && <div><span className="text-gray-500">Result:</span> <pre className="mt-0.5 font-mono bg-gray-50 rounded p-2 overflow-x-auto max-h-32">{ue.Result}</pre></div>}
          {ue.Failure && <div><span className="text-gray-500">Failure:</span> <pre className="mt-0.5 font-mono text-red-600 bg-red-50 rounded p-2 overflow-x-auto max-h-32">{ue.Failure}</pre></div>}
          {ue.Activities?.length > 0 && (
            <div>
              <p className="text-gray-500 mb-1">Activities ({ue.Activities.length}):</p>
              {ue.Activities.map((a: any, i: number) => (
                <div key={i} className="flex items-center gap-2 pl-2">
                  <Badge variant="status" status={a.Status}>{a.Status}</Badge>
                  <span className="font-mono">{a.Name}</span>
                  <span className="text-gray-400">{formatDuration(a.Duration)}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
