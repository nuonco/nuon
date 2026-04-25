import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import {
  getSandboxMode,
  upsertSandboxRunnerJobConfig,
  upsertSandboxSignalConfig,
  disableAllSignals,
  disableAllRunnerJobs,
  applyFlowTemplate,
} from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate } from '@/utils/format'

type Tab = 'runner-jobs' | 'signals' | 'stacks' | 'templates'

export const SandboxMode = () => {
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<Tab>('runner-jobs')
  const [editingJob, setEditingJob] = useState<string | null>(null)
  const [editingSignal, setEditingSignal] = useState<string | null>(null)
  const [jobForm, setJobForm] = useState({ duration: 0, should_error: false, panic: false, trigger_shutdown: false })
  const [signalForm, setSignalForm] = useState({ frequency: '', is_disabled: false })

  const { data, isLoading, error } = useQuery({
    queryKey: ['sandbox-mode'],
    queryFn: () => getSandboxMode(),
  })

  const upsertJobMutation = useMutation({
    mutationFn: ({ jobType, body }: { jobType: string; body: any }) => upsertSandboxRunnerJobConfig(jobType, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sandbox-mode'] })
      setEditingJob(null)
    },
  })

  const upsertSignalMutation = useMutation({
    mutationFn: ({ signalType, body }: { signalType: string; body: any }) => upsertSandboxSignalConfig(signalType, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sandbox-mode'] })
      setEditingSignal(null)
    },
  })

  const disableSignalsMutation = useMutation({
    mutationFn: () => disableAllSignals(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['sandbox-mode'] }),
  })

  const disableJobsMutation = useMutation({
    mutationFn: () => disableAllRunnerJobs(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['sandbox-mode'] }),
  })

  const applyTemplateMutation = useMutation({
    mutationFn: (key: string) => applyFlowTemplate(key),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['sandbox-mode'] }),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load sandbox mode'} />
  if (!data) return null

  const { runner_job_configs = [], signal_configs = [], stacks = [], templates = [] } = data

  const tabs: { key: Tab; label: string }[] = [
    { key: 'runner-jobs', label: 'Runner Jobs' },
    { key: 'signals', label: 'Signals' },
    { key: 'stacks', label: 'Stacks' },
    { key: 'templates', label: 'Templates' },
  ]

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900">Sandbox Mode</h1>

      {/* Tab Navigation */}
      <div className="mt-4 border-b border-gray-200">
        <nav className="flex -mb-px space-x-8">
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={`whitespace-nowrap border-b-2 py-3 px-1 text-sm font-medium ${
                activeTab === tab.key
                  ? 'border-primary-500 text-primary-600'
                  : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      <div className="mt-4">
        {/* Runner Jobs Tab */}
        {activeTab === 'runner-jobs' && (
          <div>
            <div className="mb-3">
              <button
                onClick={() => disableJobsMutation.mutate()}
                disabled={disableJobsMutation.isPending}
                className="rounded-md bg-red-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
              >
                {disableJobsMutation.isPending ? 'Disabling...' : 'Disable All'}
              </button>
            </div>
            <div className="overflow-x-auto">
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
                  {runner_job_configs.map((config) => (
                    <tr key={config.id}>
                      <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{config.job_type}</td>
                      {editingJob === config.job_type ? (
                        <>
                          <td className="px-4 py-3">
                            <input type="number" value={jobForm.duration} onChange={(e) => setJobForm((f) => ({ ...f, duration: Number(e.target.value) }))} className="w-24 rounded border-gray-300 text-sm" />
                          </td>
                          <td className="px-4 py-3">
                            <input type="checkbox" checked={jobForm.should_error} onChange={(e) => setJobForm((f) => ({ ...f, should_error: e.target.checked }))} />
                          </td>
                          <td className="px-4 py-3">
                            <input type="checkbox" checked={jobForm.panic} onChange={(e) => setJobForm((f) => ({ ...f, panic: e.target.checked }))} />
                          </td>
                          <td className="px-4 py-3">
                            <input type="checkbox" checked={jobForm.trigger_shutdown} onChange={(e) => setJobForm((f) => ({ ...f, trigger_shutdown: e.target.checked }))} />
                          </td>
                          <td className="px-4 py-3 flex gap-1">
                            <button onClick={() => upsertJobMutation.mutate({ jobType: config.job_type, body: jobForm })} className="text-xs text-primary-600 hover:text-primary-800 font-medium">Save</button>
                            <button onClick={() => setEditingJob(null)} className="text-xs text-gray-500 hover:text-gray-700">Cancel</button>
                          </td>
                        </>
                      ) : (
                        <>
                          <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{config.duration}ms</td>
                          <td className="whitespace-nowrap px-4 py-3 text-sm">{config.should_error ? 'Yes' : 'No'}</td>
                          <td className="whitespace-nowrap px-4 py-3 text-sm">{config.panic ? 'Yes' : 'No'}</td>
                          <td className="whitespace-nowrap px-4 py-3 text-sm">{config.trigger_shutdown ? 'Yes' : 'No'}</td>
                          <td className="whitespace-nowrap px-4 py-3 text-sm">
                            <button
                              onClick={() => {
                                setEditingJob(config.job_type)
                                setJobForm({ duration: config.duration, should_error: config.should_error, panic: config.panic, trigger_shutdown: config.trigger_shutdown })
                              }}
                              className="text-xs text-primary-600 hover:text-primary-800 font-medium"
                            >
                              Edit
                            </button>
                          </td>
                        </>
                      )}
                    </tr>
                  ))}
                  {runner_job_configs.length === 0 && (
                    <tr>
                      <td colSpan={6} className="px-4 py-8 text-center text-sm text-gray-500">No runner job configs</td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* Signals Tab */}
        {activeTab === 'signals' && (
          <div>
            <div className="mb-3">
              <button
                onClick={() => disableSignalsMutation.mutate()}
                disabled={disableSignalsMutation.isPending}
                className="rounded-md bg-red-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
              >
                {disableSignalsMutation.isPending ? 'Disabling...' : 'Disable All'}
              </button>
            </div>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Signal Type</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Frequency</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Disabled</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200 bg-white">
                  {signal_configs.map((config) => (
                    <tr key={config.id}>
                      <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{config.signal_type}</td>
                      {editingSignal === config.signal_type ? (
                        <>
                          <td className="px-4 py-3">
                            <input type="text" value={signalForm.frequency} onChange={(e) => setSignalForm((f) => ({ ...f, frequency: e.target.value }))} className="w-24 rounded border-gray-300 text-sm" />
                          </td>
                          <td className="px-4 py-3">
                            <input type="checkbox" checked={signalForm.is_disabled} onChange={(e) => setSignalForm((f) => ({ ...f, is_disabled: e.target.checked }))} />
                          </td>
                          <td className="px-4 py-3 flex gap-1">
                            <button onClick={() => upsertSignalMutation.mutate({ signalType: config.signal_type, body: signalForm })} className="text-xs text-primary-600 hover:text-primary-800 font-medium">Save</button>
                            <button onClick={() => setEditingSignal(null)} className="text-xs text-gray-500 hover:text-gray-700">Cancel</button>
                          </td>
                        </>
                      ) : (
                        <>
                          <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{config.frequency}</td>
                          <td className="whitespace-nowrap px-4 py-3 text-sm">
                            <Badge variant="status" status={config.is_disabled ? 'offline' : 'online'}>
                              {config.is_disabled ? 'Disabled' : 'Enabled'}
                            </Badge>
                          </td>
                          <td className="whitespace-nowrap px-4 py-3 text-sm">
                            <button
                              onClick={() => {
                                setEditingSignal(config.signal_type)
                                setSignalForm({ frequency: config.frequency, is_disabled: config.is_disabled })
                              }}
                              className="text-xs text-primary-600 hover:text-primary-800 font-medium"
                            >
                              Edit
                            </button>
                          </td>
                        </>
                      )}
                    </tr>
                  ))}
                  {signal_configs.length === 0 && (
                    <tr>
                      <td colSpan={4} className="px-4 py-8 text-center text-sm text-gray-500">No signal configs</td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* Stacks Tab */}
        {activeTab === 'stacks' && (
          <div>
            {stacks.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 bg-white">
                    {stacks.map((stack: any, i: number) => (
                      <tr key={stack.id || i}>
                        <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 font-mono">{stack.id || '-'}</td>
                        <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{stack.name || '-'}</td>
                        <td className="whitespace-nowrap px-4 py-3 text-sm">
                          {stack.status ? <Badge variant="status" status={stack.status}>{stack.status}</Badge> : '-'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <p className="text-sm text-gray-500">No stacks</p>
            )}
          </div>
        )}

        {/* Templates Tab */}
        {activeTab === 'templates' && (
          <div>
            {templates.length > 0 ? (
              <div className="space-y-2">
                {templates.map((template: any, i: number) => (
                  <div key={template.key || i} className="flex items-center justify-between rounded-lg border border-gray-200 bg-white p-3">
                    <div>
                      <p className="text-sm font-medium text-gray-900">{template.name || template.key || `Template ${i + 1}`}</p>
                      {template.description && (
                        <p className="text-xs text-gray-500">{template.description}</p>
                      )}
                    </div>
                    <button
                      onClick={() => applyTemplateMutation.mutate(template.key)}
                      disabled={applyTemplateMutation.isPending}
                      className="rounded-md bg-primary-600 px-3 py-1 text-sm font-medium text-white hover:bg-primary-700 disabled:opacity-50"
                    >
                      Apply
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-gray-500">No templates</p>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
