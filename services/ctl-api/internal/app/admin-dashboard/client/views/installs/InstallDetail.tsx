import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router'
import {
  getInstallDetail,
  getInstallRunnerStatus,
  getInstallSandboxStatus,
  getInstallComponentStatus,
  getInstallDriftStatus,
  getInstallActiveDeployments,
  getInstallActivity,
  getInstallWorkflows,
  addInstallLabel,
  removeInstallLabel,
} from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

export const InstallDetail = () => {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const [activityPage, setActivityPage] = useState(1)
  const [activityEntityType, setActivityEntityType] = useState('')
  const [activityStartDate, setActivityStartDate] = useState('')
  const [activityEndDate, setActivityEndDate] = useState('')
  const [workflowsPage, setWorkflowsPage] = useState(1)
  const [newLabelKey, setNewLabelKey] = useState('')
  const [newLabelValue, setNewLabelValue] = useState('')

  const { data, isLoading, error } = useQuery({
    queryKey: ['install', id],
    queryFn: () => getInstallDetail(id!),
    enabled: !!id,
  })

  const { data: runnerStatus } = useQuery({
    queryKey: ['install-runner-status', id],
    queryFn: () => getInstallRunnerStatus(id!),
    enabled: !!id,
    refetchInterval: 5000,
  })

  const { data: sandboxStatus } = useQuery({
    queryKey: ['install-sandbox-status', id],
    queryFn: () => getInstallSandboxStatus(id!),
    enabled: !!id,
    refetchInterval: 5000,
  })

  const { data: componentStatus } = useQuery({
    queryKey: ['install-component-status', id],
    queryFn: () => getInstallComponentStatus(id!),
    enabled: !!id,
    refetchInterval: 5000,
  })

  const { data: driftStatus } = useQuery({
    queryKey: ['install-drift-status', id],
    queryFn: () => getInstallDriftStatus(id!),
    enabled: !!id,
    refetchInterval: 5000,
  })

  const { data: deploymentsData } = useQuery({
    queryKey: ['install-deployments', id],
    queryFn: () => getInstallActiveDeployments(id!),
    enabled: !!id,
    refetchInterval: 5000,
  })

  const { data: activityData } = useQuery({
    queryKey: ['install-activity', id, activityPage, activityEntityType, activityStartDate, activityEndDate],
    queryFn: () =>
      getInstallActivity(id!, {
        page: activityPage,
        entity_type: activityEntityType || undefined,
        start_date: activityStartDate || undefined,
        end_date: activityEndDate || undefined,
      }),
    enabled: !!id,
  })

  const { data: workflowsData } = useQuery({
    queryKey: ['install-workflows', id, workflowsPage],
    queryFn: () => getInstallWorkflows(id!, { page: workflowsPage }),
    enabled: !!id,
  })

  const addLabelMutation = useMutation({
    mutationFn: () => addInstallLabel(id!, newLabelKey, newLabelValue),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['install', id] })
      setNewLabelKey('')
      setNewLabelValue('')
    },
  })

  const removeLabelMutation = useMutation({
    mutationFn: (key: string) => removeInstallLabel(id!, key),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['install', id] }),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load install'} />
  if (!data) return null

  const { install } = data

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-gray-900">{install.name || truncateId(install.id)}</h1>
        <p className="mt-1 text-sm text-gray-500 font-mono">{install.id}</p>
        <div className="mt-2 flex items-center gap-3 text-sm">
          <Badge variant="status" status={install.status}>{install.status}</Badge>
          <span className="text-gray-500">
            Org: <Link to={`/orgs/${install.org_id}`} className="text-primary-600 hover:text-primary-800">{install.org?.name || truncateId(install.org_id)}</Link>
          </span>
          <span className="text-gray-500">
            App: {install.app?.name || truncateId(install.app_id)}
          </span>
        </div>
      </div>

      {/* Labels */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Labels</h2>
        <div className="mt-2 flex flex-wrap gap-2">
          {Object.entries(install.labels || {}).map(([key, value]) => (
            <span key={key} className="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-700">
              {key}={value}
              <button
                onClick={() => removeLabelMutation.mutate(key)}
                className="ml-1 text-gray-400 hover:text-red-500"
                disabled={removeLabelMutation.isPending}
              >
                x
              </button>
            </span>
          ))}
        </div>
        <div className="mt-2 flex gap-2">
          <input
            type="text"
            value={newLabelKey}
            onChange={(e) => setNewLabelKey(e.target.value)}
            placeholder="Key"
            className="block w-32 rounded-md border-0 py-1 px-2 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600"
          />
          <input
            type="text"
            value={newLabelValue}
            onChange={(e) => setNewLabelValue(e.target.value)}
            placeholder="Value"
            className="block w-32 rounded-md border-0 py-1 px-2 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600"
          />
          <button
            onClick={() => newLabelKey.trim() && addLabelMutation.mutate()}
            disabled={addLabelMutation.isPending || !newLabelKey.trim()}
            className="rounded-md bg-primary-600 px-3 py-1 text-sm font-medium text-white hover:bg-primary-700 disabled:opacity-50"
          >
            Add
          </button>
        </div>
      </div>

      {/* Status Badges */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Status</h2>
        <div className="mt-2 grid grid-cols-2 gap-4 sm:grid-cols-4">
          <div>
            <p className="text-xs text-gray-500">Runner</p>
            {runnerStatus ? (
              <Badge variant="status" status={runnerStatus.status}>{runnerStatus.status}</Badge>
            ) : (
              <span className="text-xs text-gray-400">Loading...</span>
            )}
          </div>
          <div>
            <p className="text-xs text-gray-500">Sandbox</p>
            {sandboxStatus ? (
              <Badge variant="status" status={sandboxStatus.status}>{sandboxStatus.status}</Badge>
            ) : (
              <span className="text-xs text-gray-400">Loading...</span>
            )}
          </div>
          <div>
            <p className="text-xs text-gray-500">Component</p>
            {componentStatus ? (
              <Badge variant="status" status={componentStatus.status}>{componentStatus.status}</Badge>
            ) : (
              <span className="text-xs text-gray-400">Loading...</span>
            )}
          </div>
          <div>
            <p className="text-xs text-gray-500">Drift</p>
            {driftStatus ? (
              <Badge variant="status" status={driftStatus.status}>{driftStatus.status}</Badge>
            ) : (
              <span className="text-xs text-gray-400">Loading...</span>
            )}
          </div>
        </div>
      </div>

      {/* Active Deployments */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Active Deployments</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {(deploymentsData?.deployments || []).map((dep: any) => (
                <tr key={dep.id}>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 font-mono">{truncateId(dep.id)}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Badge variant="status" status={dep.status}>{dep.status}</Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(dep.created_at)}</td>
                </tr>
              ))}
              {(!deploymentsData?.deployments || deploymentsData.deployments.length === 0) && (
                <tr>
                  <td colSpan={3} className="px-4 py-8 text-center text-sm text-gray-500">No active deployments</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Activity */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Activity</h2>
        <div className="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center">
          <input
            type="text"
            value={activityEntityType}
            onChange={(e) => { setActivityEntityType(e.target.value); setActivityPage(1) }}
            placeholder="Entity type filter..."
            className="block w-48 rounded-md border-0 py-1 px-2 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600"
          />
          <input
            type="date"
            value={activityStartDate}
            onChange={(e) => { setActivityStartDate(e.target.value); setActivityPage(1) }}
            className="block rounded-md border-0 py-1 px-2 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600"
          />
          <input
            type="date"
            value={activityEndDate}
            onChange={(e) => { setActivityEndDate(e.target.value); setActivityPage(1) }}
            className="block rounded-md border-0 py-1 px-2 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600"
          />
        </div>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Action</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Entity Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Entity ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {(activityData?.entries || []).map((entry) => (
                <tr key={entry.id}>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{entry.action}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{entry.entity_type}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 font-mono">{truncateId(entry.entity_id)}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(entry.created_at)}</td>
                </tr>
              ))}
              {(!activityData?.entries || activityData.entries.length === 0) && (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-sm text-gray-500">No activity</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        {activityData && (
          <Pagination page={activityPage} totalPages={activityData.total_pages} onPageChange={setActivityPage} />
        )}
      </div>

      {/* Workflows */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Workflows</h2>
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
              {(workflowsData?.workflows || []).map((wf) => (
                <tr key={wf.id} className="hover:bg-gray-50">
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Link to={`/workflows/${wf.id}`} className="text-primary-600 hover:text-primary-800 font-mono">
                      {truncateId(wf.id)}
                    </Link>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{wf.type}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Badge variant="status" status={wf.status}>{wf.status}</Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(wf.created_at)}</td>
                </tr>
              ))}
              {(!workflowsData?.workflows || workflowsData.workflows.length === 0) && (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-sm text-gray-500">No workflows</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        {workflowsData && (
          <Pagination page={workflowsPage} totalPages={workflowsData.total_pages} onPageChange={setWorkflowsPage} />
        )}
      </div>
    </div>
  )
}
