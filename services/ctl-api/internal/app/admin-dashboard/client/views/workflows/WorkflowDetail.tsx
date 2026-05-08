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
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  const remainMinutes = minutes % 60
  if (hours < 24) return remainMinutes > 0 ? `${hours}h${remainMinutes}m` : `${hours}h`
  const days = Math.floor(hours / 24)
  const remainHours = hours % 24
  return remainHours > 0 ? `${days}d${remainHours}h` : `${days}d`
}

// -- Step detail row --

function StepRow({ stepData }: { stepData: any }) {
  const [expanded, setExpanded] = useState(false)
  const step = stepData.step
  const status = getStatus(step?.status)
  const statusHistory = getStatusHistory(step?.status)

  const maxAutoRetries = step?.status?.metadata?.max_auto_retries
  return (
    <>
      <tr className="hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer" onClick={() => setExpanded(!expanded)}>
        <td className="text-xs text-gray-500 dark:text-gray-400">{step?.idx}</td>
        <td className="text-xs text-gray-900 dark:text-gray-100 max-w-[200px] truncate" title={step?.name}>{step?.name || '-'}</td>
        <td><Badge variant="status" status={status}>{status || '-'}</Badge></td>
        <td className="text-xs text-gray-500 dark:text-gray-400 font-mono">
          <span>{step?.execution_type || '-'}</span>
          {maxAutoRetries !== undefined && maxAutoRetries !== null && (
            <Badge className="ml-1">max retries: {String(maxAutoRetries)}</Badge>
          )}
        </td>
        <td className="text-xs text-gray-500 dark:text-gray-400">{formatTime(step?.started_at)}</td>
        <td className="text-xs text-gray-500 dark:text-gray-400">{formatDur(step?.execution_time)}</td>
        <td className="text-xs space-x-1">
          {stepData.step_signal_id && stepData.step_signal_queue_id && (
            <Link to={`/queues/${stepData.step_signal_queue_id}/signals/${stepData.step_signal_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300" onClick={(e) => e.stopPropagation()}>
              signal
            </Link>
          )}
          {!stepData.step_signal_id && (
            <Link to={`/queue-signals?search=${step?.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300" onClick={(e) => e.stopPropagation()}>
              search
            </Link>
          )}
        </td>
        <td className="text-xs text-gray-400 dark:text-gray-500">{expanded ? '▾' : '▸'}</td>
      </tr>
      {expanded && (
        <tr>
          <td colSpan={8} className="px-4 py-3">
            <div className="space-y-3 text-xs">
              {/* IDs */}
              <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
                <div><span className="text-gray-500 dark:text-gray-400">Step ID:</span> <span className="font-mono">{step?.id}</span></div>
                <div><span className="text-gray-500 dark:text-gray-400">Group:</span> g{step?.group_idx}r{step?.group_retry_idx}</div>
                <div><span className="text-gray-500 dark:text-gray-400">Target type:</span> {step?.step_target_type || '-'}</div>
                <div><span className="text-gray-500 dark:text-gray-400">Target ID:</span> <span className="font-mono">{truncateId(step?.step_target_id) || '-'}</span></div>
              </div>

              {/* Step flags */}
              <div className="flex gap-1">
                {step?.retryable && <Badge>retryable</Badge>}
                {step?.skippable && <Badge>skippable</Badge>}
                {step?.retried && <Badge variant="status" status="warning">retried</Badge>}
                {step?.result_directive && <Badge>{step.result_directive}</Badge>}
                {step?.timeout > 0 && <Badge>timeout: {formatNsDuration(step.timeout)}</Badge>}
              </div>

              {/* Step target */}
              {stepData.step_target && (
                <div>
                  <p className="font-semibold text-gray-700 dark:text-gray-300 mb-1">Step target</p>
                  <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
                    <div><span className="text-gray-500 dark:text-gray-400">Type:</span> {stepData.step_target.type}</div>
                    <div><span className="text-gray-500 dark:text-gray-400">ID:</span> <span className="font-mono">{stepData.step_target.id}</span></div>
                    <div><span className="text-gray-500 dark:text-gray-400">Status:</span> {stepData.step_target.status ? <Badge variant="status" status={stepData.step_target.status}>{stepData.step_target.status}</Badge> : '-'}</div>
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
                <div>
                  <p className="font-semibold text-gray-700 dark:text-gray-300 mb-1">Approval</p>
                  <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
                    <div><span className="text-gray-500 dark:text-gray-400">Type:</span> {step.approval.type || '-'}</div>
                    {step.approval.response && (
                      <div><span className="text-gray-500 dark:text-gray-400">Response:</span> {step.approval.response.type || step.approval.response.response || '-'}</div>
                    )}
                    {step.approval.note && (
                      <div className="col-span-2"><span className="text-gray-500 dark:text-gray-400">Note:</span> {step.approval.note}</div>
                    )}
                  </div>
                </div>
              )}

              {/* Status history */}
              {step?.status && (
                <div>
                  <p className="font-semibold text-gray-700 dark:text-gray-300 mb-1">Status history</p>
                  <StatusHistory status={step.status} maxCollapsed={3} />
                </div>
              )}

              {/* Metadata */}
              {step?.metadata && Object.keys(step.metadata).length > 0 && (
                <div>
                  <p className="font-semibold text-gray-700 dark:text-gray-300 mb-1">Metadata</p>
                  <div className="grid grid-cols-2 gap-1">
                    {Object.entries(step.metadata).map(([k, v]) => (
                      <div key={k}><span className="text-gray-500 dark:text-gray-400">{k}:</span> <span className="font-mono">{String(v)}</span></div>
                    ))}
                  </div>
                </div>
              )}

              {/* Queue signal JSON */}
              {stepData.queue_signal_json && (
                <div>
                  <p className="font-semibold text-gray-700 dark:text-gray-300 mb-1">Queue signal data</p>
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

function StepGroupSection({ group }: { group: any }) {
  const [expanded, setExpanded] = useState(true)
  const g = group.group
  const status = getStatus(g?.status)
  const steps = group.steps || []
  const groupSignal = g?.queue_signal

  return (
    <div className="rounded-lg border border-gray-200 dark:border-gray-800">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center justify-between px-4 py-3 text-left hover:bg-gray-50 dark:hover:bg-gray-800"
      >
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-xs font-semibold text-gray-500 dark:text-gray-400">g{g?.group_idx}</span>
          <span className="text-sm font-medium text-gray-900 dark:text-gray-100">{g?.name || `Group ${g?.group_idx}`}</span>
          <Badge>{g?.parallel ? 'parallel' : 'sequential'}</Badge>
          {status && <Badge variant="status" status={status}>{status}</Badge>}
          {g?.result_directive && <Badge>{g.result_directive}</Badge>}
          {g?.timeout > 0 && <Badge>timeout: {formatNsDuration(g.timeout)}</Badge>}
          <span className="text-xs text-gray-400 dark:text-gray-500">{steps.length} steps</span>
        </div>
        <div className="flex items-center gap-2">
          {groupSignal && (
            <Link
              to={`/queues/${groupSignal.queue_id}/signals/${groupSignal.id}`}
              className="text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300"
              onClick={(e) => e.stopPropagation()}
            >
              group signal &rarr;
            </Link>
          )}
          {!groupSignal && g?.id && (
            <Link
              to={`/queue-signals?search=${g.id}`}
              className="text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300"
              onClick={(e) => e.stopPropagation()}
            >
              search signal
            </Link>
          )}
          <span className="text-gray-400 dark:text-gray-500">{expanded ? '▾' : '▸'}</span>
        </div>
      </button>
      {expanded && steps.length > 0 && (
        <div className="border-t border-gray-200 dark:border-gray-800 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-100 dark:divide-gray-800">
            <thead className="">
              <tr>
                <th className="px-4 py-2 text-left text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase w-10">#</th>
                <th className="px-4 py-2 text-left text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase">Name</th>
                <th className="px-4 py-2 text-left text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase">Status</th>
                <th className="px-4 py-2 text-left text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase">Exec type</th>
                <th className="px-4 py-2 text-left text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase">Started</th>
                <th className="px-4 py-2 text-left text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase">Duration</th>
                <th className="px-4 py-2 text-left text-[10px] font-medium text-gray-500 dark:text-gray-400 uppercase">Signal</th>
                <th className="px-4 py-2 w-6"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
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
  const wfStatusHistory = getStatusHistory(wf?.status)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <div className="flex items-center gap-2">
          <h1 className="page-heading font-mono text-base">{wf?.id}</h1>
        </div>
        <div className="mt-2 flex flex-wrap items-center gap-2">
          <Badge>{wf?.type}</Badge>
          <Badge variant="status" status={wfStatus}>{wfStatus || '-'}</Badge>
          {wf?.result_directive && <Badge>{wf.result_directive}</Badge>}
        </div>
        <div className="mt-2 text-sm text-gray-500 dark:text-gray-400">
          Owner: <Link to={`/installs/${wf?.owner_id}`} className="font-mono text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">{wf?.owner_id}</Link>
          <span className="text-gray-400 dark:text-gray-500 ml-1">({wf?.owner_type})</span>
          {wf?.created_by?.email && (
            <span className="ml-3">by <Link to={`/accounts/${wf.created_by_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">{wf.created_by.email}</Link></span>
          )}
        </div>
        <div className="mt-1 flex gap-3 text-xs">
          {wfSignal ? (
            <>
              <Link to={`/queues/${wfSignal.queue_id}/signals/${wfSignal.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">Workflow signal &rarr;</Link>
              <Link to={`/queues/${wfSignal.queue_id}/signals/${wfSignal.id}/graph`} className="inline-flex items-center rounded bg-primary-50 dark:bg-primary-950 border border-primary-200 dark:border-primary-800 px-1.5 py-0.5 text-[10px] font-medium text-primary-600 dark:text-primary-400 hover:bg-primary-100 dark:hover:bg-primary-900">graph</Link>
            </>
          ) : (
            <Link to={`/queue-signals?search=${wf?.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">Search workflow signals &rarr;</Link>
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
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Generate steps signal</h2>
          <div className="mt-2 grid grid-cols-2 gap-2 sm:grid-cols-4 text-sm">
            <div>
              <span className="text-gray-500 dark:text-gray-400 text-xs">Signal ID:</span>
              <Link to={`/queues/${genSignal.queue_id}/signals/${genSignal.id}`} className="ml-1 font-mono text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
                {truncateId(genSignal.id)}
              </Link>
            </div>
            <div>
              <span className="text-gray-500 dark:text-gray-400 text-xs">Queue:</span>
              <Link to={`/queues/${genSignal.queue_id}`} className="ml-1 font-mono text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
                {truncateId(genSignal.queue_id)}
              </Link>
            </div>
            <div>
              <span className="text-gray-500 dark:text-gray-400 text-xs">Type:</span>
              <span className="ml-1 font-mono text-xs">{genSignal.type}</span>
            </div>
            <div>
              <span className="text-gray-500 dark:text-gray-400 text-xs">Status:</span>
              <Badge variant="status" status={getStatus(genSignal.status)} className="ml-1">{getStatus(genSignal.status)}</Badge>
            </div>
          </div>
        </div>
      )}

      {/* Workflow status history */}
      {wf?.status && (
        <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Status history</h2>
          <div className="mt-2">
            <StatusHistory status={wf.status} maxCollapsed={5} />
          </div>
        </div>
      )}

      {/* Step Groups */}
      <div>
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-2">Step groups ({groups.length})</h2>
        {groups.length > 0 ? (
          <div className="space-y-3">
            {groups.map((group: any, i: number) => (
              <StepGroupSection key={group.group?.id || i} group={group} />
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
      <p className="text-[11px] text-gray-500 dark:text-gray-400 uppercase tracking-wider">{label}</p>
      <p className="mt-0.5 text-sm font-mono text-gray-900 dark:text-gray-100">{value}</p>
    </div>
  )
}
