import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router'
import { getWorkflowDetail } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { JsonViewer } from '@/components/common/JsonViewer'
import { StatusHistory } from '@/components/common/StatusHistory'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, formatDuration, truncateId } from '@/utils/format'

function getStatus(s: any): string {
  if (!s) return ''
  if (typeof s === 'string') return s
  if (typeof s === 'object' && s.status) return String(s.status)
  return String(s)
}

function getStatusDescription(s: any): string {
  if (!s || typeof s !== 'object') return ''
  return s.status_human_description || ''
}

function getStatusHistory(s: any): any[] {
  if (!s || typeof s !== 'object') return []
  return s.history || []
}

function formatTime(t: string | undefined): string {
  if (!t || t === '0001-01-01T00:00:00Z') return '-'
  return formatDate(t)
}

function formatDur(ns: number | undefined): string {
  if (!ns || ns <= 0) return '-'
  return formatDuration(ns)
}

function formatNsDuration(ns: number): string {
  const seconds = Math.floor(ns / 1e9)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ${seconds % 60}s`
  const hours = Math.floor(minutes / 60)
  const remainMinutes = minutes % 60
  if (hours < 24) return remainMinutes > 0 ? `${hours}h ${remainMinutes}m` : `${hours}h`
  const days = Math.floor(hours / 24)
  const remainHours = hours % 24
  return remainHours > 0 ? `${days}d ${remainHours}h` : `${days}d`
}

function directiveColor(directive: string | undefined): string {
  if (!directive) return ''
  switch (directive) {
    case 'continue': return 'bg-green-50 text-green-700 border-green-200 dark:bg-green-900/30 dark:text-green-300 dark:border-green-800'
    case 'stop': return 'bg-red-50 text-red-700 border-red-200 dark:bg-red-900/30 dark:text-red-300 dark:border-red-800'
    case 'retry': case 'retry-group': return 'bg-yellow-50 text-yellow-700 border-yellow-200 dark:bg-yellow-900/30 dark:text-yellow-300 dark:border-yellow-800'
    case 'skip-group': return 'bg-gray-100 text-gray-600 border-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700'
    case 'await-approval': return 'bg-purple-50 text-purple-700 border-purple-200 dark:bg-purple-900/30 dark:text-purple-300 dark:border-purple-800'
    default: return 'bg-gray-100 text-gray-600 border-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700'
  }
}

function DirectiveBadge({ directive }: { directive: string | undefined }) {
  if (!directive) return null
  return (
    <span className={`inline-flex items-center rounded-full border px-2 py-0.5 text-[11px] font-semibold leading-4 ${directiveColor(directive)}`}>
      {directive}
    </span>
  )
}

function SignalLink({ queueId, signalId, label, className }: { queueId?: string; signalId?: string; label: string; className?: string }) {
  if (!queueId || !signalId) return null
  return (
    <Link
      to={`/queues/${queueId}/signals/${signalId}`}
      className={`text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 ${className || ''}`}
      onClick={(e) => e.stopPropagation()}
    >
      {label} &rarr;
    </Link>
  )
}

// -- Step detail row --

function StepRow({ stepData }: { stepData: any }) {
  const [expanded, setExpanded] = useState(false)
  const step = stepData.step
  const status = getStatus(step?.status)
  const maxAutoRetries = step?.status?.metadata?.max_auto_retries

  return (
    <>
      <tr className="hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer transition-colors" onClick={() => setExpanded(!expanded)}>
        <td className="px-3 py-2 text-xs text-gray-400 dark:text-gray-500 font-mono text-right tabular-nums">{step?.idx}</td>
        <td className="px-3 py-2">
          <span className="text-xs font-medium text-gray-900 dark:text-gray-100">{step?.name || '-'}</span>
        </td>
        <td className="px-3 py-2"><Badge variant="status" status={status}>{status || '-'}</Badge></td>
        <td className="px-3 py-2"><DirectiveBadge directive={step?.result_directive} /></td>
        <td className="px-3 py-2 text-xs text-gray-500 dark:text-gray-400 font-mono">{step?.execution_type || '-'}</td>
        <td className="px-3 py-2 text-xs text-gray-500 dark:text-gray-400 font-mono tabular-nums">{formatDur(step?.execution_time)}</td>
        <td className="px-3 py-2 text-xs space-x-1">
          {maxAutoRetries !== undefined && maxAutoRetries !== null && (
            <Badge className="mr-1">retries: {String(maxAutoRetries)}</Badge>
          )}
          {step?.retried && <Badge variant="status" status="error">retried</Badge>}
        </td>
        <td className="px-3 py-2 text-xs">
          {stepData.step_signal_id && stepData.step_signal_queue_id ? (
            <SignalLink queueId={stepData.step_signal_queue_id} signalId={stepData.step_signal_id} label="signal" />
          ) : (
            <Link to={`/queue-signals?search=${step?.id}`} className="text-xs text-gray-400 dark:text-gray-500 hover:text-primary-600 dark:hover:text-primary-400" onClick={(e) => e.stopPropagation()}>
              search
            </Link>
          )}
        </td>
        <td className="px-1 py-2 text-xs text-gray-400 dark:text-gray-500">{expanded ? '▾' : '▸'}</td>
      </tr>
      {expanded && (
        <tr className="bg-gray-50/50 dark:bg-gray-900/50">
          <td colSpan={9} className="px-6 py-4">
            <div className="space-y-4 text-xs">
              {/* IDs row */}
              <div className="grid grid-cols-2 gap-x-6 gap-y-2 sm:grid-cols-4">
                <div><span className="text-gray-400 dark:text-gray-500">ID</span><br/><span className="font-mono text-gray-700 dark:text-gray-300">{step?.id}</span></div>
                <div><span className="text-gray-400 dark:text-gray-500">Group</span><br/>g{step?.group_idx} r{step?.group_retry_idx}</div>
                <div><span className="text-gray-400 dark:text-gray-500">Target type</span><br/>{step?.step_target_type || '-'}</div>
                <div><span className="text-gray-400 dark:text-gray-500">Target ID</span><br/><span className="font-mono">{truncateId(step?.step_target_id) || '-'}</span></div>
              </div>

              {/* Flags */}
              <div className="flex flex-wrap gap-1.5">
                {step?.retryable && <Badge>retryable</Badge>}
                {step?.skippable && <Badge>skippable</Badge>}
                {step?.eager_execution && <Badge>eager</Badge>}
                {step?.timeout > 0 && <Badge>timeout: {formatNsDuration(step.timeout)}</Badge>}
                <DirectiveBadge directive={step?.result_directive} />
              </div>

              {/* Times */}
              <div className="grid grid-cols-2 gap-x-6 gap-y-1 sm:grid-cols-3">
                <div><span className="text-gray-400 dark:text-gray-500">Started</span> <span className="font-mono ml-1">{formatTime(step?.started_at)}</span></div>
                <div><span className="text-gray-400 dark:text-gray-500">Finished</span> <span className="font-mono ml-1">{formatTime(step?.finished_at)}</span></div>
                <div><span className="text-gray-400 dark:text-gray-500">Duration</span> <span className="font-mono ml-1">{formatDur(step?.execution_time)}</span></div>
              </div>

              {/* Step target */}
              {stepData.step_target && (
                <div className="rounded-md border border-gray-200 dark:border-gray-700 p-3">
                  <p className="text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500 mb-2">Step target</p>
                  <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
                    <div><span className="text-gray-400 dark:text-gray-500">Type</span><br/>{stepData.step_target.type}</div>
                    <div><span className="text-gray-400 dark:text-gray-500">ID</span><br/><span className="font-mono">{truncateId(stepData.step_target.id)}</span></div>
                    <div><span className="text-gray-400 dark:text-gray-500">Status</span><br/>{stepData.step_target.status ? <Badge variant="status" status={stepData.step_target.status}>{stepData.step_target.status}</Badge> : '-'}</div>
                    {stepData.step_target.log_stream_id && (
                      <div>
                        <Link to={`/log-streams/${stepData.step_target.log_stream_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
                          View logs &rarr;
                        </Link>
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* Approval */}
              {step?.approval && (
                <div className="rounded-md border border-purple-200 dark:border-purple-800 p-3">
                  <p className="text-[10px] font-semibold uppercase tracking-wider text-purple-400 dark:text-purple-500 mb-2">Approval</p>
                  <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
                    <div><span className="text-gray-400 dark:text-gray-500">Type</span><br/>{step.approval.type || '-'}</div>
                    {step.approval.response && (
                      <div><span className="text-gray-400 dark:text-gray-500">Response</span><br/>{step.approval.response.type || step.approval.response.response || '-'}</div>
                    )}
                    {step.approval.note && (
                      <div className="col-span-2"><span className="text-gray-400 dark:text-gray-500">Note</span><br/>{step.approval.note}</div>
                    )}
                  </div>
                </div>
              )}

              {/* Status metadata */}
              {step?.status?.metadata && Object.keys(step.status.metadata).length > 0 && (
                <div>
                  <p className="text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500 mb-2">Status metadata</p>
                  <div className="grid grid-cols-2 gap-1 sm:grid-cols-3">
                    {Object.entries(step.status.metadata).map(([k, v]) => (
                      <div key={k}><span className="text-gray-400 dark:text-gray-500">{k}:</span> <span className="font-mono">{String(v)}</span></div>
                    ))}
                  </div>
                </div>
              )}

              {/* Status history */}
              {step?.status && (
                <div>
                  <p className="text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500 mb-2">Status history</p>
                  <StatusHistory status={step.status} maxCollapsed={3} />
                </div>
              )}

              {/* Metadata */}
              {step?.metadata && Object.keys(step.metadata).length > 0 && (
                <div>
                  <p className="text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500 mb-2">Metadata</p>
                  <div className="grid grid-cols-2 gap-1 sm:grid-cols-3">
                    {Object.entries(step.metadata).map(([k, v]) => (
                      <div key={k}><span className="text-gray-400 dark:text-gray-500">{k}:</span> <span className="font-mono">{String(v)}</span></div>
                    ))}
                  </div>
                </div>
              )}

              {/* Queue signal JSON */}
              {stepData.queue_signal_json && (
                <div>
                  <p className="text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500 mb-2">Queue signal data</p>
                  <JsonViewer data={stepData.queue_signal_json} collapsed />
                </div>
              )}
            </div>
          </td>
        </tr>
      )}
    </>
  )
}

// -- Step Group section --

function StepGroupSection({ group, defaultExpanded }: { group: any; defaultExpanded: boolean }) {
  const [expanded, setExpanded] = useState(defaultExpanded)
  const g = group.group
  const status = getStatus(g?.status)
  const statusDesc = getStatusDescription(g?.status)
  const steps = group.steps || []
  const groupSignal = g?.queue_signal

  const hasErrors = steps.some((sd: any) => {
    const s = getStatus(sd.step?.status)
    return s === 'error' || s === 'failed'
  })

  return (
    <div className={`rounded-lg border ${hasErrors ? 'border-red-300 dark:border-red-800' : 'border-gray-200 dark:border-gray-800'}`}>
      {/* Group header */}
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center justify-between px-4 py-3 text-left hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
      >
        <div className="flex items-center gap-2 flex-wrap">
          <span className="inline-flex items-center justify-center rounded bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-300 text-[10px] font-bold px-1.5 py-0.5 tabular-nums">g{g?.group_idx}</span>
          <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">{g?.name || `Group ${g?.group_idx}`}</span>
          <Badge>{g?.parallel ? 'parallel' : 'sequential'}</Badge>
          <Badge variant="status" status={status}>{status || '-'}</Badge>
          <DirectiveBadge directive={g?.result_directive} />
          {g?.timeout > 0 && <span className="text-[10px] text-gray-400 dark:text-gray-500 font-mono">timeout: {formatNsDuration(g.timeout)}</span>}
          <span className="text-[10px] text-gray-400 dark:text-gray-500">{steps.length} step{steps.length !== 1 ? 's' : ''}</span>
        </div>
        <div className="flex items-center gap-3">
          {groupSignal && <SignalLink queueId={groupSignal.queue_id} signalId={groupSignal.id} label="signal" />}
          {!groupSignal && g?.id && (
            <Link
              to={`/queue-signals?search=${g.id}`}
              className="text-xs text-gray-400 dark:text-gray-500 hover:text-primary-600 dark:hover:text-primary-400"
              onClick={(e) => e.stopPropagation()}
            >
              search
            </Link>
          )}
          <span className="text-gray-400 dark:text-gray-500 text-sm">{expanded ? '▾' : '▸'}</span>
        </div>
      </button>

      {/* Status description + metadata */}
      {expanded && (statusDesc || (g?.status?.metadata && Object.keys(g.status.metadata).length > 0)) && (
        <div className="px-4 pb-2 -mt-1 space-y-1.5">
          {statusDesc && <p className="text-xs text-gray-500 dark:text-gray-400 italic">{statusDesc}</p>}
          {g?.status?.metadata && Object.keys(g.status.metadata).length > 0 && (
            <div className="flex flex-wrap gap-x-4 gap-y-0.5 text-xs">
              {Object.entries(g.status.metadata).map(([k, v]) => (
                <span key={k}><span className="text-gray-400 dark:text-gray-500">{k}:</span> <span className="font-mono text-gray-600 dark:text-gray-300">{String(v)}</span></span>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Steps table */}
      {expanded && steps.length > 0 && (
        <div className="border-t border-gray-200 dark:border-gray-800 overflow-x-auto">
          <table className="min-w-full">
            <thead>
              <tr className="border-b border-gray-100 dark:border-gray-800">
                <th className="px-3 py-2 text-left text-[10px] font-medium text-gray-400 dark:text-gray-500 uppercase w-10">#</th>
                <th className="px-3 py-2 text-left text-[10px] font-medium text-gray-400 dark:text-gray-500 uppercase">Name</th>
                <th className="px-3 py-2 text-left text-[10px] font-medium text-gray-400 dark:text-gray-500 uppercase w-24">Status</th>
                <th className="px-3 py-2 text-left text-[10px] font-medium text-gray-400 dark:text-gray-500 uppercase w-28">Directive</th>
                <th className="px-3 py-2 text-left text-[10px] font-medium text-gray-400 dark:text-gray-500 uppercase w-20">Exec</th>
                <th className="px-3 py-2 text-left text-[10px] font-medium text-gray-400 dark:text-gray-500 uppercase w-20">Duration</th>
                <th className="px-3 py-2 text-left text-[10px] font-medium text-gray-400 dark:text-gray-500 uppercase w-24">Flags</th>
                <th className="px-3 py-2 text-left text-[10px] font-medium text-gray-400 dark:text-gray-500 uppercase w-16">Signal</th>
                <th className="px-1 py-2 w-6"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-800/50">
              {steps.map((sd: any) => (
                <StepRow key={sd.step?.id} stepData={sd} />
              ))}
            </tbody>
          </table>
        </div>
      )}
      {expanded && steps.length === 0 && (
        <div className="border-t border-gray-200 dark:border-gray-800 px-4 py-4 text-sm text-gray-500 dark:text-gray-400">No steps in this group</div>
      )}
    </div>
  )
}

// -- Main page --

export const WorkflowDetail = () => {
  const { workflowId } = useParams<{ workflowId: string }>()

  const { data, isLoading, error } = useQuery({
    queryKey: ['workflow', workflowId],
    queryFn: () => getWorkflowDetail(workflowId!),
    enabled: !!workflowId,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load workflow'} />
  if (!data) return null

  const wf = data.workflow
  const groups = data.group_details || []
  const genSignal = data.generate_steps_signal
  const wfSignal = data.workflow_signal
  const wfStatus = getStatus(wf?.status)
  const wfStatusDesc = getStatusDescription(wf?.status)

  // Auto-expand groups that are in-progress or have errors
  const shouldExpand = (g: any) => {
    const s = getStatus(g?.group?.status)
    if (s === 'in-progress' || s === 'error' || s === 'failed') return true
    const steps = g?.steps || []
    return steps.some((sd: any) => {
      const ss = getStatus(sd.step?.status)
      return ss === 'error' || ss === 'failed' || ss === 'in-progress'
    })
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <div className="flex items-start justify-between">
          <div>
            <p className="text-[10px] font-medium uppercase tracking-wider text-gray-400 dark:text-gray-500">Workflow</p>
            <h1 className="font-mono text-sm text-gray-900 dark:text-gray-100 mt-0.5">{wf?.id}</h1>
          </div>
          <div className="flex items-center gap-2">
            <Badge>{wf?.type}</Badge>
            <Badge variant="status" status={wfStatus}>{wfStatus || '-'}</Badge>
            <DirectiveBadge directive={wf?.result_directive} />
          </div>
        </div>

        {wfStatusDesc && (
          <p className="mt-2 text-xs text-gray-500 dark:text-gray-400 italic">{wfStatusDesc}</p>
        )}

        <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-gray-500 dark:text-gray-400">
          <span>
            Owner: <Link to={`/installs/${wf?.owner_id}`} className="font-mono text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">{truncateId(wf?.owner_id)}</Link>
            <span className="text-gray-400 dark:text-gray-500 ml-1">({wf?.owner_type})</span>
          </span>
          {wf?.created_by?.email && (
            <span>
              by <Link to={`/accounts/${wf.created_by_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">{wf.created_by.email}</Link>
            </span>
          )}
          {wf?.role && <span>role: <span className="font-mono">{wf.role}</span></span>}
          {wf?.plan_only && <Badge>plan-only</Badge>}
        </div>

        <div className="mt-3 flex gap-3 text-xs">
          {wfSignal ? (
            <>
              <SignalLink queueId={wfSignal.queue_id} signalId={wfSignal.id} label="Workflow signal" />
              <Link to={`/queues/${wfSignal.queue_id}/signals/${wfSignal.id}/graph`} className="inline-flex items-center rounded bg-primary-50 dark:bg-primary-950 border border-primary-200 dark:border-primary-800 px-1.5 py-0.5 text-[10px] font-medium text-primary-600 dark:text-primary-400 hover:bg-primary-100 dark:hover:bg-primary-900">graph</Link>
            </>
          ) : (
            <Link to={`/queue-signals?search=${wf?.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">Search signals &rarr;</Link>
          )}
        </div>
      </div>

      {/* Timeline */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <TimelineCard label="Created" value={formatTime(wf?.created_at)} />
        <TimelineCard label="Started" value={formatTime(wf?.started_at)} />
        <TimelineCard label="Finished" value={formatTime(wf?.finished_at)} />
        <TimelineCard label="Duration" value={formatDur(wf?.execution_time)} />
      </div>

      {/* Generate steps signal */}
      {genSignal && (
        <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
          <p className="text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500 mb-2">Generate steps signal</p>
          <div className="flex flex-wrap items-center gap-3 text-xs">
            <SignalLink queueId={genSignal.queue_id} signalId={genSignal.id} label={truncateId(genSignal.id)} />
            <span className="font-mono text-gray-500 dark:text-gray-400">{genSignal.type}</span>
            <Badge variant="status" status={getStatus(genSignal.status)}>{getStatus(genSignal.status)}</Badge>
          </div>
        </div>
      )}

      {/* Workflow status history */}
      {wf?.status && (
        <details className="rounded-lg border border-gray-200 dark:border-gray-800">
          <summary className="px-4 py-3 text-sm font-semibold text-gray-900 dark:text-gray-100 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
            Status history
          </summary>
          <div className="px-4 pb-4">
            <StatusHistory status={wf.status} maxCollapsed={10} />
          </div>
        </details>
      )}

      {/* Step Groups */}
      <div>
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Step groups <span className="text-gray-400 dark:text-gray-500 font-normal">({groups.length})</span></h2>
        </div>
        {groups.length > 0 ? (
          <div className="space-y-2">
            {groups.map((group: any, i: number) => (
              <StepGroupSection key={group.group?.id || i} group={group} defaultExpanded={shouldExpand(group)} />
            ))}
          </div>
        ) : (
          <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4 text-sm text-gray-500 dark:text-gray-400">
            No steps recorded
          </div>
        )}
      </div>
    </div>
  )
}

function TimelineCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-3">
      <p className="text-[10px] font-medium uppercase tracking-wider text-gray-400 dark:text-gray-500">{label}</p>
      <p className="mt-1 text-sm font-mono text-gray-900 dark:text-gray-100">{value}</p>
    </div>
  )
}
