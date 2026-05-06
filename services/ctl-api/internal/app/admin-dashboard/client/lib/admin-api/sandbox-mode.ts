import { api } from '@/lib/api'
import type { TSandboxModeResponse } from '@/types/admin.types'

export const getSandboxMode = () =>
  api<TSandboxModeResponse>({ path: 'sandbox-mode' })

export const getSandboxModeRunnerJobs = () =>
  api<{ configs: any[] }>({ path: 'sandbox-mode/runner-jobs' })

export const getSandboxModeSignals = () =>
  api<{ configs: any[] }>({ path: 'sandbox-mode/signals' })

export const getSandboxModeStacks = () =>
  api<{ stacks: any[] }>({ path: 'sandbox-mode/stacks' })

export const upsertSandboxSignalConfig = (signalType: string, body: any) =>
  api<{ status: string }>({ path: `sandbox-mode/signals/${encodeURIComponent(signalType)}`, method: 'PUT', body })

export const upsertSandboxRunnerJobConfig = (jobType: string, body: any) =>
  api<{ status: string }>({ path: `sandbox-mode/runner-jobs/${encodeURIComponent(jobType)}`, method: 'PUT', body })

export const disableAllSignals = () =>
  api<{ status: string }>({ path: 'sandbox-mode/signals/disable-all', method: 'POST' })

export const deleteRunnerJobConfig = (configId: string) =>
  api<{ deleted: boolean }>({ path: `sandbox-mode/runner-jobs/${encodeURIComponent(configId)}`, method: 'DELETE' })

export const disableAllRunnerJobs = () =>
  api<{ status: string }>({ path: 'sandbox-mode/runner-jobs/disable-all', method: 'POST' })

export const applyFlowTemplate = (templateKey: string) =>
  api<{ status: string }>({ path: `sandbox-mode/templates/${encodeURIComponent(templateKey)}/apply`, method: 'POST' })
