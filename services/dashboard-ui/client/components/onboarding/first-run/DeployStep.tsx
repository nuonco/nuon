import { useEffect, useRef, useState, type ReactNode } from 'react'
import { useForm } from '@tanstack/react-form'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormSelect } from '@/components/common/form/FormSelect'
import { FormToggle } from '@/components/common/form/FormToggle'
import { Text } from '@/components/common/Text'
import { useAuth } from '@/hooks/use-auth'
import { useFirstRun } from '@/hooks/use-first-run'
import { openSupportChat } from '@/lib/pylon-chat'
import { trackEvent } from '@/lib/posthog-analytics'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import { getWorkflowSteps, triggerBranchRun } from '@/lib'
import type { TAPIError } from '@/types'
import {
  createTestInstall,
  defaultInputs,
  findActiveAppConfig,
  getAppConfigWithInputs,
  getBranchRuns,
} from './api'
import {
  CLOUD_SANDBOX,
  DEPLOY_HELP_MESSAGE,
  DEPLOY_SLOW_MS,
  KITCHEN_SINK_VARIANTS,
  KITCHEN_SINK_APP,
  KITCHEN_SINK_LABEL,
  TRACKED_GIT_BRANCH,
  defaultRegion,
  isCloud,
  isKnownRegion,
  regionFieldLabel,
  regionOptions,
  type TCloud,
  type TPath,
} from './constants'
import { waitItems, type IWaitItem } from './progress'
import { deploySchema, type DeployValues } from './schema'
import { NextButton, RepoChip } from './shared'
import { ByocReferences, WaitProgress } from './WaitProgress'

const CONFIG_POLL_MS = 2000
const BUILDS_STEP = 'building components and sandbox'
const BUILDS_DONE = ['success', 'skipped', 'user-skipped', 'auto-skipped', 'noop']

const InstallSummaryCard = ({
  path,
  appName,
  repo,
  cloud,
}: {
  path: TPath
  appName: string
  repo: string
  cloud: TCloud
}) => {
  const own = path === 'own'
  const facts: { label: string; value: ReactNode }[] = [
    {
      label: 'Source',
      value: (
        <>
          <RepoChip repo={repo} />
          <Code variant="inline">{TRACKED_GIT_BRANCH}</Code>
        </>
      ),
    },
    {
      label: 'Nuon sandbox',
      value: (
        <>
          <RepoChip
            repo={own || cloud === 'azure' ? CLOUD_SANDBOX[cloud] : KITCHEN_SINK_VARIANTS[cloud].sandbox}
          />
          {own ? (
            <Text variant="subtext" theme="neutral">
              from sandbox.toml
            </Text>
          ) : null}
        </>
      ),
    },
    {
      label: 'Components',
      value: own
        ? 'From components/ in your repo'
        : 'Terraform modules, Helm charts, container images',
    },
  ]

  return (
    <Card className="!gap-5">
      <div className="flex items-center gap-3">
        <Icon variant={own ? 'GitBranchIcon' : 'TireIcon'} size={24} theme="brand" />
        <div className="flex flex-col">
          <Text variant="base" weight="strong">
            {own ? appName : `Nuon ${KITCHEN_SINK_LABEL} app`}
          </Text>
          <Text variant="subtext" theme="neutral">
            {own
              ? 'Your app template. This is what every install of it will contain.'
              : 'Everything you can do with Nuon, in one example app.'}
          </Text>
        </div>
      </div>
      <dl className="grid gap-4 sm:grid-cols-3">
        {facts.map((fact) => (
          <div key={fact.label} className="flex flex-col gap-1.5 rounded-md border p-3">
            <dt>
              <Text variant="subtext" theme="neutral">
                {fact.label}
              </Text>
            </dt>
            <dd className="flex flex-wrap items-center gap-1.5">
              {typeof fact.value === 'string' ? (
                <Text variant="body" weight="strong">
                  {fact.value}
                </Text>
              ) : (
                fact.value
              )}
            </dd>
          </div>
        ))}
      </dl>
    </Card>
  )
}

export type TDeployPhase = 'idle' | 'waiting-config' | 'creating'

export interface IDeployStepView {
  path: TPath
  appName: string
  repo: string
  cloud: TCloud
  region: string
  onRegion?: (region: string) => void
  onAutoApprove?: (value: boolean) => void
  autoApprove?: boolean
  phase: TDeployPhase
  installCreated: boolean
  missingInputs: string[]
  syncError?: string
  progress?: IWaitItem[]
  progressSince?: number
  retrying?: boolean
  onRetry?: () => void
  onGetHelp?: () => void
  error?: string
  onCreate: (values: DeployValues) => void
  onBack?: () => void
}

export const DeployStepView = ({
  path,
  appName,
  repo,
  cloud,
  region,
  onRegion,
  onAutoApprove,
  autoApprove = true,
  phase,
  installCreated,
  missingInputs,
  syncError,
  progress,
  progressSince,
  retrying,
  onRetry,
  onGetHelp,
  error,
  onCreate,
  onBack,
}: IDeployStepView) => {
  const schema = deploySchema(cloud)
  const form = useForm({
    defaultValues: { region, autoApprove } as DeployValues,
    validators: { onMount: schema, onChange: schema },
    onSubmit: ({ value }) => onCreate(value),
  })
  const busy = phase !== 'idle'
  const label = installCreated
    ? 'Continue'
    : phase === 'waiting-config'
      ? 'Waiting for your app config...'
      : phase === 'creating'
        ? 'Creating install...'
        : 'Create install'

  return (
    <form
      className="flex flex-col gap-6"
      autoComplete="off"
      noValidate
      onSubmit={(e) => {
        e.preventDefault()
        e.stopPropagation()
        void form.handleSubmit()
      }}
    >
      <Card className="!gap-4 !p-5">
        <div className="flex flex-col gap-1">
          <Text variant="base" weight="strong">
            Install settings
          </Text>
          <Text variant="body" theme="neutral">
            Pre-selected for a quick first run. Change anything you like.
          </Text>
        </div>
        <form.Field
          name="region"
          listeners={{ onChange: ({ value }) => onRegion?.(value) }}
        >
          {(field) => (
            <FormSelect
              field={field}
              id="first-run-region"
              options={regionOptions(cloud)}
              labelProps={{ labelText: regionFieldLabel(cloud) }}
              disabled={busy}
              searchable
              placeholder={
                cloud === 'azure'
                  ? 'Choose Azure location'
                  : cloud === 'gcp'
                    ? 'Choose GCP region'
                    : 'Choose AWS region'
              }
            />
          )}
        </form.Field>
        <form.Field
          name="autoApprove"
          listeners={{ onChange: ({ value }) => onAutoApprove?.(value) }}
        >
          {(field) => (
            <FormToggle
              field={field}
              disabled={busy}
              label="Auto-approve"
              description="Applies each plan as soon as it is ready. On by default for a faster first run."
            />
          )}
        </form.Field>
      </Card>

      <InstallSummaryCard path={path} appName={appName} repo={repo} cloud={cloud} />

      {busy && progress ? (
        <>
          <WaitProgress
            title="Getting your app config ready"
            items={progress}
            since={progressSince}
            slowAfterMs={DEPLOY_SLOW_MS}
            onGetHelp={onGetHelp}
          />
          <ByocReferences />
        </>
      ) : null}

      {missingInputs.length ? (
        <Banner theme="error">
          <div className="flex flex-col gap-1">
            <Text weight="strong">Your app config has required inputs with no default</Text>
            <Text variant="subtext">
              Add a default to {missingInputs.join(', ')} in your app config, push, then try again.
            </Text>
          </div>
        </Banner>
      ) : null}
      {phase === 'waiting-config' && syncError ? (
        <Banner theme="warn">
          <div className="flex flex-col gap-3">
            <div className="flex flex-col gap-1">
              <Text weight="strong">
                {path === 'own'
                  ? 'The last sync of your app config failed'
                  : `The ${KITCHEN_SINK_LABEL} app config did not sync`}
              </Text>
              <Text variant="subtext">
                {syncError}{' '}
                {path === 'own'
                  ? `Push a fix to ${TRACKED_GIT_BRANCH} and this page picks it up, or run the sync again.`
                  : 'Nothing in the example app needs fixing on your side. Run the sync again, or get help.'}
              </Text>
            </div>
            <div className="flex flex-wrap gap-2">
              {onRetry ? (
                <Button variant="secondary" size="sm" onClick={onRetry} disabled={retrying}>
                  {retrying ? <Icon variant="Loading" size={14} /> : <Icon variant="ArrowClockwiseIcon" size={14} />}{' '}
                  Run the sync again
                </Button>
              ) : null}
              {onGetHelp ? (
                <Button variant="ghost" size="sm" onClick={onGetHelp}>
                  <Icon variant="ChatCircleIcon" size={14} /> Get help
                </Button>
              ) : null}
            </div>
          </div>
        </Banner>
      ) : null}
      <FormErrorBanner
        error={error ? { error, description: '', user_error: true } : null}
        fallback="Unable to create the install."
      />

      <NextButton label={label} onClick={() => void form.handleSubmit()} onBack={onBack} loading={busy} />
    </form>
  )
}

const readString = (value: unknown) => (typeof value === 'string' ? value : '')

export const DeployStep = ({ sharedData, setSharedData, onAdvance, onGoBack }: IWizardStepComponentProps) => {
  const { orgId, journey } = useFirstRun()
  const { user } = useAuth()
  const path: TPath = sharedData.path === 'own' ? 'own' : 'example'
  const cloud: TCloud = isCloud(sharedData.cloud) ? sharedData.cloud : 'aws'
  const appName = readString(sharedData.app_name) || KITCHEN_SINK_APP
  const appId = readString(sharedData.app_id)
  const branchId = readString(sharedData.app_branch_id)
  const repo = readString(sharedData.repo)
  const installId = readString(sharedData.install_id)
  const savedRegion = readString(sharedData.region)
  const region = isKnownRegion(cloud, savedRegion) ? savedRegion : defaultRegion(cloud)

  const [phase, setPhase] = useState<TDeployPhase>('idle')
  const [missingInputs, setMissingInputs] = useState<string[]>([])
  const [error, setError] = useState<string>()
  const creating = useRef(false)
  const queryClient = useQueryClient()
  const [retrying, setRetrying] = useState(false)
  const choice = useRef<DeployValues>({ region, autoApprove: true })

  const { data: runs } = useQuery({
    queryKey: ['first-run-branch-runs', orgId, appId, branchId],
    queryFn: () => getBranchRuns({ orgId, appId, branchId }),
    enabled: phase === 'waiting-config' && !!appId && !!branchId,
    refetchInterval: CONFIG_POLL_MS,
  })
  const latestRun = runs?.[0]

  const { data: runSteps } = useQuery({
    queryKey: ['first-run-branch-run-steps', orgId, latestRun?.id],
    queryFn: () => getWorkflowSteps({ orgId, workflowId: latestRun!.id! }),
    enabled: phase === 'waiting-config' && !!latestRun?.id,
    refetchInterval: CONFIG_POLL_MS,
  })
  const syncError =
    latestRun?.status?.status === 'error'
      ? latestRun.status.status_human_description || 'The branch run ended in an error.'
      : undefined

  const { data: activeConfig } = useQuery({
    queryKey: ['first-run-active-config', orgId, appId, branchId],
    queryFn: () => findActiveAppConfig({ orgId, appId, branchId }).then((c) => c ?? null),
    enabled: phase === 'waiting-config' && !!appId,
    // Keep polling: a new run makes a newer config the active one.
    refetchInterval: CONFIG_POLL_MS,
  })

  // A branch-run config pins its builds, and an install provisioned before they
  // finish fails, so wait for the run's builds step too. Only a config that no
  // branch run produced (synced from the CLI) has no builds to wait for.
  const buildsStatus = runSteps?.find((step) => step.name === BUILDS_STEP)?.status?.status as string | undefined
  const buildsReady =
    (!!buildsStatus && BUILDS_DONE.includes(buildsStatus)) ||
    (!!runs && !runs.length && !!activeConfig && !activeConfig.app_branch_id)

  useEffect(() => {
    if (phase !== 'waiting-config' || !activeConfig?.id || !buildsReady || creating.current) return
    creating.current = true
    setPhase('creating')
    const { region: chosenRegion, autoApprove } = choice.current
    const run = async () => {
      try {
        const config = await getAppConfigWithInputs({ orgId, appId, appConfigId: activeConfig.id! })
        const { inputs, missing } = defaultInputs(config)
        if (missing.length) {
          trackEvent({
            event: 'install_create',
            status: 'error',
            user,
            props: { appId, path, cloud, source: 'onboarding', err: 'missing_input_defaults' },
          })
          setMissingInputs(missing)
          setPhase('idle')
          return
        }
        const install = await createTestInstall({
          orgId,
          appId,
          appName,
          branchId,
          cloud,
          region: chosenRegion,
          autoApprove,
          inputs,
        })
        const saved = {
          region: chosenRegion,
          install_id: install.id as string,
          workflow_id: (install as { workflow_id?: string }).workflow_id ?? '',
        }
        trackEvent({
          event: 'install_create',
          status: 'ok',
          user,
          props: { appId, installId: saved.install_id, path, cloud, source: 'onboarding' },
        })
        await journey.saveStep('deploy', saved, { complete: true })
        Object.entries(saved).forEach(([key, value]) => setSharedData(key, value))
        onAdvance()
      } catch (err) {
        trackEvent({
          event: 'install_create',
          status: 'error',
          user,
          props: { appId, path, cloud, source: 'onboarding', err: (err as TAPIError)?.error },
        })
        setError((err as TAPIError)?.description || 'Unable to create the install.')
        setPhase('idle')
      } finally {
        creating.current = false
      }
    }
    run()
  }, [phase, activeConfig, buildsReady, orgId, appId, appName, branchId, cloud, journey, setSharedData, onAdvance, path, user])

  const progress = waitItems(
    [
      { label: `Start a branch run on ${TRACKED_GIT_BRANCH}`, done: !!latestRun },
      { label: `Fetch the latest commit from ${repo}`, step: 'fetch commit' },
      { label: 'Read and validate the app config', step: 'fetch app config', done: !!activeConfig },
      { label: 'Build the components and sandbox', step: BUILDS_STEP, done: buildsReady },
      { label: 'Create the install', done: false },
    ],
    runSteps
  )
  const progressSince = latestRun?.created_at ? Date.parse(latestRun.created_at) : undefined

  // A forced run rebuilds every component, so a failed build gets a fresh attempt.
  const retrySync = async () => {
    setRetrying(true)
    setError(undefined)
    try {
      await triggerBranchRun({ orgId, appId, branchId, request: { force: true } })
      trackEvent({ event: 'branch_run_trigger', status: 'ok', user, props: { appId, branchId, source: 'onboarding' } })
      await queryClient.invalidateQueries({ queryKey: ['first-run-branch-runs', orgId, appId, branchId] })
    } catch (err) {
      trackEvent({
        event: 'branch_run_trigger',
        status: 'error',
        user,
        props: { appId, branchId, source: 'onboarding', err: (err as TAPIError)?.error },
      })
      setError((err as TAPIError)?.description || 'Unable to run the sync again.')
    } finally {
      setRetrying(false)
    }
  }

  const getHelp = () => {
    openSupportChat(DEPLOY_HELP_MESSAGE)
    trackEvent({ event: 'support_chat_open', status: 'ok', user, props: { source: 'onboarding_deploy' } })
  }

  // Changing a setting after the install exists drops it from the flow, so
  // the button creates a new install with the new settings.
  const clearInstall = () => {
    if (!installId) return
    setSharedData('install_id', '')
    setSharedData('workflow_id', '')
  }

  const onCreate = (values: DeployValues) => {
    choice.current = values
    if (values.region !== region) setSharedData('region', values.region)
    if (installId) {
      onAdvance()
      return
    }
    setError(undefined)
    setMissingInputs([])
    setPhase('waiting-config')
  }

  return (
    <DeployStepView
      path={path}
      appName={appName}
      repo={repo}
      cloud={cloud}
      region={region}
      onRegion={(value) => {
        if (value !== region) clearInstall()
        setSharedData('region', value)
      }}
      onAutoApprove={clearInstall}
      phase={phase}
      installCreated={!!installId}
      missingInputs={missingInputs}
      syncError={syncError}
      progress={progress}
      progressSince={progressSince}
      retrying={retrying}
      onRetry={appId && branchId ? () => void retrySync() : undefined}
      onGetHelp={getHelp}
      error={error}
      onCreate={onCreate}
      onBack={onGoBack}
    />
  )
}
