import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router'
import { getAppBranchDetail, getAppBranchRuns, getAppBranchWorkflows } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, formatRelativeDate, truncateId } from '@/utils/format'

function workflowStatus(status: any): string {
  if (!status) return ''
  if (typeof status === 'string') return status
  return String(status.status || '')
}

export const AppBranchDetail = () => {
  const { id } = useParams<{ id: string }>()
  const [runsPage, setRunsPage] = useState(1)
  const [workflowsPage, setWorkflowsPage] = useState(1)

  const { data, isLoading, error } = useQuery({
    queryKey: ['app-branch', id],
    queryFn: () => getAppBranchDetail(id!),
    enabled: !!id,
    refetchInterval: 20000,
  })

  const { data: runsData } = useQuery({
    queryKey: ['app-branch-runs', id, runsPage],
    queryFn: () => getAppBranchRuns(id!, { page: runsPage }),
    enabled: !!id,
    refetchInterval: 20000,
  })

  const { data: workflowsData } = useQuery({
    queryKey: ['app-branch-workflows', id, workflowsPage],
    queryFn: () => getAppBranchWorkflows(id!, { page: workflowsPage }),
    enabled: !!id,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load app branch'} />
  if (!data) return null

  const { app_branch: branch, org_name: orgName, app_name: appName, created_by: createdBy, queues = [] } = data
  const runs = runsData?.runs ?? data.runs ?? []
  const runsTotalPages = runsData?.total_pages ?? data.runs_total_pages ?? 1
  const workflows = workflowsData?.workflows ?? []

  return (
    <div className="space-y-6">
      <nav className="text-sm text-gray-500 dark:text-gray-400">
        <Link to="/app-branches" className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200">App branches</Link>
        <span className="mx-1">/</span>
        <span className="font-mono">{truncateId(branch.id)}</span>
      </nav>

      <div>
        <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">{branch.name || truncateId(branch.id)}</h1>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400 font-mono">{branch.id}</p>
        <div className="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-gray-500 dark:text-gray-400">
          <span>
            Org: <Link to={`/orgs/${branch.org_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200">{orgName || truncateId(branch.org_id)}</Link>
          </span>
          <span>App: {appName || truncateId(branch.app_id)} <span className="font-mono text-xs">({truncateId(branch.app_id)})</span></span>
          {branch.managed_by && <span>Managed by: <Badge>{branch.managed_by}</Badge></span>}
          <span>Workflows: {branch.workflow_count ?? 0}</span>
          <span>Created: {formatDate(branch.created_at)}</span>
          {createdBy?.email && (
            <span>
              by <Link to={`/accounts/${branch.created_by_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200">{createdBy.email}</Link>
            </span>
          )}
        </div>
      </div>

      {/* Queues */}
      <div className="table-card rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Queues</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
            <thead>
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Emitters</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {queues.map((queue) => (
                <tr key={queue.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Link to={`/queues/${queue.id}`} className="font-mono text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
                      {truncateId(queue.id)}
                    </Link>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{queue.name}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{queue.emitters?.length ?? 0}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDate(queue.created_at)}</td>
                </tr>
              ))}
              {queues.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No queues</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Runs */}
      <div className="table-card rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Runs</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
            <thead>
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Workflow</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Commit / PR</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {runs.map((run) => (
                <tr key={run.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="whitespace-nowrap px-4 py-3 text-sm font-mono text-xs text-gray-500 dark:text-gray-400">{truncateId(run.id)}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Badge variant="status" status={run.status}>{run.status || '-'}</Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{run.run_type}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    {run.workflow_id ? (
                      <Link to={`/workflows/${run.workflow_id}`} className="font-mono text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
                        {truncateId(run.workflow_id)}
                      </Link>
                    ) : (
                      <span className="text-xs text-gray-400 dark:text-gray-500">-</span>
                    )}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-xs text-gray-500 dark:text-gray-400">
                    {run.head_sha ? <span className="font-mono">{run.head_sha.slice(0, 8)}</span> : '-'}
                    {run.pr_number ? <span className="ml-1">(#{run.pr_number})</span> : null}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatRelativeDate(run.created_at)}</td>
                </tr>
              ))}
              {runs.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No runs</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        <Pagination page={runsData?.page ?? runsPage} totalPages={runsTotalPages} onPageChange={setRunsPage} />
      </div>

      {/* Workflows */}
      <div className="table-card rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Workflows</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
            <thead>
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {workflows.map((wf) => {
                const status = workflowStatus(wf.status)
                return (
                  <tr key={wf.id} className="hover:bg-gray-50 dark:hover:bg-gray-800">
                    <td className="whitespace-nowrap px-4 py-3 text-sm">
                      <Link to={`/workflows/${wf.id}`} className="font-mono text-xs text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
                        {truncateId(wf.id)}
                      </Link>
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{wf.type}</td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm">
                      <Badge variant="status" status={status}>{status || '-'}</Badge>
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatRelativeDate(wf.created_at)}</td>
                  </tr>
                )
              })}
              {workflows.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No workflows</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        <Pagination
          page={workflowsData?.page ?? workflowsPage}
          totalPages={workflowsData?.total_pages ?? 1}
          onPageChange={setWorkflowsPage}
        />
      </div>
    </div>
  )
}
