import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router'
import {
  getOrgDetail, addOrgLabels, removeOrgLabel, addSupportUsers, migrateOrgQueues,
  clearOrgQueues, forceRestartOrgQueues, removeOldRunnerProcesses,
  shutdownOrgRunnerProcesses, shutdownHintOrgRunnerProcesses,
  deprovisionOrg, forgetOrg, forgetOrgInstalls, deprovisionOrgApps,
  forgetInstall, deprovisionInstall,
  getOrgWorkflows, terminateOrgWorkflows,
  getOrgQueueSignals, getOrgQueueSignalStats, deleteOrgQueueSignals,
} from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { DotGraph } from '@/components/common/DotGraph'
import { formatDate, truncateId } from '@/utils/format'

function getStatus(s: any): string {
  if (!s) return ''
  if (typeof s === 'string') return s
  if (typeof s === 'object' && s.status) return String(s.status)
  return String(s)
}

type Tab = 'overview' | 'queues' | 'cleanup'

export const OrgDetail = () => {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const [searchParams, setSearchParams] = useSearchParams()
  const activeTab = (searchParams.get('tab') as Tab) || 'overview'
  const [installsPage, setInstallsPage] = useState(1)
  const [labelKey, setLabelKey] = useState('')
  const [labelValue, setLabelValue] = useState('')

  const setTab = (tab: Tab) => {
    setSearchParams(tab === 'overview' ? {} : { tab })
  }

  const { data, isLoading, error } = useQuery({
    queryKey: ['org', id, installsPage],
    queryFn: () => getOrgDetail(id!, { page: installsPage }),
    enabled: !!id,
    refetchInterval: 20000,
  })

  // --- Overview mutations ---
  const addLabelMutation = useMutation({
    mutationFn: (labels: Record<string, string>) => addOrgLabels(id!, labels),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org', id] })
      setLabelKey('')
      setLabelValue('')
    },
  })

  const removeLabelMutation = useMutation({
    mutationFn: (key: string) => removeOrgLabel(id!, key),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', id] }),
  })

  const supportMutation = useMutation({
    mutationFn: () => addSupportUsers(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', id] }),
  })

  const removeOldProcessesMutation = useMutation({
    mutationFn: () => removeOldRunnerProcesses(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', id] }),
  })

  const shutdownProcessesMutation = useMutation({
    mutationFn: () => shutdownOrgRunnerProcesses(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', id] }),
  })

  const shutdownHintProcessesMutation = useMutation({
    mutationFn: () => shutdownHintOrgRunnerProcesses(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', id] }),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load organization'} />
  if (!data) return null

  const { org, installs = [], recent_app, graph_dot, app_url, installs_total_pages = 1 } = data
  const orgLabels = org.labels || {}

  const handleAddLabel = () => {
    if (!labelKey.trim()) return
    addLabelMutation.mutate({ [labelKey.trim()]: labelValue.trim() })
  }

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <nav className="flex items-center gap-1 text-sm text-gray-500 dark:text-gray-400">
        <Link to="/orgs" className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">Orgs</Link>
        <span>/</span>
        <span className="text-gray-900 dark:text-gray-100 font-medium">{org.name}</span>
      </nav>

      {/* Header */}
      <div>
        <div className="flex items-center gap-3">
          <h1 className="page-heading">{org.name}</h1>
          {app_url && (
            <a
              href={`${app_url}/${org.id}`}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 rounded-md bg-primary-600 dark:bg-primary-500 px-3 py-1.5 text-sm font-medium text-white hover:bg-primary-700 dark:hover:bg-primary-600"
            >
              View Dashboard
            </a>
          )}
        </div>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400 font-mono">{org.id}</p>
        <div className="mt-2 flex items-center gap-2">
          {org.status && (
            <Badge variant="status" status={getStatus(org.status)}>{getStatus(org.status)}</Badge>
          )}
          {org.status_description && (
            <span className="text-xs text-gray-500 dark:text-gray-400">{org.status_description}</span>
          )}
        </div>
        <div className="mt-2 text-xs text-gray-500 dark:text-gray-400 flex gap-4">
          <span>Created: {formatDate(org.created_at)}</span>
          <span>Updated: {formatDate(org.updated_at)}</span>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200 dark:border-gray-800">
        <nav className="flex gap-4">
          {(['overview', 'queues', 'cleanup'] as Tab[]).map((tab) => (
            <button
              key={tab}
              onClick={() => setTab(tab)}
              className={`pb-2 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab
                  ? 'border-primary-600 text-primary-600 dark:border-primary-400 dark:text-primary-400'
                  : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
              }`}
            >
              {tab.charAt(0).toUpperCase() + tab.slice(1)}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      {activeTab === 'overview' && (
        <OverviewTab
          org={org}
          orgLabels={orgLabels}
          installs={installs}
          installsPage={installsPage}
          setInstallsPage={setInstallsPage}
          installs_total_pages={installs_total_pages}
          recent_app={recent_app}
          graph_dot={graph_dot}
          labelKey={labelKey}
          setLabelKey={setLabelKey}
          labelValue={labelValue}
          setLabelValue={setLabelValue}
          handleAddLabel={handleAddLabel}
          addLabelMutation={addLabelMutation}
          removeLabelMutation={removeLabelMutation}
          supportMutation={supportMutation}
          removeOldProcessesMutation={removeOldProcessesMutation}
          shutdownProcessesMutation={shutdownProcessesMutation}
          shutdownHintProcessesMutation={shutdownHintProcessesMutation}
        />
      )}
      {activeTab === 'queues' && <QueuesTab orgId={id!} />}
      {activeTab === 'cleanup' && <CleanupTab orgId={id!} installs={installs} />}
    </div>
  )
}

// ---------- Overview Tab ----------
function OverviewTab({
  org, orgLabels, installs, installsPage, setInstallsPage, installs_total_pages,
  recent_app, graph_dot,
  labelKey, setLabelKey, labelValue, setLabelValue, handleAddLabel,
  addLabelMutation, removeLabelMutation, supportMutation,
  removeOldProcessesMutation, shutdownProcessesMutation, shutdownHintProcessesMutation,
}: any) {
  return (
    <>
      {/* Labels */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Labels</h2>
        <div className="mt-2 flex flex-wrap gap-1.5">
          {[
            { key: 'demo', label: 'Demo' },
            { key: 'delete', label: 'Delete' },
            { key: 'customer-poc', label: 'Customer POC' },
            { key: 'customer-production', label: 'Customer Production' },
            { key: 'internal-team', label: 'Internal Team' },
          ].map(({ key, label }) => {
            const isSet = orgLabels[key] === 'true'
            return (
              <button
                key={key}
                onClick={() => {
                  if (isSet) {
                    removeLabelMutation.mutate(key)
                  } else {
                    addLabelMutation.mutate({ [key]: 'true' })
                  }
                }}
                disabled={addLabelMutation.isPending || removeLabelMutation.isPending}
                className={`rounded-full px-3 py-1 text-xs font-medium border transition-colors ${
                  isSet
                    ? 'bg-primary-100 dark:bg-primary-900/40 border-primary-300 dark:border-primary-700 text-primary-700 dark:text-primary-300'
                    : 'bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700 text-gray-500 dark:text-gray-400 hover:border-gray-300 dark:hover:border-gray-600'
                }`}
              >
                {label}
              </button>
            )
          })}
        </div>
        {Object.keys(orgLabels).length > 0 ? (
          <div className="mt-2 flex flex-wrap gap-2">
            {Object.entries(orgLabels).map(([key, value]) => (
              <span key={key} className="inline-flex items-center gap-1 rounded-md bg-blue-50 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-800 px-2 py-0.5 text-xs font-mono">
                <span className="text-blue-700 dark:text-blue-300">{key}</span>
                <span className="text-blue-400">=</span>
                <span className="text-blue-600 dark:text-blue-400">{String(value)}</span>
                <button
                  onClick={() => removeLabelMutation.mutate(key)}
                  disabled={removeLabelMutation.isPending}
                  className="ml-1 text-blue-400 hover:text-red-500"
                >
                  &times;
                </button>
              </span>
            ))}
          </div>
        ) : (
          <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">No labels</p>
        )}
        <div className="mt-2 flex gap-2">
          <input
            type="text"
            value={labelKey}
            onChange={(e) => setLabelKey(e.target.value)}
            placeholder="Key"
            className="block w-32 rounded-md border-0 py-1.5 px-2.5 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 placeholder:text-gray-400"
          />
          <input
            type="text"
            value={labelValue}
            onChange={(e) => setLabelValue(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAddLabel()}
            placeholder="Value"
            className="block w-40 rounded-md border-0 py-1.5 px-2.5 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 placeholder:text-gray-400"
          />
          <button
            onClick={handleAddLabel}
            disabled={addLabelMutation.isPending}
            className="rounded-md bg-primary-600 dark:bg-primary-500 px-3 py-1.5 text-sm font-medium text-white hover:bg-primary-700 dark:hover:bg-primary-600 disabled:opacity-50"
          >
            Add
          </button>
        </div>
      </div>

      {/* Most Recent App */}
      {recent_app && (
        <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Most Recent App</h2>
          <div className="mt-2 space-y-1">
            <p className="text-sm text-gray-900 dark:text-gray-100 font-medium">{recent_app.name}</p>
            <p className="text-xs text-gray-500 dark:text-gray-400 font-mono">{recent_app.id}</p>
            {recent_app.status && (
              <Badge variant="status" status={getStatus(recent_app.status)}>{getStatus(recent_app.status)}</Badge>
            )}
          </div>
        </div>
      )}

      {/* Component Graph */}
      {graph_dot && (
        <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Component dependency graph</h2>
          <div className="mt-2">
            <DotGraph dot={graph_dot} height="28rem" />
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Actions</h2>
        <div className="mt-2 flex gap-3">
          <ActionButton mutation={supportMutation} label="Add support users" pendingLabel="Adding..." />
          <ActionButton mutation={removeOldProcessesMutation} label="Remove old runner processes" pendingLabel="Removing..." color="orange" />
          <ActionButton mutation={shutdownProcessesMutation} label="Shutdown runner processes" pendingLabel="Shutting down..." color="red" />
          <ActionButton mutation={shutdownHintProcessesMutation} label="Send shutdown hint" pendingLabel="Sending..." color="red" />
        </div>
      </div>

      {/* Installs */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Installs</h2>
        <div className="mt-2 table-card">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>ID</th>
                <th>App</th>
                <th>Runner</th>
                <th>Sandbox</th>
                <th>Components</th>
                <th>Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {installs.map((install: any) => {
                const isDeleted = !!install.deleted_at
                const runnerStatus = getStatus(install.runner_status)
                const sandboxStatus = getStatus(install.sandbox_status)
                const componentStatus = getStatus(install.composite_component_status)
                return (
                  <tr key={install.id} className={isDeleted ? 'opacity-50' : ''}>
                    <td>
                      <div className="flex items-center gap-2">
                        <Link to={`/installs/${install.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 font-medium">
                          {install.name || truncateId(install.id)}
                        </Link>
                        {isDeleted && <Badge variant="status" status="error">Deleted</Badge>}
                      </div>
                    </td>
                    <td className="text-gray-500 dark:text-gray-400 font-mono text-xs">{truncateId(install.id)}</td>
                    <td className="text-sm text-gray-700 dark:text-gray-300">
                      {install.app?.name || truncateId(install.app_id || '')}
                    </td>
                    <td>
                      {runnerStatus && (
                        <Badge variant="status" status={runnerStatus}>{runnerStatus}</Badge>
                      )}
                    </td>
                    <td>
                      {sandboxStatus && (
                        <Badge variant="status" status={sandboxStatus}>{sandboxStatus}</Badge>
                      )}
                    </td>
                    <td>
                      {componentStatus && (
                        <Badge variant="status" status={componentStatus}>{componentStatus}</Badge>
                      )}
                    </td>
                    <td className="text-gray-500 dark:text-gray-400">{formatDate(install.created_at)}</td>
                  </tr>
                )
              })}
              {installs.length === 0 && (
                <tr><td colSpan={7} className="text-center text-gray-500 dark:text-gray-400 py-6">No installs found</td></tr>
              )}
            </tbody>
          </table>
        </div>
        <Pagination page={installsPage} totalPages={installs_total_pages} onPageChange={setInstallsPage} />
      </div>
    </>
  )
}

// ---------- Queues Tab ----------
function QueuesTab({ orgId }: { orgId: string }) {
  const queryClient = useQueryClient()

  const migrateMutation = useMutation({
    mutationFn: () => migrateOrgQueues(orgId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org-queue-signal-stats', orgId] }),
  })

  const clearQueuesMutation = useMutation({
    mutationFn: () => clearOrgQueues(orgId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org-queue-signal-stats', orgId] })
      queryClient.invalidateQueries({ queryKey: ['org-queue-signals', orgId] })
    },
  })

  const forceRestartQueuesMutation = useMutation({
    mutationFn: () => forceRestartOrgQueues(orgId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org-queue-signal-stats', orgId] })
      queryClient.invalidateQueries({ queryKey: ['org-queue-signals', orgId] })
    },
  })

  const signalStatsQuery = useQuery({
    queryKey: ['org-queue-signal-stats', orgId],
    queryFn: () => getOrgQueueSignalStats(orgId),
    refetchInterval: 10000,
  })

  const queueSignalsQuery = useQuery({
    queryKey: ['org-queue-signals', orgId],
    queryFn: () => getOrgQueueSignals(orgId),
    refetchInterval: 10000,
  })

  const deleteSignalsMutation = useMutation({
    mutationFn: () => deleteOrgQueueSignals(orgId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org-queue-signals', orgId] })
      queryClient.invalidateQueries({ queryKey: ['org-queue-signal-stats', orgId] })
    },
  })

  const signals = queueSignalsQuery.data?.signals ?? []

  return (
    <div className="space-y-6">
      {/* Queue Actions */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Queue Actions</h2>
        <div className="mt-2 flex flex-wrap gap-3">
          <ActionButton mutation={migrateMutation} label="Migrate queues" pendingLabel="Migrating..." color="orange" />
          <SignalActionButton mutation={clearQueuesMutation} label="Clear org queues" pendingLabel="Enqueueing..." color="red" />
          <ActionButton mutation={forceRestartQueuesMutation} label="Force restart org queues" pendingLabel="Restarting..." color="red" />
        </div>
      </div>

      {/* Queue Signal Stats */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">
          Queue Signal Stats
          {signalStatsQuery.data && <span className="ml-2 text-xs text-gray-500 dark:text-gray-400">({signalStatsQuery.data.total} total)</span>}
        </h2>
        {signalStatsQuery.isLoading ? (
          <div className="mt-2"><LoadingSpinner /></div>
        ) : signalStatsQuery.data && signalStatsQuery.data.stats.length > 0 ? (
          <QueueSignalStatsTable stats={signalStatsQuery.data.stats} />
        ) : (
          <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">No queue signals</p>
        )}
      </div>

      {/* Queue Signals List */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">
            Recent Signals
            {signals.length > 0 && <span className="ml-2 text-xs text-gray-500 dark:text-gray-400">({signals.length})</span>}
          </h2>
          {signals.length > 0 && (
            <ConfirmButton
              mutation={deleteSignalsMutation}
              label="Delete All"
              pendingLabel="Deleting..."
              confirmMessage={`Delete all queue signals for this org?`}
            />
          )}
        </div>
        {queueSignalsQuery.isLoading ? (
          <div className="mt-2"><LoadingSpinner /></div>
        ) : (
          <div className="mt-2 table-card">
            <table>
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Type</th>
                  <th>Queue</th>
                  <th>Status</th>
                  <th>Enqueued</th>
                  <th>Executions</th>
                  <th>Created</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
                {signals.map((sig: any) => (
                  <tr key={sig.id}>
                    <td className="text-gray-500 dark:text-gray-400 font-mono text-xs">{truncateId(sig.id)}</td>
                    <td><Badge>{sig.type || '-'}</Badge></td>
                    <td className="text-gray-500 dark:text-gray-400 font-mono text-xs">
                      {sig.queue_id ? (
                        <Link to={`/queues/${sig.queue_id}`} className="text-primary-600 dark:text-primary-400 hover:underline">
                          {truncateId(sig.queue_id)}
                        </Link>
                      ) : '-'}
                    </td>
                    <td>
                      {getStatus(sig.status) && (
                        <Badge variant="status" status={getStatus(sig.status)}>{getStatus(sig.status)}</Badge>
                      )}
                    </td>
                    <td>
                      {sig.enqueued
                        ? <Badge variant="status" status="yes">Yes</Badge>
                        : <Badge variant="status" status="no">No</Badge>}
                    </td>
                    <td className="text-gray-500 dark:text-gray-400 font-mono text-xs">{sig.execution_count ?? 0}</td>
                    <td className="text-gray-500 dark:text-gray-400 text-xs">{formatDate(sig.created_at)}</td>
                  </tr>
                ))}
                {signals.length === 0 && (
                  <tr><td colSpan={7} className="text-center text-gray-500 dark:text-gray-400 py-6">No queue signals found</td></tr>
                )}
              </tbody>
            </table>
          </div>
        )}
        {deleteSignalsMutation.isSuccess && (
          <p className="mt-2 text-sm text-green-600 dark:text-green-400">
            Deleted {(deleteSignalsMutation.data as any)?.signals_deleted ?? 0} signal(s)
          </p>
        )}
      </div>
    </div>
  )
}

// ---------- Cleanup Tab ----------
function CleanupTab({ orgId, installs }: { orgId: string; installs: any[] }) {
  const queryClient = useQueryClient()

  const markDeletedMutation = useMutation({
    mutationFn: () => addOrgLabels(orgId, { delete: 'true' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', orgId] }),
  })

  const deprovisionOrgMutation = useMutation({
    mutationFn: () => deprovisionOrg(orgId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', orgId] }),
  })

  const forgetOrgMutation = useMutation({
    mutationFn: () => forgetOrg(orgId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', orgId] }),
  })

  const deprovisionAppsMutation = useMutation({
    mutationFn: () => deprovisionOrgApps(orgId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', orgId] }),
  })

  const forgetInstallsMutation = useMutation({
    mutationFn: () => forgetOrgInstalls(orgId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', orgId] }),
  })

  // Workflows
  const workflowsQuery = useQuery({
    queryKey: ['org-workflows', orgId],
    queryFn: () => getOrgWorkflows(orgId),
    refetchInterval: 10000,
  })

  const terminateWorkflowsMutation = useMutation({
    mutationFn: () => terminateOrgWorkflows(orgId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org-workflows', orgId] }),
  })

  // Per-install mutations
  const forgetInstallMutation = useMutation({
    mutationFn: (installId: string) => forgetInstall(installId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', orgId] }),
  })

  const deprovisionInstallMutation = useMutation({
    mutationFn: (installId: string) => deprovisionInstall(installId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', orgId] }),
  })

  const workflows = workflowsQuery.data?.workflows ?? []

  return (
    <div className="space-y-6">
      {/* Cleanup Actions */}
      <div className="rounded-lg border border-red-200 dark:border-red-900 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Cleanup Actions</h2>
        <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">Destructive operations for org cleanup. Use with care.</p>
        <div className="mt-3 flex flex-wrap gap-3">
          <ActionButton mutation={markDeletedMutation} label="Mark Deleted" pendingLabel="Marking..." color="orange" />
          <ActionButton mutation={deprovisionOrgMutation} label="Deprovision Org" pendingLabel="Deprovisioning..." color="red" />
          <ConfirmButton
            mutation={forgetOrgMutation}
            label="Forget Org"
            pendingLabel="Forgetting..."
            confirmMessage="Are you sure you want to forget this org? This is irreversible."
          />
          <ActionButton mutation={deprovisionAppsMutation} label="Deprovision Apps" pendingLabel="Deprovisioning..." color="red" />
          <ConfirmButton
            mutation={forgetInstallsMutation}
            label="Forget Installs"
            pendingLabel="Forgetting..."
            confirmMessage="Are you sure you want to forget all installs for this org? This is irreversible."
          />
        </div>
      </div>

      {/* Installs with per-install actions */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Installs</h2>
        <div className="mt-2 table-card">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>ID</th>
                <th>Status</th>
                <th>Created</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {installs.map((install: any) => {
                const isDeleted = !!install.deleted_at
                return (
                  <tr key={install.id} className={isDeleted ? 'opacity-50' : ''}>
                    <td>
                      <div className="flex items-center gap-2">
                        <Link to={`/installs/${install.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 font-medium">
                          {install.name || truncateId(install.id)}
                        </Link>
                        {isDeleted && <Badge variant="status" status="error">Deleted</Badge>}
                      </div>
                    </td>
                    <td className="text-gray-500 dark:text-gray-400 font-mono text-xs">{truncateId(install.id)}</td>
                    <td>
                      {getStatus(install.runner_status) && (
                        <Badge variant="status" status={getStatus(install.runner_status)}>{getStatus(install.runner_status)}</Badge>
                      )}
                    </td>
                    <td className="text-gray-500 dark:text-gray-400">{formatDate(install.created_at)}</td>
                    <td>
                      <div className="flex gap-2">
                        <button
                          onClick={() => deprovisionInstallMutation.mutate(install.id)}
                          disabled={deprovisionInstallMutation.isPending || isDeleted}
                          className="rounded-md bg-red-600 dark:bg-red-500 px-2 py-1 text-xs font-medium text-white hover:bg-red-700 dark:hover:bg-red-600 disabled:opacity-50"
                        >
                          Deprovision
                        </button>
                        <button
                          onClick={() => {
                            if (window.confirm(`Forget install ${install.name || install.id}?`)) {
                              forgetInstallMutation.mutate(install.id)
                            }
                          }}
                          disabled={forgetInstallMutation.isPending || isDeleted}
                          className="rounded-md bg-red-600 dark:bg-red-500 px-2 py-1 text-xs font-medium text-white hover:bg-red-700 dark:hover:bg-red-600 disabled:opacity-50"
                        >
                          Forget
                        </button>
                      </div>
                    </td>
                  </tr>
                )
              })}
              {installs.length === 0 && (
                <tr><td colSpan={5} className="text-center text-gray-500 dark:text-gray-400 py-6">No installs found</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Org Workflows */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">
            Org Workflows
            {workflows.length > 0 && <span className="ml-2 text-xs text-gray-500 dark:text-gray-400">({workflows.length})</span>}
          </h2>
          {workflows.length > 0 && (
            <ConfirmButton
              mutation={terminateWorkflowsMutation}
              label="Terminate All"
              pendingLabel="Terminating..."
              confirmMessage={`Terminate all ${workflows.length} workflow(s)?`}
            />
          )}
        </div>
        {workflowsQuery.isLoading ? (
          <div className="mt-2"><LoadingSpinner /></div>
        ) : (
          <div className="mt-2 table-card">
            <table>
              <thead>
                <tr>
                  <th>Workflow ID</th>
                  <th>Type</th>
                  <th>Namespace</th>
                  <th>Status</th>
                  <th>Start Time</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
                {workflows.map((wf: any) => (
                  <tr key={`${wf.namespace}-${wf.workflow_id}-${wf.run_id}`}>
                    <td className="text-gray-700 dark:text-gray-300 font-mono text-xs max-w-xs truncate" title={wf.workflow_id}>{wf.workflow_id}</td>
                    <td className="text-gray-700 dark:text-gray-300 text-xs">{wf.workflow_type}</td>
                    <td className="text-gray-500 dark:text-gray-400 text-xs">{wf.namespace}</td>
                    <td><Badge variant="status" status={wf.status}>{wf.status}</Badge></td>
                    <td className="text-gray-500 dark:text-gray-400 text-xs">{wf.start_time ? formatDate(wf.start_time) : '-'}</td>
                  </tr>
                ))}
                {workflows.length === 0 && (
                  <tr><td colSpan={5} className="text-center text-gray-500 dark:text-gray-400 py-6">No workflows found</td></tr>
                )}
              </tbody>
            </table>
          </div>
        )}
        {terminateWorkflowsMutation.isSuccess && (
          <p className="mt-2 text-sm text-green-600 dark:text-green-400">
            {(terminateWorkflowsMutation.data as any)?.message || 'Terminate signal enqueued'}
          </p>
        )}
      </div>
    </div>
  )
}

// ---------- Queue Signal Stats ----------
function QueueSignalStatsTable({ stats }: { stats: { type: string; status: string; count: number }[] }) {
  // Pivot: rows = signal types, columns = statuses
  const statuses = [...new Set(stats.map((s) => s.status))].sort()
  const typeMap = new Map<string, Map<string, number>>()
  const typeTotals = new Map<string, number>()

  for (const row of stats) {
    if (!typeMap.has(row.type)) typeMap.set(row.type, new Map())
    typeMap.get(row.type)!.set(row.status, row.count)
    typeTotals.set(row.type, (typeTotals.get(row.type) || 0) + row.count)
  }

  // Sort types by total count descending
  const types = [...typeMap.keys()].sort((a, b) => (typeTotals.get(b) || 0) - (typeTotals.get(a) || 0))

  return (
    <div className="mt-2 table-card overflow-x-auto">
      <table>
        <thead>
          <tr>
            <th>Signal Type</th>
            {statuses.map((s) => (
              <th key={s} className="text-right">{s}</th>
            ))}
            <th className="text-right">Total</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
          {types.map((type) => (
            <tr key={type}>
              <td className="text-gray-700 dark:text-gray-300 text-xs font-mono">{type}</td>
              {statuses.map((status) => {
                const count = typeMap.get(type)?.get(status) || 0
                return (
                  <td key={status} className="text-right text-xs tabular-nums">
                    {count > 0 ? (
                      <span className={
                        status === 'error' ? 'text-red-600 dark:text-red-400 font-medium' :
                        status === 'in-progress' ? 'text-blue-600 dark:text-blue-400' :
                        status === 'queued' ? 'text-yellow-600 dark:text-yellow-400' :
                        status === 'success' ? 'text-green-600 dark:text-green-400' :
                        'text-gray-700 dark:text-gray-300'
                      }>{count}</span>
                    ) : (
                      <span className="text-gray-300 dark:text-gray-700">-</span>
                    )}
                  </td>
                )
              })}
              <td className="text-right text-xs font-medium tabular-nums text-gray-900 dark:text-gray-100">{typeTotals.get(type) || 0}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

// ---------- Shared Components ----------
function SignalActionButton({ mutation, label, pendingLabel, color = 'red' }: {
  mutation: any
  label: string
  pendingLabel: string
  color?: 'primary' | 'red' | 'orange'
}) {
  const colorClasses = {
    primary: 'bg-primary-600 dark:bg-primary-500 hover:bg-primary-700 dark:hover:bg-primary-600',
    red: 'bg-red-600 dark:bg-red-500 hover:bg-red-700 dark:hover:bg-red-600',
    orange: 'bg-orange-600 hover:bg-orange-700 dark:hover:bg-orange-600',
  }

  return (
    <div>
      <button
        onClick={() => mutation.mutate()}
        disabled={mutation.isPending}
        className={`rounded-md px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50 ${colorClasses[color]}`}
      >
        {mutation.isPending ? pendingLabel : label}
      </button>
      {mutation.isSuccess && mutation.data?.signal_id && (
        <Link
          to={`/queues/${mutation.data.queue_id}/signals/${mutation.data.signal_id}`}
          className="ml-2 text-sm text-primary-600 dark:text-primary-400 hover:underline"
        >
          View signal
        </Link>
      )}
      {mutation.isError && <span className="ml-2 text-sm text-red-600 dark:text-red-400">Failed</span>}
    </div>
  )
}

function ActionButton({ mutation, label, pendingLabel, color = 'primary', successMessage }: {
  mutation: any
  label: string
  pendingLabel: string
  color?: 'primary' | 'red' | 'orange'
  successMessage?: string
}) {
  const colorClasses = {
    primary: 'bg-primary-600 dark:bg-primary-500 hover:bg-primary-700 dark:hover:bg-primary-600',
    red: 'bg-red-600 dark:bg-red-500 hover:bg-red-700 dark:hover:bg-red-600',
    orange: 'bg-orange-600 hover:bg-orange-700 dark:hover:bg-orange-600',
  }

  return (
    <div>
      <button
        onClick={() => mutation.mutate()}
        disabled={mutation.isPending}
        className={`rounded-md px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50 ${colorClasses[color]}`}
      >
        {mutation.isPending ? pendingLabel : label}
      </button>
      {mutation.isSuccess && <span className="ml-2 text-sm text-green-600 dark:text-green-400">{successMessage || 'Done'}</span>}
      {mutation.isError && <span className="ml-2 text-sm text-red-600 dark:text-red-400">Failed</span>}
    </div>
  )
}

function ConfirmButton({ mutation, label, pendingLabel, confirmMessage }: {
  mutation: any
  label: string
  pendingLabel: string
  confirmMessage: string
}) {
  return (
    <div>
      <button
        onClick={() => {
          if (window.confirm(confirmMessage)) {
            mutation.mutate()
          }
        }}
        disabled={mutation.isPending}
        className="rounded-md bg-red-600 dark:bg-red-500 px-3 py-1.5 text-sm font-medium text-white hover:bg-red-700 dark:hover:bg-red-600 disabled:opacity-50"
      >
        {mutation.isPending ? pendingLabel : label}
      </button>
      {mutation.isSuccess && <span className="ml-2 text-sm text-green-600 dark:text-green-400">Done</span>}
      {mutation.isError && <span className="ml-2 text-sm text-red-600 dark:text-red-400">Failed</span>}
    </div>
  )
}
