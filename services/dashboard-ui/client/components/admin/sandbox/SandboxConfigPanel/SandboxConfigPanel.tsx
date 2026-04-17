import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import type {
  TAdminSandboxConfig,
  TSandboxLogTemplate,
  TSandboxPlanTemplate,
} from '@/types'
import {
  adminUpsertSandboxConfig,
  adminResetSandboxConfigs,
} from '@/lib'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { useToast } from '@/hooks/use-toast'
import { Toast } from '@/components/surfaces/Toast'

const NS_PER_SECOND = 1_000_000_000

const nsToSeconds = (ns: number): string => {
  if (!ns) return ''
  return String(ns / NS_PER_SECOND)
}

const secondsToNs = (s: string): number => {
  const parsed = parseFloat(s)
  if (isNaN(parsed)) return 0
  return Math.round(parsed * NS_PER_SECOND)
}

const PRESETS = ['default', 'success_slow', 'success_fast', 'failure', 'failure_timeout', 'failure_panic', 'partial_failure']

const FAIL_AT_STEPS = ['', 'resetting', 'fetching', 'validate', 'initialize', 'execute', 'outputs', 'cleanup']

const JOB_CATEGORIES: { label: string; jobTypes: string[] }[] = [
  {
    label: 'Deploy',
    jobTypes: [
      'terraform-apply',
      'terraform-destroy',
      'helm-install',
      'helm-uninstall',
      'kubernetes-manifest-deploy',
      'kubernetes-manifest-destroy',
      'pulumi-up',
      'pulumi-destroy',
    ],
  },
  {
    label: 'Build',
    jobTypes: ['docker-build', 'terraform-module-build', 'helm-chart-build'],
  },
  {
    label: 'Sync',
    jobTypes: [
      'terraform-module-sync',
      'helm-chart-sync',
      'docker-build-sync',
      'container-image-sync',
      'oci-artifact-sync',
      'sync-secrets',
    ],
  },
  {
    label: 'Plan',
    jobTypes: ['create-apply-plan', 'create-teardown-plan'],
  },
  {
    label: 'Actions',
    jobTypes: ['actions-workflow'],
  },
  {
    label: 'Sandbox',
    jobTypes: ['sandbox-terraform'],
  },
  {
    label: 'Runner',
    jobTypes: [
      'health-check',
      'shut-down',
      'noop',
      'force-shut-down',
      'restart',
      'graceful-shut-down',
      'mng-update',
      'mng-shutdown',
      'mng-shutdown-vm',
      'mng-restart',
      'mng-fetch-token',
      'prune-tokens',
    ],
  },
]

const PLAN_JOB_TYPES = new Set(['create-apply-plan', 'create-teardown-plan'])

interface IConfigCardForm {
  jobType: string
  config: TAdminSandboxConfig | undefined
  logTemplates: TSandboxLogTemplate[]
  planTemplates: TSandboxPlanTemplate[]
  runnerId: string
  adminEmail: string
  onSaved: () => void
}

const ConfigCard = ({
  jobType,
  config,
  logTemplates,
  planTemplates,
  runnerId,
  adminEmail,
  onSaved,
}: IConfigCardForm) => {
  const { addToast } = useToast()
  const [preset, setPreset] = useState(config?.preset ?? 'default')
  const [duration, setDuration] = useState(nsToSeconds(config?.duration ?? 0))
  const [faultRate, setFaultRate] = useState(String(Math.round((config?.fault_rate ?? 0) * 100)))
  const [errorMessage, setErrorMessage] = useState(config?.error_message ?? '')
  const [failAtStep, setFailAtStep] = useState(config?.fail_at_step ?? '')
  const [sleepDuration, setSleepDuration] = useState(nsToSeconds(config?.sleep_duration ?? 0))
  const [timeout, setTimeout] = useState(nsToSeconds(config?.timeout ?? 0))
  const [triggerShutdown, setTriggerShutdown] = useState(config?.trigger_shutdown ?? false)
  const [logLines, setLogLines] = useState((config?.log_lines ?? []).join('\n'))
  const [planContents, setPlanContents] = useState(config?.plan_contents ?? '')

  const { mutate: upsert, isPending: isSaving } = useMutation({
    mutationFn: () =>
      adminUpsertSandboxConfig({
        runnerId,
        adminEmail,
        body: {
          job_type: jobType,
          preset,
          duration: secondsToNs(duration),
          fault_rate: parseFloat(faultRate) / 100,
          error_message: errorMessage,
          fail_at_step: failAtStep,
          sleep_duration: secondsToNs(sleepDuration),
          timeout: secondsToNs(timeout),
          trigger_shutdown: triggerShutdown,
          log_lines: logLines ? logLines.split('\n').filter(Boolean) : null,
          plan_contents: planContents,
        },
      }),
    onSuccess: () => {
      addToast(
        <Toast heading="Saved" theme="success">
          <Text>Config for {jobType} saved</Text>
        </Toast>
      )
      onSaved()
    },
    onError: () => {
      addToast(
        <Toast heading="Save failed" theme="error">
          <Text>Failed to save config for {jobType}</Text>
        </Toast>
      )
    },
  })

  const applyLogTemplate = (key: string) => {
    const tmpl = logTemplates.find((t) => t.key === key)
    if (tmpl) setLogLines(tmpl.lines.join('\n'))
  }

  const applyPlanTemplate = (key: string) => {
    const tmpl = planTemplates.find((t) => t.key === key)
    if (tmpl) setPlanContents(tmpl.contents)
  }

  const isPlanJob = PLAN_JOB_TYPES.has(jobType)
  const hasConfig = !!config

  return (
    <Card className={`gap-4 ${hasConfig ? 'border-blue-400 dark:border-blue-600' : ''}`}>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Text variant="base" weight="strong" className="font-mono">
            {jobType}
          </Text>
          {hasConfig && (
            <span className="inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full bg-blue-50 text-blue-700 border border-blue-300 dark:bg-blue-950 dark:text-blue-300 dark:border-blue-700">
              configured
            </span>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div className="flex flex-col gap-1">
          <Text variant="subtext">Preset</Text>
          <select
            className="border rounded-md px-3 py-1.5 text-sm bg-white dark:bg-dark-grey-900 dark:border-dark-grey-600"
            value={preset}
            onChange={(e) => setPreset(e.target.value)}
          >
            {PRESETS.map((p) => (
              <option key={p} value={p}>
                {p}
              </option>
            ))}
          </select>
        </div>

        <div className="flex flex-col gap-1">
          <Text variant="subtext">Fail at step</Text>
          <select
            className="border rounded-md px-3 py-1.5 text-sm bg-white dark:bg-dark-grey-900 dark:border-dark-grey-600"
            value={failAtStep}
            onChange={(e) => setFailAtStep(e.target.value)}
          >
            {FAIL_AT_STEPS.map((s) => (
              <option key={s} value={s}>
                {s || '— none —'}
              </option>
            ))}
          </select>
        </div>

        <div className="flex flex-col gap-1">
          <Text variant="subtext">Duration (seconds)</Text>
          <input
            type="number"
            min="0"
            step="0.1"
            className="border rounded-md px-3 py-1.5 text-sm bg-white dark:bg-dark-grey-900 dark:border-dark-grey-600"
            value={duration}
            onChange={(e) => setDuration(e.target.value)}
            placeholder="0"
          />
        </div>

        <div className="flex flex-col gap-1">
          <Text variant="subtext">Sleep duration (seconds)</Text>
          <input
            type="number"
            min="0"
            step="0.1"
            className="border rounded-md px-3 py-1.5 text-sm bg-white dark:bg-dark-grey-900 dark:border-dark-grey-600"
            value={sleepDuration}
            onChange={(e) => setSleepDuration(e.target.value)}
            placeholder="0"
          />
        </div>

        <div className="flex flex-col gap-1">
          <Text variant="subtext">Timeout (seconds)</Text>
          <input
            type="number"
            min="0"
            step="0.1"
            className="border rounded-md px-3 py-1.5 text-sm bg-white dark:bg-dark-grey-900 dark:border-dark-grey-600"
            value={timeout}
            onChange={(e) => setTimeout(e.target.value)}
            placeholder="0"
          />
        </div>

        <div className="flex flex-col gap-1">
          <div className="flex items-center justify-between">
            <Text variant="subtext">Fault rate: {faultRate}%</Text>
          </div>
          <input
            type="range"
            min="0"
            max="100"
            className="w-full accent-blue-600"
            value={faultRate}
            onChange={(e) => setFaultRate(e.target.value)}
          />
        </div>
      </div>

      <div className="flex flex-col gap-1">
        <Text variant="subtext">Error message</Text>
        <input
          type="text"
          className="border rounded-md px-3 py-1.5 text-sm bg-white dark:bg-dark-grey-900 dark:border-dark-grey-600"
          value={errorMessage}
          onChange={(e) => setErrorMessage(e.target.value)}
          placeholder="Custom error message"
        />
      </div>

      <div className="flex items-center gap-3">
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            className="w-4 h-4 accent-blue-600"
            checked={triggerShutdown}
            onChange={(e) => setTriggerShutdown(e.target.checked)}
          />
          <Text variant="subtext">Trigger shutdown after job</Text>
        </label>
      </div>

      <div className="flex flex-col gap-1">
        <div className="flex items-center justify-between">
          <Text variant="subtext">Log lines</Text>
          {logTemplates.length > 0 && (
            <select
              className="border rounded-md px-2 py-1 text-xs bg-white dark:bg-dark-grey-900 dark:border-dark-grey-600"
              defaultValue=""
              onChange={(e) => {
                if (e.target.value) applyLogTemplate(e.target.value)
              }}
            >
              <option value="">Use template...</option>
              {logTemplates.map((t) => (
                <option key={t.key} value={t.key}>
                  {t.name}
                </option>
              ))}
            </select>
          )}
        </div>
        <textarea
          rows={4}
          className="border rounded-md px-3 py-1.5 text-sm font-mono bg-white dark:bg-dark-grey-900 dark:border-dark-grey-600 resize-y"
          value={logLines}
          onChange={(e) => setLogLines(e.target.value)}
          placeholder="One log line per row"
        />
      </div>

      {isPlanJob && (
        <div className="flex flex-col gap-1">
          <div className="flex items-center justify-between">
            <Text variant="subtext">Plan contents</Text>
            {planTemplates.length > 0 && (
              <select
                className="border rounded-md px-2 py-1 text-xs bg-white dark:bg-dark-grey-900 dark:border-dark-grey-600"
                defaultValue=""
                onChange={(e) => {
                  if (e.target.value) applyPlanTemplate(e.target.value)
                }}
              >
                <option value="">Use template...</option>
                {planTemplates.map((t) => (
                  <option key={t.key} value={t.key}>
                    {t.name}
                  </option>
                ))}
              </select>
            )}
          </div>
          <textarea
            rows={6}
            className="border rounded-md px-3 py-1.5 text-sm font-mono bg-white dark:bg-dark-grey-900 dark:border-dark-grey-600 resize-y"
            value={planContents}
            onChange={(e) => setPlanContents(e.target.value)}
            placeholder="Plan output contents"
          />
        </div>
      )}

      <div className="flex justify-end">
        <Button variant="primary" size="sm" onClick={() => upsert()} disabled={isSaving}>
          {isSaving ? (
            <>
              <Icon variant="Loading" className="animate-spin" />
              Saving...
            </>
          ) : (
            'Save config'
          )}
        </Button>
      </div>
    </Card>
  )
}

export interface ISandboxConfigPanel {
  runnerId: string
  adminEmail: string
  configs: TAdminSandboxConfig[]
  logTemplates: TSandboxLogTemplate[]
  planTemplates: TSandboxPlanTemplate[]
  onRefresh: () => void
}

export const SandboxConfigPanel = ({
  runnerId,
  adminEmail,
  configs,
  logTemplates,
  planTemplates,
  onRefresh,
}: ISandboxConfigPanel) => {
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const configsByJobType = (jobType: string) => configs.find((c) => c.job_type === jobType)

  const { mutate: setAllPreset, isPending: isSettingAll } = useMutation({
    mutationFn: async (preset: string) => {
      const ALL_JOB_TYPES = JOB_CATEGORIES.flatMap((cat) => cat.jobTypes)
      for (const jobType of ALL_JOB_TYPES) {
        await adminUpsertSandboxConfig({
          runnerId,
          adminEmail,
          body: { job_type: jobType, preset },
        })
      }
    },
    onSuccess: (_data, preset) => {
      addToast(
        <Toast heading="Done" theme="success">
          <Text>All job types set to {preset}</Text>
        </Toast>
      )
      onRefresh()
    },
    onError: () => {
      addToast(
        <Toast heading="Failed" theme="error">
          <Text>Failed to update all configs</Text>
        </Toast>
      )
    },
  })

  const { mutate: resetAll, isPending: isResetting } = useMutation({
    mutationFn: () => adminResetSandboxConfigs({ runnerId, adminEmail }),
    onSuccess: () => {
      addToast(
        <Toast heading="Reset" theme="success">
          <Text>All sandbox configs reset</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['sandbox-configs', runnerId] })
      onRefresh()
    },
    onError: () => {
      addToast(
        <Toast heading="Reset failed" theme="error">
          <Text>Failed to reset sandbox configs</Text>
        </Toast>
      )
    },
  })

  const isActioning = isSettingAll || isResetting

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between flex-wrap gap-3">
        <Text variant="base" weight="strong">
          Job type configs
        </Text>
        <div className="flex items-center gap-2 flex-wrap">
          <Button
            variant="secondary"
            size="sm"
            disabled={isActioning}
            onClick={() => setAllPreset('success_fast')}
          >
            All success fast
          </Button>
          <Button
            variant="secondary"
            size="sm"
            disabled={isActioning}
            onClick={() => setAllPreset('failure')}
          >
            All failure
          </Button>
          <Button
            variant="danger"
            size="sm"
            disabled={isActioning}
            onClick={() => resetAll()}
          >
            {isResetting ? (
              <>
                <Icon variant="Loading" className="animate-spin" />
                Resetting...
              </>
            ) : (
              'Reset all'
            )}
          </Button>
        </div>
      </div>

      {JOB_CATEGORIES.map((category) => (
        <div key={category.label} className="flex flex-col gap-3">
          <Text variant="base" weight="strong" className="text-gray-500 dark:text-gray-400 uppercase tracking-wide text-xs">
            {category.label}
          </Text>
          <div className="flex flex-col gap-4">
            {category.jobTypes.map((jobType) => (
              <ConfigCard
                key={jobType}
                jobType={jobType}
                config={configsByJobType(jobType)}
                logTemplates={logTemplates}
                planTemplates={planTemplates}
                runnerId={runnerId}
                adminEmail={adminEmail}
                onSaved={onRefresh}
              />
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}
