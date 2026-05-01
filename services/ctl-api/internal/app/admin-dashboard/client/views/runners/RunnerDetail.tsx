import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router'
import { getRunnerDetail, upsertRunnerConfig, deleteRunnerConfig, resetRunnerConfigs } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

const SANDBOX_JOB_TYPES = [
  'terraform-deploy',
  'helm-chart-deploy',
  'kubernetes-manifest-deploy',
  'job-deploy',
  'noop-deploy',
  'docker-build',
  'container-image-build',
  'terraform-module-build',
  'helm-chart-build',
  'kubernetes-manifest-build',
  'noop-build',
  'oci-sync',
  'noop-sync',
  'fetch-image-metadata',
  'actions-workflow',
  'sandbox-terraform',
  'sandbox-terraform-plan',
  'sandbox-sync-secrets',
  'runner-helm',
  'runner-terraform',
]

const NS_PER_MS = 1_000_000

export const RunnerDetail = () => {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const [configForm, setConfigForm] = useState({
    job_type: '',
    duration_ms: 0,
    should_error: false,
    panic: false,
    trigger_shutdown: false,
  })

  const { data, isLoading, error } = useQuery({
    queryKey: ['runner', id],
    queryFn: () => getRunnerDetail(id!),
    enabled: !!id,
  })

  const upsertMutation = useMutation({
    mutationFn: () =>
      upsertRunnerConfig(id!, {
        job_type: configForm.job_type,
        duration: configForm.duration_ms * NS_PER_MS,
        should_error: configForm.should_error,
        panic: configForm.panic,
        trigger_shutdown: configForm.trigger_shutdown,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['runner', id] })
      setConfigForm({ job_type: '', duration_ms: 0, should_error: false, panic: false, trigger_shutdown: false })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (jobType: string) => deleteRunnerConfig(id!, jobType),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['runner', id] }),
  })

  const resetMutation = useMutation({
    mutationFn: () => resetRunnerConfigs(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['runner', id] }),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load runner'} />
  if (!data) return null

  const { runner, install_id, install_name, process, process_online, configs } = data

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-gray-900">{runner.name || truncateId(runner.id)}</h1>
        <p className="mt-1 text-sm text-gray-500 font-mono">{runner.id}</p>
        <div className="mt-2 flex items-center gap-3 text-sm">
          <Badge variant="status" status={process_online ? 'online' : 'offline'}>
            {process_online ? 'Online' : 'Offline'}
          </Badge>
          <span className="text-gray-500">
            Install: <Link to={`/installs/${install_id}`} className="text-primary-600 hover:text-primary-800">{install_name || truncateId(install_id)}</Link>
          </span>
        </div>
      </div>

      {/* Process Info */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Process</h2>
        {process ? (
          <dl className="mt-2 grid grid-cols-2 gap-2 text-sm">
            <div>
              <dt className="text-gray-500">Version</dt>
              <dd className="text-gray-900">{process.version}</dd>
            </div>
            <div>
              <dt className="text-gray-500">Status</dt>
              <dd><Badge variant="status" status={process.status}>{process.status}</Badge></dd>
            </div>
            <div>
              <dt className="text-gray-500">Created</dt>
              <dd className="text-gray-900">{formatDate(process.created_at)}</dd>
            </div>
          </dl>
        ) : (
          <p className="mt-2 text-sm text-gray-500">No process info available</p>
        )}
      </div>

      {/* Configs */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-semibold text-gray-900">Configs</h2>
          <button
            onClick={() => resetMutation.mutate()}
            disabled={resetMutation.isPending}
            className="rounded-md bg-red-600 px-3 py-1 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
          >
            {resetMutation.isPending ? 'Resetting...' : 'Reset All'}
          </button>
        </div>

        <div className="mt-3 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Job Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Duration</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Error</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Panic</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Shutdown</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {Object.entries(configs || {}).map(([jobType, config]) => (
                <tr key={jobType}>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{jobType}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{Math.round((config.duration ?? 0) / NS_PER_MS)}ms</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Badge variant="status" status={config.should_error ? 'error' : 'healthy'}>
                      {config.should_error ? 'Yes' : 'No'}
                    </Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Badge variant="status" status={config.panic ? 'error' : 'healthy'}>
                      {config.panic ? 'Yes' : 'No'}
                    </Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Badge variant="status" status={config.trigger_shutdown ? 'error' : 'healthy'}>
                      {config.trigger_shutdown ? 'Yes' : 'No'}
                    </Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <button
                      onClick={() => deleteMutation.mutate(jobType)}
                      disabled={deleteMutation.isPending}
                      className="text-red-600 hover:text-red-800 text-xs font-medium"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
              {Object.keys(configs || {}).length === 0 && (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-sm text-gray-500">No configs</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        {/* Upsert Form */}
        <div className="mt-4 rounded-md border border-gray-200 p-3">
          <h3 className="text-xs font-semibold text-gray-700 mb-2">Add/Update Config</h3>
          <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:flex-wrap">
            <input
              type="text"
              list="sandbox-job-types"
              value={configForm.job_type}
              onChange={(e) => setConfigForm((f) => ({ ...f, job_type: e.target.value }))}
              placeholder="Job type"
              className="block w-56 rounded-md border-0 py-1 px-2 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600 font-mono"
            />
            <datalist id="sandbox-job-types">
              {SANDBOX_JOB_TYPES.map((t) => (
                <option key={t} value={t} />
              ))}
            </datalist>
            <input
              type="number"
              value={configForm.duration_ms}
              onChange={(e) => setConfigForm((f) => ({ ...f, duration_ms: Number(e.target.value) }))}
              placeholder="Duration (ms)"
              className="block w-32 rounded-md border-0 py-1 px-2 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600"
            />
            <label className="flex items-center gap-1 text-xs text-gray-700">
              <input type="checkbox" checked={configForm.should_error} onChange={(e) => setConfigForm((f) => ({ ...f, should_error: e.target.checked }))} className="rounded border-gray-300" />
              Error
            </label>
            <label className="flex items-center gap-1 text-xs text-gray-700">
              <input type="checkbox" checked={configForm.panic} onChange={(e) => setConfigForm((f) => ({ ...f, panic: e.target.checked }))} className="rounded border-gray-300" />
              Panic
            </label>
            <label className="flex items-center gap-1 text-xs text-gray-700">
              <input type="checkbox" checked={configForm.trigger_shutdown} onChange={(e) => setConfigForm((f) => ({ ...f, trigger_shutdown: e.target.checked }))} className="rounded border-gray-300" />
              Shutdown
            </label>
            <button
              onClick={() => configForm.job_type.trim() && upsertMutation.mutate()}
              disabled={upsertMutation.isPending || !configForm.job_type.trim()}
              className="rounded-md bg-primary-600 px-3 py-1 text-sm font-medium text-white hover:bg-primary-700 disabled:opacity-50"
            >
              Save
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
