import { useEffect, useRef, useState, type ReactNode } from 'react'
import { useForm } from '@tanstack/react-form'
import { useQuery } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormSelect } from '@/components/common/form/FormSelect'
import { FormToggle } from '@/components/common/form/FormToggle'
import { Text } from '@/components/common/Text'
import { useAuth } from '@/hooks/use-auth'
import { useFirstRun } from '@/hooks/use-first-run'
import { trackEvent } from '@/lib/posthog-analytics'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
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
import { deploySchema, type DeployValues } from './schema'
import { MiniArch, NextButton, RepoChip } from './shared'

const CONFIG_POLL_MS = 5000

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
          <RepoChip repo={CLOUD_SANDBOX[cloud]} />
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
      <div className="flex flex-col gap-2">
        <Text variant="subtext" theme="neutral">
          What lands in the customer&apos;s account
        </Text>
        <MiniArch />
      </div>
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
  autoApprove?: boolean
  phase: TDeployPhase
  installCreated: boolean
  missingInputs: string[]
  syncError?: string
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
  autoApprove = true,
  phase,
  installCreated,
  missingInputs,
  syncError,
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
              disabled={busy || installCreated}
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
        <form.Field name="autoApprove">
          {(field) => (
            <FormToggle
              field={field}
              disabled={busy || installCreated}
              label="Auto-approve"
              description="Applies each plan as soon as it is ready. On by default for a faster first run."
            />
          )}
        </form.Field>
      </Card>

      <InstallSummaryCard path={path} appName={appName} repo={repo} cloud={cloud} />

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
          <div className="flex flex-col gap-1">
            <Text weight="strong">The last sync of your app config failed</Text>
            <Text variant="subtext">
              {syncError} Push a fix to {TRACKED_GIT_BRANCH}; this page keeps waiting for it.
            </Text>
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
  const choice = useRef<DeployValues>({ region, autoApprove: true })

  // Install create needs an active app config, which exists only once the
  // branch has synced one from the repo.
  const { data: runs } = useQuery({
    queryKey: ['first-run-branch-runs', orgId, appId, branchId],
    queryFn: () => getBranchRuns({ orgId, appId, branchId }),
    enabled: phase === 'waiting-config' && !!appId && !!branchId,
    refetchInterval: CONFIG_POLL_MS,
  })
  const latestRun = runs?.[0]
  const syncError =
    latestRun?.status?.status === 'error'
      ? latestRun.status.status_human_description || 'The branch run ended in an error.'
      : undefined

  const { data: activeConfig } = useQuery({
    queryKey: ['first-run-active-config', orgId, appId, branchId],
    queryFn: () => findActiveAppConfig({ orgId, appId, branchId }).then((c) => c ?? null),
    enabled: phase === 'waiting-config' && !!appId,
    refetchInterval: (query) => (query.state.data ? false : CONFIG_POLL_MS),
  })

  useEffect(() => {
    if (phase !== 'waiting-config' || !activeConfig?.id || creating.current) return
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
  }, [phase, activeConfig, orgId, appId, appName, branchId, cloud, journey, setSharedData, onAdvance, path, user])

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
      onRegion={(value) => setSharedData('region', value)}
      phase={phase}
      installCreated={!!installId}
      missingInputs={missingInputs}
      syncError={syncError}
      error={error}
      onCreate={onCreate}
      onBack={onGoBack}
    />
  )
}
