import { useEffect, useRef, useState, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Select } from '@/components/common/form/Select'
import { Toggle } from '@/components/common/form/Toggle'
import { Text } from '@/components/common/Text'
import { useFirstRun } from '@/hooks/use-first-run'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import type { TAPIError } from '@/types'
import {
  createTestInstall,
  defaultInputs,
  findActiveAppConfig,
  getAppConfigWithInputs,
} from './api'
import {
  CLOUD_REGIONS,
  CLOUD_SANDBOX,
  KITCHEN_SINK_APP,
  KITCHEN_SINK_LABEL,
  TRACKED_GIT_BRANCH,
  defaultRegion,
  isCloud,
  regionOptions,
  type TCloud,
  type TPath,
} from './constants'
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
          <Badge size="sm" variant="code">
            {TRACKED_GIT_BRANCH}
          </Badge>
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
  onRegion: (region: string) => void
  autoApprove: boolean
  onAutoApprove: (value: boolean) => void
  phase: TDeployPhase
  installCreated: boolean
  missingInputs: string[]
  error?: string
  onCreate: () => void
  onBack?: () => void
}

export const DeployStepView = ({
  path,
  appName,
  repo,
  cloud,
  region,
  onRegion,
  autoApprove,
  onAutoApprove,
  phase,
  installCreated,
  missingInputs,
  error,
  onCreate,
  onBack,
}: IDeployStepView) => {
  const busy = phase !== 'idle'
  const label = installCreated
    ? 'Continue'
    : phase === 'waiting-config'
      ? 'Waiting for your app config...'
      : phase === 'creating'
        ? 'Creating install...'
        : 'Create install'

  return (
    <div className="flex flex-col gap-6">
      <Card className="!gap-4 !p-5">
        <div className="flex flex-col gap-1">
          <Text variant="base" weight="strong">
            Install settings
          </Text>
          <Text variant="body" theme="neutral">
            Pre-selected for a quick first run. Change anything you like.
          </Text>
        </div>
        <Select
          id="first-run-region"
          options={regionOptions(cloud)}
          labelProps={{ labelText: CLOUD_REGIONS[cloud].label }}
          value={region}
          onChange={onRegion}
          disabled={busy || installCreated}
        />
        <Toggle
          checked={autoApprove}
          onChange={onAutoApprove}
          disabled={busy || installCreated}
          label="Auto-approve"
          description="Applies each plan as soon as it is ready. On by default for a faster first run."
        />
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
      {error ? <Banner theme="error">{error}</Banner> : null}

      <NextButton label={label} onClick={onCreate} onBack={onBack} loading={busy} />
    </div>
  )
}

const readString = (value: unknown) => (typeof value === 'string' ? value : '')

export const DeployStep = ({ sharedData, setSharedData, onAdvance, onGoBack }: IWizardStepComponentProps) => {
  const { orgId, journey } = useFirstRun()
  const path: TPath = sharedData.path === 'own' ? 'own' : 'example'
  const cloud: TCloud = isCloud(sharedData.cloud) ? sharedData.cloud : 'aws'
  const appName = readString(sharedData.app_name) || KITCHEN_SINK_APP
  const appId = readString(sharedData.app_id)
  const branchId = readString(sharedData.app_branch_id)
  const repo = readString(sharedData.repo)
  const installId = readString(sharedData.install_id)
  const region = CLOUD_REGIONS[cloud].options.includes(readString(sharedData.region))
    ? readString(sharedData.region)
    : defaultRegion(cloud)

  const [autoApprove, setAutoApprove] = useState(true)
  const [phase, setPhase] = useState<TDeployPhase>('idle')
  const [missingInputs, setMissingInputs] = useState<string[]>([])
  const [error, setError] = useState<string>()
  const creating = useRef(false)

  // Install create needs an active app config, which exists only once the
  // branch has synced one from the repo.
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
    const run = async () => {
      try {
        const config = await getAppConfigWithInputs({ orgId, appId, appConfigId: activeConfig.id! })
        const { inputs, missing } = defaultInputs(config)
        if (missing.length) {
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
          region,
          autoApprove,
          inputs,
        })
        const saved = {
          region,
          install_id: install.id as string,
          workflow_id: (install as { workflow_id?: string }).workflow_id ?? '',
        }
        await journey.saveStep('deploy', saved, { complete: true })
        Object.entries(saved).forEach(([key, value]) => setSharedData(key, value))
        onAdvance()
      } catch (err) {
        setError((err as TAPIError)?.description || 'Unable to create the install.')
        setPhase('idle')
      } finally {
        creating.current = false
      }
    }
    run()
  }, [phase, activeConfig, orgId, appId, appName, branchId, cloud, region, autoApprove, journey, setSharedData, onAdvance])

  const onCreate = () => {
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
      autoApprove={autoApprove}
      onAutoApprove={setAutoApprove}
      phase={phase}
      installCreated={!!installId}
      missingInputs={missingInputs}
      error={error}
      onCreate={onCreate}
      onBack={onGoBack}
    />
  )
}
