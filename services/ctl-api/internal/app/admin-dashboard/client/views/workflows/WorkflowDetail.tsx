import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useParams } from 'react-router'
import { getWorkflowDetail } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { JsonViewer } from '@/components/common/JsonViewer'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, formatDuration, truncateId } from '@/utils/format'
import type { TGroupDetailData, TStepDetailData } from '@/types/admin.types'

const StepRow = ({ stepData }: { stepData: TStepDetailData }) => {
  const [expanded, setExpanded] = useState(false)
  const { step, queue_signal_json, step_target } = stepData

  return (
    <>
      <tr className="hover:bg-gray-50">
        <td className="whitespace-nowrap px-4 py-2 text-sm text-gray-900">{step.idx}</td>
        <td className="whitespace-nowrap px-4 py-2 text-sm text-gray-900">{step.step_target_type}</td>
        <td className="whitespace-nowrap px-4 py-2 text-sm text-gray-500 font-mono">{truncateId(step.step_target_id)}</td>
        <td className="whitespace-nowrap px-4 py-2 text-sm">
          <Badge variant="status" status={getStatus(step.status)}>{getStatus(step.status) || '-'}</Badge>
        </td>
        <td className="whitespace-nowrap px-4 py-2 text-sm text-gray-500">
          {step.approval ? JSON.stringify(step.approval) : '-'}
        </td>
        <td className="whitespace-nowrap px-4 py-2 text-sm">
          {queue_signal_json && (
            <button
              onClick={() => setExpanded(!expanded)}
              className="text-xs text-primary-600 hover:text-primary-800"
            >
              {expanded ? 'Hide' : 'Show'} JSON
            </button>
          )}
        </td>
      </tr>
      {expanded && queue_signal_json && (
        <tr>
          <td colSpan={6} className="px-4 py-2">
            <JsonViewer data={queue_signal_json} />
          </td>
        </tr>
      )}
    </>
  )
}

const StepGroupSection = ({ group }: { group: TGroupDetailData }) => (
  <div className="rounded-md border border-gray-200 p-3">
    <div className="flex items-center gap-2 mb-2">
      <span className="text-xs font-semibold text-gray-700">
        Group {group.group.group_idx} (retry {group.group.group_retry_idx})
      </span>
      <Badge variant="status" status={getStatus(group.group.status)}>{getStatus(group.group.status) || '-'}</Badge>
    </div>
    <table className="min-w-full divide-y divide-gray-200">
      <thead className="bg-gray-50">
        <tr>
          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Idx</th>
          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Target</th>
          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Approval</th>
          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Signal</th>
        </tr>
      </thead>
      <tbody className="divide-y divide-gray-200 bg-white">
        {group.steps.map((sd) => (
          <StepRow key={sd.step.id} stepData={sd} />
        ))}
      </tbody>
    </table>
  </div>
)

function getStatus(status: any): string {
  if (!status) return ''
  if (typeof status === 'string') return status
  if (typeof status === 'object' && status.status) return String(status.status)
  return String(status)
}

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

  const workflow = data.workflow
  const groups: TGroupDetailData[] = data.group_details || []
  const workflow_info = data.workflow_info || null
  const wfStatus = getStatus(workflow?.status)

  return (
    <div className="space-y-6">
      <div>
        <h1 className="page-heading">Workflow detail</h1>
        <p className="mt-1 text-sm text-gray-500 font-mono">{workflow?.id}</p>
        <div className="mt-2 flex items-center gap-3 text-sm">
          <Badge variant="status" status={wfStatus}>{wfStatus || '-'}</Badge>
          <span className="text-gray-500">Type: <span className="font-mono">{workflow?.type}</span></span>
          <span className="text-gray-500">Created: {formatDate(workflow?.created_at)}</span>
        </div>
        <div className="mt-1 text-sm text-gray-500">
          Owner: <span className="font-mono text-xs">{truncateId(workflow?.owner_id)}</span> ({workflow?.owner_type})
          {workflow?.created_by?.email && (
            <span className="ml-2">by {workflow.created_by.email}</span>
          )}
        </div>
      </div>

      {/* Step Groups */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Step Groups</h2>
        <div className="mt-2 space-y-3">
          {groups.map((group, i) => (
            <StepGroupSection key={i} group={group} />
          ))}
          {groups.length === 0 && (
            <p className="text-sm text-gray-500">No step groups</p>
          )}
        </div>
      </div>

      {/* Workflow Info */}
      {workflow_info && (
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h2 className="text-sm font-semibold text-gray-900">Workflow Info</h2>

          {/* Activities */}
          {workflow_info.activities && workflow_info.activities.length > 0 && (
            <div className="mt-3">
              <h3 className="text-xs font-semibold text-gray-700 mb-1">Activities ({workflow_info.activities.length})</h3>
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Duration</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Attempt</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 bg-white">
                    {workflow_info.activities.map((act, i) => (
                      <tr key={i}>
                        <td className="whitespace-nowrap px-3 py-2 text-xs text-gray-900">{act.name}</td>
                        <td className="whitespace-nowrap px-3 py-2 text-xs">
                          <Badge variant="status" status={act.status}>{act.status}</Badge>
                        </td>
                        <td className="whitespace-nowrap px-3 py-2 text-xs text-gray-500">{formatDuration(act.duration)}</td>
                        <td className="whitespace-nowrap px-3 py-2 text-xs text-gray-500">{act.attempt}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Child Workflows */}
          {workflow_info.child_workflows && workflow_info.child_workflows.length > 0 && (
            <div className="mt-3">
              <h3 className="text-xs font-semibold text-gray-700 mb-1">Child Workflows ({workflow_info.child_workflows.length})</h3>
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Namespace</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Duration</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 bg-white">
                    {workflow_info.child_workflows.map((cw, i) => (
                      <tr key={i}>
                        <td className="whitespace-nowrap px-3 py-2 text-xs text-gray-900">{cw.workflow_type}</td>
                        <td className="whitespace-nowrap px-3 py-2 text-xs">
                          <Badge variant="status" status={cw.status}>{cw.status}</Badge>
                        </td>
                        <td className="whitespace-nowrap px-3 py-2 text-xs text-gray-500">{cw.namespace}</td>
                        <td className="whitespace-nowrap px-3 py-2 text-xs text-gray-500">{formatDuration(cw.duration)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Update Executions */}
          {workflow_info.update_executions && workflow_info.update_executions.length > 0 && (
            <div className="mt-3">
              <h3 className="text-xs font-semibold text-gray-700 mb-1">Update Executions ({workflow_info.update_executions.length})</h3>
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Update ID</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Duration</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 bg-white">
                    {workflow_info.update_executions.map((ue, i) => (
                      <tr key={i}>
                        <td className="whitespace-nowrap px-3 py-2 text-xs text-gray-900">{ue.name}</td>
                        <td className="whitespace-nowrap px-3 py-2 text-xs">
                          <Badge variant="status" status={ue.status}>{ue.status}</Badge>
                        </td>
                        <td className="whitespace-nowrap px-3 py-2 text-xs text-gray-500 font-mono">{truncateId(ue.update_id)}</td>
                        <td className="whitespace-nowrap px-3 py-2 text-xs text-gray-500">{formatDuration(ue.duration)}</td>
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
