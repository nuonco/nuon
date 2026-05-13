import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useSearchParams } from 'react-router'
import { getTemporalWorkflows } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { JsonViewer } from '@/components/common/JsonViewer'
import { TemporalWorkflowCard } from '@/components/common/TemporalWorkflowCard'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDuration, truncateId } from '@/utils/format'

export const TemporalWorkflows = () => {
  const [searchParams] = useSearchParams()
  const [workflowId, setWorkflowId] = useState(searchParams.get('workflow_id') || '')
  const [namespace, setNamespace] = useState(searchParams.get('namespace') || '')
  const [submitted, setSubmitted] = useState(!!(searchParams.get('workflow_id') && searchParams.get('namespace')))

  const { data, isLoading, error } = useQuery({
    queryKey: ['temporal-workflows', workflowId, namespace],
    queryFn: () =>
      getTemporalWorkflows({
        workflow_id: workflowId || undefined,
        namespace: namespace || undefined,
      }),
    enabled: submitted && !!workflowId && !!namespace,
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitted(true)
  }

  const wfInfo = data?.workflow_info
  const temporalUIUrl = data?.temporal_ui_url

  return (
    <div className="space-y-6">
      <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">Temporal Workflows</h1>

      {/* Search form */}
      <form onSubmit={handleSubmit} className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Workflow ID</label>
            <input
              type="text"
              value={workflowId}
              onChange={(e) => { setWorkflowId(e.target.value); setSubmitted(false) }}
              placeholder="Workflow ID..."
              className="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 focus:ring-2 focus:ring-primary-600 dark:focus:ring-primary-500"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Namespace</label>
            <input
              type="text"
              value={namespace}
              onChange={(e) => { setNamespace(e.target.value); setSubmitted(false) }}
              placeholder="Namespace..."
              className="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 focus:ring-2 focus:ring-primary-600 dark:focus:ring-primary-500"
            />
          </div>
        </div>
        <div className="mt-3">
          <button
            type="submit"
            disabled={!workflowId || !namespace}
            className="rounded-md bg-primary-600 dark:bg-primary-500 px-4 py-1.5 text-sm font-medium text-white hover:bg-primary-700 dark:hover:bg-primary-600 disabled:opacity-50"
          >
            Search
          </button>
        </div>
      </form>

      {submitted && isLoading && <LoadingSpinner />}
      {submitted && error && <ErrorMessage message={(error as Error).message || 'Failed to load workflow'} />}

      {/* Workflow header */}
      {submitted && wfInfo && (
        <>
          <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
            <div className="flex flex-wrap items-center gap-2 mb-3">
              <h2 className="text-lg font-semibold">Workflow</h2>
              <Badge variant="status" status={wfInfo.status}>{wfInfo.status}</Badge>
            </div>
            <div className="space-y-1 text-xs">
              <div className="flex items-center gap-2">
                <span className="text-gray-500 dark:text-gray-400 uppercase w-28">Workflow ID</span>
                <span className="font-mono break-all select-all">{data?.workflow_id}</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-gray-500 dark:text-gray-400 uppercase w-28">Namespace</span>
                <span className="font-mono">{data?.namespace}</span>
              </div>
            </div>

            {/* Execution stats */}
            <div className="mt-4 pt-3 border-t border-gray-200 dark:border-gray-800 flex flex-wrap gap-6 text-xs">
              <div>
                <p className="text-gray-500 dark:text-gray-400 uppercase">Status</p>
                <p className="font-mono font-medium mt-0.5">{wfInfo.status}</p>
              </div>
              {wfInfo.update_executions?.length > 0 && (
                <div>
                  <p className="text-gray-500 dark:text-gray-400 uppercase">Updates</p>
                  <p className="font-mono font-medium mt-0.5">{wfInfo.update_executions.length}</p>
                </div>
              )}
              {(wfInfo.activities?.length > 0 || wfInfo.orphan_activities?.length > 0) && (
                <div>
                  <p className="text-gray-500 dark:text-gray-400 uppercase">Activities</p>
                  <p className="font-mono font-medium mt-0.5">
                    {(wfInfo.activities?.length || 0) + (wfInfo.orphan_activities?.length || 0)}
                  </p>
                </div>
              )}
              {wfInfo.child_workflows?.length > 0 && (
                <div>
                  <p className="text-gray-500 dark:text-gray-400 uppercase">Child Workflows</p>
                  <p className="font-mono font-medium mt-0.5">{wfInfo.child_workflows.length}</p>
                </div>
              )}
              {wfInfo.awaited_signals?.length > 0 && (
                <div>
                  <p className="text-gray-500 dark:text-gray-400 uppercase">Awaited Signals</p>
                  <p className="font-mono font-medium mt-0.5">{wfInfo.awaited_signals.length}</p>
                </div>
              )}
            </div>
          </div>

          {/* Temporal workflow stats card */}
          {temporalUIUrl && data?.namespace && data?.workflow_id && (
            <TemporalWorkflowCard
              temporalUIUrl={temporalUIUrl}
              namespace={data.namespace}
              workflowId={data.workflow_id}
            />
          )}

          {/* Failures section */}
          {(wfInfo.status === 'Failed' || wfInfo.status === 'Timed Out') && (
            <div className="rounded-lg border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/30 p-4">
              <h2 className="text-sm font-semibold text-red-800 dark:text-red-200 mb-2">Failures</h2>
              <div className="space-y-2">
                {wfInfo.update_executions?.filter((ue: any) => ue.failure).map((ue: any, i: number) => (
                  <div key={`ue-${i}`}>
                    <Badge variant="status" status="failed">{ue.name}</Badge>
                    <pre className="mt-1 text-xs text-red-600 dark:text-red-400 font-mono whitespace-pre-wrap border border-red-200 dark:border-red-800 rounded p-2">{ue.failure}</pre>
                  </div>
                ))}
                {wfInfo.orphan_activities?.filter((a: any) => a.failure).map((a: any, i: number) => (
                  <div key={`oa-${i}`}>
                    <Badge variant="status" status="failed">{a.name}</Badge>
                    <pre className="mt-1 text-xs text-red-600 dark:text-red-400 font-mono whitespace-pre-wrap border border-red-200 dark:border-red-800 rounded p-2">{a.failure}</pre>
                  </div>
                ))}
                {wfInfo.activities?.filter((a: any) => a.failure).map((a: any, i: number) => (
                  <div key={`a-${i}`}>
                    <Badge variant="status" status="failed">{a.name}</Badge>
                    <pre className="mt-1 text-xs text-red-600 dark:text-red-400 font-mono whitespace-pre-wrap border border-red-200 dark:border-red-800 rounded p-2">{a.failure}</pre>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Activities */}
          {wfInfo.activities?.length > 0 && (
            <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
              <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Activities ({wfInfo.activities.length})</h2>
              <div className="mt-2 table-card">
                <table>
                  <thead>
                    <tr>
                      <th>Name</th><th>Status</th><th>Duration</th><th>Attempt</th><th>Failure</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
                    {wfInfo.activities.map((act: any, i: number) => (
                      <tr key={i}>
                        <td className="font-mono text-xs">{act.name}</td>
                        <td><Badge variant="status" status={act.status}>{act.status}</Badge></td>
                        <td className="font-mono text-xs text-gray-500 dark:text-gray-400">{formatDuration(act.duration)}</td>
                        <td className="text-xs text-gray-500 dark:text-gray-400">{act.attempt}</td>
                        <td className="text-xs text-red-500 max-w-[200px] truncate">{act.failure || '-'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Orphan activities (main workflow body) */}
          {wfInfo.orphan_activities?.length > 0 && (
            <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
              <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Orphan Activities ({wfInfo.orphan_activities.length})</h2>
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1 mb-2">Activities not associated with any update execution (main workflow body)</p>
              <div className="table-card">
                <table>
                  <thead>
                    <tr>
                      <th>Name</th><th>Status</th><th>Duration</th><th>Attempt</th><th>Failure</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
                    {wfInfo.orphan_activities.map((act: any, i: number) => (
                      <tr key={i}>
                        <td className="font-mono text-xs">{act.name}</td>
                        <td><Badge variant="status" status={act.status}>{act.status}</Badge></td>
                        <td className="font-mono text-xs text-gray-500 dark:text-gray-400">{formatDuration(act.duration)}</td>
                        <td className="text-xs text-gray-500 dark:text-gray-400">{act.attempt}</td>
                        <td className="text-xs text-red-500 max-w-[200px] truncate">{act.failure || '-'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Update executions */}
          {wfInfo.update_executions?.length > 0 && (
            <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
              <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Update Executions ({wfInfo.update_executions.length})</h2>
              <div className="mt-2 space-y-2">
                {wfInfo.update_executions.map((ue: any, i: number) => (
                  <UpdateExecutionCard key={i} ue={ue} />
                ))}
              </div>
            </div>
          )}

          {/* Child workflows */}
          {wfInfo.child_workflows?.length > 0 && (
            <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
              <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Child Workflows ({wfInfo.child_workflows.length})</h2>
              <div className="mt-2 table-card">
                <table>
                  <thead>
                    <tr>
                      <th>Type</th><th>Status</th><th>Namespace</th><th>Duration</th><th>Workflow ID</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
                    {wfInfo.child_workflows.map((cw: any, i: number) => (
                      <tr key={i}>
                        <td className="font-mono text-xs">{cw.workflow_type}</td>
                        <td><Badge variant="status" status={cw.status}>{cw.status}</Badge></td>
                        <td className="text-xs text-gray-500 dark:text-gray-400">{cw.namespace}</td>
                        <td className="font-mono text-xs text-gray-500 dark:text-gray-400">{formatDuration(cw.duration)}</td>
                        <td className="font-mono text-xs">
                          {temporalUIUrl ? (
                            <a
                              href={`${temporalUIUrl}/namespaces/${cw.namespace}/workflows/${cw.workflow_id}`}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300"
                            >
                              {truncateId(cw.workflow_id)}
                            </a>
                          ) : truncateId(cw.workflow_id)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Awaited signals */}
          {wfInfo.awaited_signals?.length > 0 && (
            <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
              <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Awaited Signals ({wfInfo.awaited_signals.length})</h2>
              <div className="mt-2 table-card">
                <table>
                  <thead>
                    <tr>
                      <th>Signal ID</th><th>Status</th><th>Duration</th><th>Failure</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
                    {wfInfo.awaited_signals.map((as: any, i: number) => (
                      <tr key={i}>
                        <td className="font-mono text-xs">
                          {as.queue_signal_id ? (
                            <Link
                              to={`/queue-signals?search=${as.queue_signal_id}`}
                              className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300"
                            >
                              {truncateId(as.queue_signal_id)}
                            </Link>
                          ) : '-'}
                        </td>
                        <td><Badge variant="status" status={as.status}>{as.status}</Badge></td>
                        <td className="font-mono text-xs text-gray-500 dark:text-gray-400">{formatDuration(as.duration)}</td>
                        <td className="text-xs text-red-500 max-w-[200px] truncate">{as.failure || '-'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Update handlers */}
          {wfInfo.update_handlers?.length > 0 && (
            <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
              <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Update Handlers</h2>
              <div className="mt-2 flex flex-wrap gap-2">
                {wfInfo.update_handlers.map((h: string) => <Badge key={h}>{h}</Badge>)}
              </div>
            </div>
          )}
        </>
      )}

      {submitted && !isLoading && !error && !wfInfo && (
        <p className="text-sm text-gray-500 dark:text-gray-400">No workflow found with the given parameters</p>
      )}
    </div>
  )
}

function UpdateExecutionCard({ ue }: { ue: any }) {
  const [expanded, setExpanded] = useState(false)
  return (
    <div className="border border-gray-200 dark:border-gray-800 rounded-md">
      <button onClick={() => setExpanded(!expanded)} className="w-full flex items-center justify-between px-3 py-2 text-left hover:bg-gray-50 dark:hover:bg-gray-800">
        <div className="flex items-center gap-2 text-xs">
          <Badge variant="status" status={ue.status}>{ue.status}</Badge>
          <span className="font-mono font-medium">{ue.name}</span>
          <span className="text-gray-400 dark:text-gray-500">{formatDuration(ue.duration)}</span>
        </div>
        <span className="text-gray-400 dark:text-gray-500 text-xs">{expanded ? '▾' : '▸'}</span>
      </button>
      {expanded && (
        <div className="border-t border-gray-200 dark:border-gray-800 px-3 py-2 text-xs space-y-2">
          <div><span className="text-gray-500 dark:text-gray-400">Update ID:</span> <span className="font-mono select-all">{ue.update_id}</span></div>
          <div><span className="text-gray-500 dark:text-gray-400">Duration:</span> <span className="font-mono">{formatDuration(ue.duration)}</span></div>
          {ue.input && (
            <div>
              <span className="text-gray-500 dark:text-gray-400">Input:</span>
              <pre className="mt-0.5 font-mono rounded p-2 overflow-x-auto max-h-32">{ue.input}</pre>
            </div>
          )}
          {ue.result && (
            <div>
              <span className="text-gray-500 dark:text-gray-400">Result:</span>
              <pre className="mt-0.5 font-mono rounded p-2 overflow-x-auto max-h-32">{ue.result}</pre>
            </div>
          )}
          {ue.failure && (
            <div>
              <span className="text-gray-500 dark:text-gray-400">Failure:</span>
              <pre className="mt-0.5 font-mono text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/30 rounded p-2 overflow-x-auto max-h-32">{ue.failure}</pre>
            </div>
          )}
          {ue.activities?.length > 0 && (
            <div>
              <p className="text-gray-500 dark:text-gray-400 mb-1">Activities ({ue.activities.length}):</p>
              <div className="table-card">
                <table>
                  <thead>
                    <tr>
                      <th>Name</th><th>Status</th><th>Duration</th><th>Attempt</th><th>Failure</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
                    {ue.activities.map((a: any, i: number) => (
                      <tr key={i}>
                        <td className="font-mono text-xs">{a.name}</td>
                        <td><Badge variant="status" status={a.status}>{a.status}</Badge></td>
                        <td className="font-mono text-xs text-gray-500 dark:text-gray-400">{formatDuration(a.duration)}</td>
                        <td className="text-xs text-gray-500 dark:text-gray-400">{a.attempt}</td>
                        <td className="text-xs text-red-500 max-w-[200px] truncate">{a.failure || '-'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
