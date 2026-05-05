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
import { formatDate, formatDuration, truncateId } from '@/utils/format'

function getStatus(s: any): string {
  if (!s) return ''
  if (typeof s === 'string') return s
  if (typeof s === 'object' && s.status) return String(s.status)
  return String(s)
}

function getStatusDescription(s: any): string {
  if (!s) return ''
  if (typeof s === 'object' && s.status_human_description) return String(s.status_human_description)
  return ''
}

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

  const { install, app_url: appUrl } = data
  const dashboardUrl = appUrl ? `${appUrl}/${install.org_id}/installs/${install.id}` : null

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <nav className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
        <Link to="/installs" className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200">Installs</Link>
        <span>/</span>
        <span className="text-gray-900 dark:text-gray-100">{install.name || truncateId(install.id)}</span>
      </nav>

      {/* Page Heading */}
      <div className="page-heading">
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">{install.name || truncateId(install.id)}</h1>
          <Badge variant="status" status={install.status}>{install.status}</Badge>
          {install.status_description && (
            <span className="text-xs text-gray-500 dark:text-gray-400" title={install.status_description}>{install.status_description}</span>
          )}
        </div>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400 font-mono">{install.id}</p>
        <div className="mt-2 flex flex-wrap items-center gap-3 text-sm">
          <span className="text-gray-500 dark:text-gray-400">
            Org: <Link to={`/orgs/${install.org_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200">{install.org?.name || truncateId(install.org_id)}</Link>
          </span>
          <span className="text-gray-500 dark:text-gray-400">
            App: {install.app?.name || truncateId(install.app_id)}
          </span>
          {install.cloud_platform && (
            <span className="text-gray-500 dark:text-gray-400">
              Cloud: <Badge>{install.cloud_platform}</Badge>
            </span>
          )}
          {install.runner_type && (
            <span className="text-gray-500 dark:text-gray-400">
              Runner Type: <Badge>{install.runner_type}</Badge>
            </span>
          )}
          {install.app_config?.version ? (
            <span className="text-gray-500 dark:text-gray-400">Config v{install.app_config.version}</span>
          ) : null}
        </div>
        <div className="mt-1 flex items-center gap-4 text-xs text-gray-400 dark:text-gray-500">
          <span>Created {formatDate(install.created_at)}</span>
          <span>Updated {formatDate(install.updated_at)}</span>
        </div>
        <div className="mt-2 flex items-center gap-3">
          {dashboardUrl && (
            <a href={dashboardUrl} target="_blank" rel="noopener noreferrer" className="text-sm text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200 underline">
              View Dashboard
            </a>
          )}
          <Link to={`/queues?search=${install.id}`} className="text-sm text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200 underline">
            View Queues
          </Link>
        </div>
      </div>

      {/* Labels */}
      <div className="table-card p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Labels</h2>
        <div className="mt-2 flex flex-wrap gap-2">
          {Object.entries(install.labels || {}).map(([key, value]) => (
            <span key={key} className="inline-flex items-center gap-1 rounded-full bg-gray-100 dark:bg-gray-800 px-2.5 py-0.5 text-xs font-medium text-gray-700 dark:text-gray-300">
              {key}={value}
              <button
                onClick={() => removeLabelMutation.mutate(key)}
                className="ml-1 text-gray-400 dark:text-gray-500 hover:text-red-500"
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
            className="block w-32 rounded-md border-0 py-1 px-2 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 focus:ring-2 focus:ring-primary-600 dark:focus:ring-primary-500"
          />
          <input
            type="text"
            value={newLabelValue}
            onChange={(e) => setNewLabelValue(e.target.value)}
            placeholder="Value"
            className="block w-32 rounded-md border-0 py-1 px-2 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 focus:ring-2 focus:ring-primary-600 dark:focus:ring-primary-500"
          />
          <button
            onClick={() => newLabelKey.trim() && addLabelMutation.mutate()}
            disabled={addLabelMutation.isPending || !newLabelKey.trim()}
            className="rounded-md bg-primary-600 dark:bg-primary-500 px-3 py-1 text-sm font-medium text-white hover:bg-primary-700 dark:hover:bg-primary-600 disabled:opacity-50"
          >
            Add
          </button>
        </div>
      </div>

      {/* Status Badges */}
      <div className="table-card p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Status</h2>
        <div className="mt-2 grid grid-cols-2 gap-4 sm:grid-cols-4">
          <div>
            <p className="text-xs text-gray-500 dark:text-gray-400">Runner</p>
            {runnerStatus ? (
              <div className="flex items-center gap-1">
                <Badge variant="status" status={runnerStatus.status}>{runnerStatus.status}</Badge>
                {runnerStatus.description && (
                  <span className="text-xs text-gray-400 dark:text-gray-500" title={runnerStatus.description}>{runnerStatus.description}</span>
                )}
              </div>
            ) : (
              <span className="text-xs text-gray-400 dark:text-gray-500">Loading...</span>
            )}
          </div>
          <div>
            <p className="text-xs text-gray-500 dark:text-gray-400">Sandbox</p>
            {sandboxStatus ? (
              <div className="flex items-center gap-1">
                <Badge variant="status" status={sandboxStatus.status}>{sandboxStatus.status}</Badge>
                {sandboxStatus.description && (
                  <span className="text-xs text-gray-400 dark:text-gray-500" title={sandboxStatus.description}>{sandboxStatus.description}</span>
                )}
              </div>
            ) : (
              <span className="text-xs text-gray-400 dark:text-gray-500">Loading...</span>
            )}
          </div>
          <div>
            <p className="text-xs text-gray-500 dark:text-gray-400">Component</p>
            {componentStatus ? (
              <div className="flex items-center gap-1">
                <Badge variant="status" status={componentStatus.status}>{componentStatus.status}</Badge>
                {componentStatus.description && (
                  <span className="text-xs text-gray-400 dark:text-gray-500" title={componentStatus.description}>{componentStatus.description}</span>
                )}
              </div>
            ) : (
              <span className="text-xs text-gray-400 dark:text-gray-500">Loading...</span>
            )}
          </div>
          <div>
            <p className="text-xs text-gray-500 dark:text-gray-400">Drift</p>
            {driftStatus ? (
              <div className="flex items-center gap-1">
                <Badge variant="status" status={driftStatus.status}>{driftStatus.status}</Badge>
                {driftStatus.description && (
                  <span className="text-xs text-gray-400 dark:text-gray-500" title={driftStatus.description}>{driftStatus.description}</span>
                )}
              </div>
            ) : (
              <span className="text-xs text-gray-400 dark:text-gray-500">Loading...</span>
            )}
          </div>
        </div>
      </div>

      {/* Active Deployments */}
      <div className="table-card p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Active Deployments</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
            <thead className="bg-gray-50 dark:bg-gray-900">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Component</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Build ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800 bg-white dark:bg-gray-900">
              {(deploymentsData?.deployments || []).map((dep: any) => (
                <tr key={dep.id}>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400 font-mono">{truncateId(dep.id)}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{dep.component_name || '-'}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{dep.install_deploy_type || '-'}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400 font-mono">{dep.build_id ? truncateId(dep.build_id) : '-'}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <div className="flex items-center gap-1">
                      <Badge variant="status" status={dep.status}>{dep.status}</Badge>
                      {dep.status_description && (
                        <span className="text-xs text-gray-400 dark:text-gray-500" title={dep.status_description}>{dep.status_description}</span>
                      )}
                    </div>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDate(dep.created_at)}</td>
                </tr>
              ))}
              {(!deploymentsData?.deployments || deploymentsData.deployments.length === 0) && (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No active deployments</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Workflows */}
      <div className="table-card p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Workflows</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
            <thead className="bg-gray-50 dark:bg-gray-900">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Duration</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800 bg-white dark:bg-gray-900">
              {(workflowsData?.workflows || []).map((wf: any) => {
                const wfStatus = getStatus(wf.status)
                const wfStatusDesc = getStatusDescription(wf.status)
                return (
                  <tr key={wf.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                    <td className="whitespace-nowrap px-4 py-3 text-sm">
                      <Link to={`/workflows/${wf.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200 font-mono">
                        {truncateId(wf.id)}
                      </Link>
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{wf.type}</td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm">
                      <div className="flex items-center gap-1">
                        <Badge variant="status" status={wfStatus}>{wfStatus}</Badge>
                        {wfStatusDesc && (
                          <span className="text-xs text-gray-400 dark:text-gray-500" title={wfStatusDesc}>{wfStatusDesc}</span>
                        )}
                      </div>
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDuration(wf.execution_time)}</td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDate(wf.created_at)}</td>
                  </tr>
                )
              })}
              {(!workflowsData?.workflows || workflowsData.workflows.length === 0) && (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No workflows</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        {workflowsData && (
          <Pagination page={workflowsPage} totalPages={workflowsData.total_pages} onPageChange={setWorkflowsPage} />
        )}
      </div>

      {/* Activity */}
      <div className="table-card p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Activity</h2>
        <div className="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center">
          <select
            value={activityEntityType}
            onChange={(e) => { setActivityEntityType(e.target.value); setActivityPage(1) }}
            className="block w-48 rounded-md border-0 py-1 px-2 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 focus:ring-2 focus:ring-primary-600 dark:focus:ring-primary-500"
          >
            <option value="">All types</option>
            <option value="runner_job">Runner Job</option>
            <option value="workflow">Workflow</option>
          </select>
          <input
            type="date"
            value={activityStartDate}
            onChange={(e) => { setActivityStartDate(e.target.value); setActivityPage(1) }}
            className="block rounded-md border-0 py-1 px-2 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 focus:ring-2 focus:ring-primary-600 dark:focus:ring-primary-500"
          />
          <input
            type="date"
            value={activityEndDate}
            onChange={(e) => { setActivityEndDate(e.target.value); setActivityPage(1) }}
            className="block rounded-md border-0 py-1 px-2 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 focus:ring-2 focus:ring-primary-600 dark:focus:ring-primary-500"
          />
        </div>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
            <thead className="bg-gray-50 dark:bg-gray-900">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Entity Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Entity ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Description</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800 bg-white dark:bg-gray-900">
              {(activityData?.activity_logs || []).map((entry: any, idx: number) => (
                <tr key={entry.entity_id + '-' + idx}>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Badge>{entry.entity_type}</Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{entry.entity_name || '-'}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm font-mono">
                    {entry.entity_type === 'workflow' ? (
                      <Link to={`/workflows/${entry.entity_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200">
                        {truncateId(entry.entity_id)}
                      </Link>
                    ) : (
                      <span className="text-gray-500 dark:text-gray-400">{truncateId(entry.entity_id)}</span>
                    )}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{entry.description || '-'}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDate(entry.created_at)}</td>
                </tr>
              ))}
              {(!activityData?.activity_logs || activityData.activity_logs.length === 0) && (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No activity</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        {activityData && (
          <Pagination page={activityPage} totalPages={activityData.total_pages} onPageChange={setActivityPage} />
        )}
      </div>
    </div>
  )
}
