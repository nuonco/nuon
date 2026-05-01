import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useSearchParams } from 'react-router'
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

type Tab = 'runner-jobs' | 'signals' | 'stacks' | 'templates'

const defaultJobForm = {
  operation: '',
  enabled: true,
  duration_ms: 1000,
  sleep_duration_ms: 0,
  should_error: false,
  panic: false,
  trigger_shutdown: false,
  log_template: '',
  plan_template: '',
  plan_display_template: '',
  state_template: '',
  output_template: '',
}

const defaultSignalForm = {
  enabled: true,
  deadlock_sleep_seconds: 0,
  workflow_sleep_seconds: 0,
  panic: false,
  error: '',
  validate_error: '',
}

export const SandboxMode = () => {
  const queryClient = useQueryClient()
  const [searchParams, setSearchParams] = useSearchParams()
  const tabParam = searchParams.get('tab') as Tab | null
  const validTabs: Tab[] = ['runner-jobs', 'signals', 'stacks', 'templates']
  const activeTab: Tab = tabParam && validTabs.includes(tabParam) ? tabParam : 'runner-jobs'
  const setActiveTab = (t: Tab) => {
    const next = new URLSearchParams(searchParams)
    next.set('tab', t)
    setSearchParams(next, { replace: true })
  }

  // Runner jobs state - editingJobType is the job_type being edited (or '__new__' for new)
  const [editingJobType, setEditingJobType] = useState<string | null>(null)
  const [selectedNewJobType, setSelectedNewJobType] = useState('')
  const [jobForm, setJobForm] = useState(defaultJobForm)

  // Signals state
  const [editingSignalType, setEditingSignalType] = useState<string | null>(null)
  const [selectedNewSignalType, setSelectedNewSignalType] = useState('')
  const [signalForm, setSignalForm] = useState(defaultSignalForm)

  const { data, isLoading, error } = useQuery({
    queryKey: ['sandbox-mode'],
    queryFn: () => getSandboxMode(),
  })

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['sandbox-mode'] })

  const upsertJobMutation = useMutation({
    mutationFn: ({ jobType, body }: { jobType: string; body: any }) => upsertSandboxRunnerJobConfig(jobType, body),
    onSuccess: () => { invalidate(); setEditingJobType(null) },
  })

  const upsertSignalMutation = useMutation({
    mutationFn: ({ signalType, body }: { signalType: string; body: any }) => upsertSandboxSignalConfig(signalType, body),
    onSuccess: () => { invalidate(); setEditingSignalType(null) },
  })

  const disableSignalsMutation = useMutation({ mutationFn: disableAllSignals, onSuccess: invalidate })
  const disableJobsMutation = useMutation({ mutationFn: disableAllRunnerJobs, onSuccess: invalidate })
  const applyTemplateMutation = useMutation({ mutationFn: applyFlowTemplate, onSuccess: invalidate })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load sandbox mode'} />
  if (!data) return null

  const runnerJobConfigs = data.runner_job_configs || []
  const signalConfigs = data.signal_configs || []
  const stackConfig = data.stack_config
  const flowTemplates = data.flow_templates || []
  const allTemplates = data.templates || []
  const allRunnerJobTypes: string[] = data.all_runner_job_types || []
  const allSignalTypes: string[] = data.all_signal_types || []

  const templateOptions = allTemplates.map((t: any) => t.key || t.Key || String(t))

  const openNewJob = () => {
    setSelectedNewJobType('')
    setJobForm(defaultJobForm)
    setEditingJobType('__new__')
  }

  const openEditJob = (config: any) => {
    setJobForm({
      operation: config.operation || '',
      enabled: config.enabled ?? true,
      duration_ms: config.duration ? config.duration / 1_000_000 : 0,
      sleep_duration_ms: config.sleep_duration ? config.sleep_duration / 1_000_000 : 0,
      should_error: config.should_error ?? false,
      panic: config.panic ?? false,
      trigger_shutdown: config.trigger_shutdown ?? false,
      log_template: config.log_template || '',
      plan_template: config.plan_template || '',
      plan_display_template: config.plan_display_template || '',
      state_template: config.state_template || '',
      output_template: config.output_template || '',
    })
    setEditingJobType(config.job_type)
  }

  const saveJob = () => {
    const jobType = editingJobType === '__new__' ? selectedNewJobType : editingJobType
    if (!jobType) return
    upsertJobMutation.mutate({ jobType, body: jobForm })
  }

  const openNewSignal = () => {
    setSelectedNewSignalType('')
    setSignalForm(defaultSignalForm)
    setEditingSignalType('__new__')
  }

  const openEditSignal = (config: any) => {
    setSignalForm({
      enabled: config.enabled ?? false,
      deadlock_sleep_seconds: config.deadlock_sleep ? config.deadlock_sleep / 1_000_000_000 : 0,
      workflow_sleep_seconds: config.workflow_sleep ? config.workflow_sleep / 1_000_000_000 : 0,
      panic: config.panic ?? false,
      error: config.error || '',
      validate_error: config.validate_error || '',
    })
    setEditingSignalType(config.signal_type)
  }

  const saveSignal = () => {
    const signalType = editingSignalType === '__new__' ? selectedNewSignalType : editingSignalType
    if (!signalType) return
    upsertSignalMutation.mutate({ signalType, body: signalForm })
  }

  const tabs: { key: Tab; label: string }[] = [
    { key: 'runner-jobs', label: 'Runner jobs' },
    { key: 'signals', label: 'Signals' },
    { key: 'stacks', label: 'Stacks' },
    { key: 'templates', label: 'Templates' },
  ]

  return (
    <div>
      <h1 className="page-heading">Sandbox mode</h1>

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
            <div className="mb-3 flex gap-2">
              <button onClick={openNewJob} className="rounded-md bg-primary-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-primary-700">
                New override
              </button>
              <button
                onClick={() => disableJobsMutation.mutate()}
                disabled={disableJobsMutation.isPending}
                className="rounded-md bg-red-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
              >
                Disable all
              </button>
            </div>

            {/* Edit/New form panel */}
            {editingJobType && (
              <JobFormPanel
                isNew={editingJobType === '__new__'}
                jobType={editingJobType === '__new__' ? selectedNewJobType : editingJobType}
                allTypes={allRunnerJobTypes}
                onJobTypeChange={setSelectedNewJobType}
                form={jobForm}
                setForm={setJobForm}
                templateOptions={templateOptions}
                onSave={saveJob}
                onCancel={() => setEditingJobType(null)}
                isPending={upsertJobMutation.isPending}
              />
            )}

            <div className="table-card">
              <table>
                <thead>
                  <tr>
                    <th>Job type</th>
                    <th>Enabled</th>
                    <th>Duration</th>
                    <th>Sleep</th>
                    <th>Error</th>
                    <th>Panic</th>
                    <th>Shutdown</th>
                    <th>Templates</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {runnerJobConfigs.map((config: any) => (
                    <tr key={config.id || config.job_type} className={editingJobType === config.job_type ? 'bg-primary-50' : ''}>
                      <td className="font-mono text-xs text-gray-900">{config.job_type}{config.operation ? ` (${config.operation})` : ''}</td>
                      <td><Badge variant="status" status={config.enabled ? 'online' : 'offline'}>{config.enabled ? 'Yes' : 'No'}</Badge></td>
                      <td className="font-mono text-xs text-gray-500">{config.duration ? `${config.duration / 1_000_000}ms` : '-'}</td>
                      <td className="font-mono text-xs text-gray-500">{config.sleep_duration ? `${config.sleep_duration / 1_000_000}ms` : '-'}</td>
                      <td className="text-xs">{config.should_error ? '✓' : '-'}</td>
                      <td className="text-xs">{config.panic ? '✓' : '-'}</td>
                      <td className="text-xs">{config.trigger_shutdown ? '✓' : '-'}</td>
                      <td className="text-xs text-gray-500">
                        {[config.log_template, config.plan_template, config.state_template, config.output_template].filter(Boolean).join(', ') || '-'}
                      </td>
                      <td>
                        <button
                          onClick={() => openEditJob(config)}
                          className="text-xs text-primary-600 hover:text-primary-800 font-medium"
                        >
                          Edit
                        </button>
                      </td>
                    </tr>
                  ))}
                  {runnerJobConfigs.length === 0 && (
                    <tr><td colSpan={9} className="text-center text-gray-500 py-6">No runner job overrides</td></tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* Signals Tab */}
        {activeTab === 'signals' && (
          <div>
            <div className="mb-3 flex gap-2">
              <button onClick={openNewSignal} className="rounded-md bg-primary-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-primary-700">
                New override
              </button>
              <button
                onClick={() => disableSignalsMutation.mutate()}
                disabled={disableSignalsMutation.isPending}
                className="rounded-md bg-red-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
              >
                Disable all
              </button>
            </div>

            {/* Edit/New form panel */}
            {editingSignalType && (
              <SignalFormPanel
                isNew={editingSignalType === '__new__'}
                signalType={editingSignalType === '__new__' ? selectedNewSignalType : editingSignalType}
                allTypes={allSignalTypes}
                onSignalTypeChange={setSelectedNewSignalType}
                form={signalForm}
                setForm={setSignalForm}
                onSave={saveSignal}
                onCancel={() => setEditingSignalType(null)}
                isPending={upsertSignalMutation.isPending}
              />
            )}

            <div className="table-card">
              <table>
                <thead>
                  <tr>
                    <th>Signal type</th>
                    <th>Enabled</th>
                    <th>Deadlock sleep</th>
                    <th>Workflow sleep</th>
                    <th>Panic</th>
                    <th>Error</th>
                    <th>Validate error</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {signalConfigs.map((config: any) => (
                    <tr key={config.id || config.signal_type} className={editingSignalType === config.signal_type ? 'bg-primary-50' : ''}>
                      <td className="font-mono text-xs text-gray-900">{config.signal_type}</td>
                      <td><Badge variant="status" status={config.enabled ? 'online' : 'offline'}>{config.enabled ? 'Yes' : 'No'}</Badge></td>
                      <td className="font-mono text-xs text-gray-500">{config.deadlock_sleep ? `${config.deadlock_sleep / 1_000_000_000}s` : '-'}</td>
                      <td className="font-mono text-xs text-gray-500">{config.workflow_sleep ? `${config.workflow_sleep / 1_000_000_000}s` : '-'}</td>
                      <td className="text-xs">{config.panic ? '✓' : '-'}</td>
                      <td className="text-xs text-gray-500 max-w-[120px] truncate">{config.error || '-'}</td>
                      <td className="text-xs text-gray-500 max-w-[120px] truncate">{config.validate_error || '-'}</td>
                      <td>
                        <button
                          onClick={() => openEditSignal(config)}
                          className="text-xs text-primary-600 hover:text-primary-800 font-medium"
                        >
                          Edit
                        </button>
                      </td>
                    </tr>
                  ))}
                  {signalConfigs.length === 0 && (
                    <tr><td colSpan={8} className="text-center text-gray-500 py-6">No signal overrides</td></tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* Stacks Tab */}
        {activeTab === 'stacks' && (
          <div>
            {stackConfig ? (
              <div className="rounded-lg border border-gray-200 bg-white p-4">
                <h3 className="text-sm font-medium text-gray-900">Sandbox terraform config</h3>
                <dl className="mt-3 grid grid-cols-2 gap-3 text-sm">
                  <div><dt className="text-gray-500">Job type</dt><dd className="font-mono">{stackConfig.job_type}</dd></div>
                  <div><dt className="text-gray-500">Enabled</dt><dd><Badge variant="status" status={stackConfig.enabled ? 'online' : 'offline'}>{stackConfig.enabled ? 'Yes' : 'No'}</Badge></dd></div>
                  <div><dt className="text-gray-500">Duration</dt><dd className="font-mono">{stackConfig.duration ? `${stackConfig.duration / 1_000_000}ms` : '-'}</dd></div>
                  <div><dt className="text-gray-500">Should error</dt><dd>{stackConfig.should_error ? 'Yes' : 'No'}</dd></div>
                </dl>
              </div>
            ) : (
              <p className="text-sm text-gray-500">No stack config</p>
            )}
          </div>
        )}

        {/* Templates Tab */}
        {activeTab === 'templates' && (
          <div>
            {flowTemplates.length > 0 ? (
              <div className="space-y-2">
                {flowTemplates.map((template: any, i: number) => (
                  <div key={template.key || template.Key || i} className="flex items-center justify-between rounded-lg border border-gray-200 bg-white p-4">
                    <div>
                      <p className="text-sm font-medium text-gray-900">{template.name || template.Name || template.key || template.Key}</p>
                      {(template.description || template.Description) && (
                        <p className="mt-0.5 text-xs text-gray-500">{template.description || template.Description}</p>
                      )}
                    </div>
                    <button
                      onClick={() => applyTemplateMutation.mutate(template.key || template.Key)}
                      disabled={applyTemplateMutation.isPending}
                      className="rounded-md bg-primary-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-primary-700 disabled:opacity-50"
                    >
                      Apply
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-gray-500">No flow templates</p>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

// -- Form panels --

function JobFormPanel({
  isNew,
  jobType,
  allTypes,
  onJobTypeChange,
  form,
  setForm,
  templateOptions,
  onSave,
  onCancel,
  isPending,
}: {
  isNew: boolean
  jobType: string
  allTypes: string[]
  onJobTypeChange: (v: string) => void
  form: typeof defaultJobForm
  setForm: React.Dispatch<React.SetStateAction<typeof defaultJobForm>>
  templateOptions: string[]
  onSave: () => void
  onCancel: () => void
  isPending: boolean
}) {
  return (
    <div className="mb-4 rounded-lg border border-primary-200 bg-primary-50/50 p-4">
      <h3 className="text-sm font-semibold text-gray-900 mb-3">
        {isNew ? 'New runner job override' : `Edit: ${jobType}`}
      </h3>
      <div className="grid grid-cols-2 gap-x-4 gap-y-3 sm:grid-cols-3 lg:grid-cols-4">
        {isNew && (
          <div className="col-span-2">
            <Field label="Job type">
              <TypeCombobox value={jobType} onChange={onJobTypeChange} options={allTypes} placeholder="Search job types..." />
            </Field>
          </div>
        )}
        <Field label="Enabled">
          <input type="checkbox" checked={form.enabled} onChange={(e) => setForm((f) => ({ ...f, enabled: e.target.checked }))} className="mt-1" />
        </Field>
        <Field label="Duration (ms)">
          <DurationMsInput value={form.duration_ms} onChange={(v) => setForm((f) => ({ ...f, duration_ms: v }))} />
        </Field>
        <Field label="Sleep (ms)">
          <DurationMsInput value={form.sleep_duration_ms} onChange={(v) => setForm((f) => ({ ...f, sleep_duration_ms: v }))} />
        </Field>
        <Field label="Should error">
          <input type="checkbox" checked={form.should_error} onChange={(e) => setForm((f) => ({ ...f, should_error: e.target.checked }))} className="mt-1" />
        </Field>
        <Field label="Panic">
          <input type="checkbox" checked={form.panic} onChange={(e) => setForm((f) => ({ ...f, panic: e.target.checked }))} className="mt-1" />
        </Field>
        <Field label="Trigger shutdown">
          <input type="checkbox" checked={form.trigger_shutdown} onChange={(e) => setForm((f) => ({ ...f, trigger_shutdown: e.target.checked }))} className="mt-1" />
        </Field>
        <Field label="Operation">
          <input type="text" value={form.operation} onChange={(e) => setForm((f) => ({ ...f, operation: e.target.value }))} placeholder="optional" className="w-full rounded-md border-gray-300 text-sm py-1.5 px-2" />
        </Field>
        <Field label="Log template">
          <TemplateSelect value={form.log_template} onChange={(v) => setForm((f) => ({ ...f, log_template: v }))} options={templateOptions} />
        </Field>
        <Field label="Plan template">
          <TemplateSelect value={form.plan_template} onChange={(v) => setForm((f) => ({ ...f, plan_template: v }))} options={templateOptions} />
        </Field>
        <Field label="Plan display template">
          <TemplateSelect value={form.plan_display_template} onChange={(v) => setForm((f) => ({ ...f, plan_display_template: v }))} options={templateOptions} />
        </Field>
        <Field label="State template">
          <TemplateSelect value={form.state_template} onChange={(v) => setForm((f) => ({ ...f, state_template: v }))} options={templateOptions} />
        </Field>
        <Field label="Output template">
          <TemplateSelect value={form.output_template} onChange={(v) => setForm((f) => ({ ...f, output_template: v }))} options={templateOptions} />
        </Field>
      </div>
      <div className="mt-3 flex gap-2">
        <button onClick={onSave} disabled={isPending || (isNew && !jobType)} className="rounded-md bg-primary-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-primary-700 disabled:opacity-50">
          {isPending ? 'Saving...' : 'Save'}
        </button>
        <button onClick={onCancel} className="rounded-md px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-100">Cancel</button>
      </div>
    </div>
  )
}

function SignalFormPanel({
  isNew,
  signalType,
  allTypes,
  onSignalTypeChange,
  form,
  setForm,
  onSave,
  onCancel,
  isPending,
}: {
  isNew: boolean
  signalType: string
  allTypes: string[]
  onSignalTypeChange: (v: string) => void
  form: typeof defaultSignalForm
  setForm: React.Dispatch<React.SetStateAction<typeof defaultSignalForm>>
  onSave: () => void
  onCancel: () => void
  isPending: boolean
}) {
  return (
    <div className="mb-4 rounded-lg border border-primary-200 bg-primary-50/50 p-4">
      <h3 className="text-sm font-semibold text-gray-900 mb-3">
        {isNew ? 'New signal override' : `Edit: ${signalType}`}
      </h3>
      <div className="grid grid-cols-2 gap-x-4 gap-y-3 sm:grid-cols-3 lg:grid-cols-4">
        {isNew && (
          <div className="col-span-2">
            <Field label="Signal type">
              <TypeCombobox value={signalType} onChange={onSignalTypeChange} options={allTypes} placeholder="Search signal types..." />
            </Field>
          </div>
        )}
        <Field label="Enabled">
          <input type="checkbox" checked={form.enabled} onChange={(e) => setForm((f) => ({ ...f, enabled: e.target.checked }))} className="mt-1" />
        </Field>
        <Field label="Deadlock sleep (sec)">
          <DurationSecInput value={form.deadlock_sleep_seconds} onChange={(v) => setForm((f) => ({ ...f, deadlock_sleep_seconds: v }))} />
        </Field>
        <Field label="Workflow sleep (sec)">
          <DurationSecInput value={form.workflow_sleep_seconds} onChange={(v) => setForm((f) => ({ ...f, workflow_sleep_seconds: v }))} />
        </Field>
        <Field label="Panic">
          <input type="checkbox" checked={form.panic} onChange={(e) => setForm((f) => ({ ...f, panic: e.target.checked }))} className="mt-1" />
        </Field>
        <Field label="Error message">
          <input type="text" value={form.error} onChange={(e) => setForm((f) => ({ ...f, error: e.target.value }))} placeholder="Execute error" className="w-full rounded-md border-gray-300 text-sm py-1.5 px-2" />
        </Field>
        <Field label="Validate error">
          <input type="text" value={form.validate_error} onChange={(e) => setForm((f) => ({ ...f, validate_error: e.target.value }))} placeholder="Validate error" className="w-full rounded-md border-gray-300 text-sm py-1.5 px-2" />
        </Field>
      </div>
      <div className="mt-3 flex gap-2">
        <button onClick={onSave} disabled={isPending || (isNew && !signalType)} className="rounded-md bg-primary-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-primary-700 disabled:opacity-50">
          {isPending ? 'Saving...' : 'Save'}
        </button>
        <button onClick={onCancel} className="rounded-md px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-100">Cancel</button>
      </div>
    </div>
  )
}

// -- Shared form helpers --

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <label className="block text-xs font-medium text-gray-700 mb-1">{label}</label>
      {children}
    </div>
  )
}

function TemplateSelect({ value, onChange, options }: { value: string; onChange: (v: string) => void; options: string[] }) {
  return (
    <select value={value} onChange={(e) => onChange(e.target.value)} className="w-full rounded-md border-gray-300 text-sm py-1.5 px-2">
      <option value="">None</option>
      {options.map((t) => <option key={t} value={t}>{t}</option>)}
    </select>
  )
}

/** Filterable type selector - type to search, click to select */
function TypeCombobox({ value, onChange, options, placeholder }: { value: string; onChange: (v: string) => void; options: string[]; placeholder: string }) {
  const [filter, setFilter] = useState(value)
  const [open, setOpen] = useState(false)

  const filtered = filter
    ? options.filter((o) => o.toLowerCase().includes(filter.toLowerCase()))
    : options

  return (
    <div className="relative">
      <input
        type="text"
        value={filter}
        onChange={(e) => { setFilter(e.target.value); setOpen(true) }}
        onFocus={() => setOpen(true)}
        placeholder={placeholder}
        className="w-full rounded-md border-gray-300 text-sm py-1.5 px-2 font-mono"
      />
      {open && filtered.length > 0 && (
        <div className="absolute z-10 mt-1 max-h-48 w-full overflow-y-auto rounded-md border border-gray-200 bg-white shadow-lg">
          {filtered.map((opt) => (
            <button
              key={opt}
              type="button"
              onClick={() => { onChange(opt); setFilter(opt); setOpen(false) }}
              className={`block w-full text-left px-2 py-1.5 text-xs font-mono hover:bg-primary-50 ${opt === value ? 'bg-primary-50 text-primary-700 font-medium' : 'text-gray-700'}`}
            >
              {opt}
            </button>
          ))}
        </div>
      )}
      {open && (
        <div className="fixed inset-0 z-[5]" onClick={() => setOpen(false)} />
      )}
    </div>
  )
}

const msPresets = [
  { label: '1s', value: 1_000 },
  { label: '5s', value: 5_000 },
  { label: '1m', value: 60_000 },
  { label: '2m', value: 120_000 },
  { label: '5m', value: 300_000 },
]

const secPresets = [
  { label: '1s', value: 1 },
  { label: '5s', value: 5 },
  { label: '1m', value: 60 },
  { label: '2m', value: 120 },
  { label: '5m', value: 300 },
]

/** Duration input in milliseconds with preset buttons */
function DurationMsInput({ value, onChange }: { value: number; onChange: (v: number) => void }) {
  return (
    <div>
      <input type="number" value={value} onChange={(e) => onChange(Number(e.target.value))} className="w-full rounded-md border-gray-300 text-sm py-1.5 px-2" />
      <div className="mt-1 flex gap-1">
        {msPresets.map((p) => (
          <button
            key={p.label}
            type="button"
            onClick={() => onChange(p.value)}
            className={`rounded px-1.5 py-0.5 text-[10px] font-medium border transition-colors ${value === p.value ? 'bg-primary-100 border-primary-300 text-primary-700' : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'}`}
          >
            {p.label}
          </button>
        ))}
      </div>
    </div>
  )
}

/** Duration input in seconds with preset buttons */
function DurationSecInput({ value, onChange }: { value: number; onChange: (v: number) => void }) {
  return (
    <div>
      <input type="number" step="0.1" value={value} onChange={(e) => onChange(Number(e.target.value))} className="w-full rounded-md border-gray-300 text-sm py-1.5 px-2" />
      <div className="mt-1 flex gap-1">
        {secPresets.map((p) => (
          <button
            key={p.label}
            type="button"
            onClick={() => onChange(p.value)}
            className={`rounded px-1.5 py-0.5 text-[10px] font-medium border transition-colors ${value === p.value ? 'bg-primary-100 border-primary-300 text-primary-700' : 'bg-gray-50 border-gray-200 text-gray-500 hover:bg-gray-100'}`}
          >
            {p.label}
          </button>
        ))}
      </div>
    </div>
  )
}
