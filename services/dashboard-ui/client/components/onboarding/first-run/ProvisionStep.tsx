import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { useFirstRun } from '@/hooks/use-first-run'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import type { TAPIError } from '@/types'
import { cn } from '@/utils/classnames'
import {
  CLI_INSTALL,
  CLI_LOGIN,
  CLOUD_CONNECT,
  CLOUD_ICON,
  KITCHEN_SINK_LABEL,
  SANDBOX_PARTS,
  buildStages,
  defaultRegion,
  isCloud,
  watchCommand,
  type IBuildStage,
  type TCloud,
  type TStageId,
} from './constants'
import { CopyTextButton, NextButton } from './shared'

const ProvisionAccountView = ({
  stages,
  cloud,
  region,
}: {
  stages: IBuildStage[]
  cloud: TCloud
  region: string
}) => {
  const [picked, setPicked] = useState<TStageId>('runner')
  const [hovered, setHovered] = useState<TStageId | null>(null)
  const focus = hovered ?? picked
  const regionClass = (id: TStageId) =>
    focus === id
      ? 'ring-2 ring-primary-500 bg-primary-50 dark:bg-primary-950/40'
      : 'ring-1 ring-neutral-200 dark:ring-neutral-700 bg-background'

  return (
    <div className="flex flex-col gap-5 md:flex-row">
      <ol className="flex shrink-0 flex-col gap-2 md:w-60" aria-label="What the install builds">
        {stages.map((stage, index) => {
          const on = focus === stage.id
          return (
            <li key={stage.id}>
              <button
                type="button"
                aria-pressed={on}
                onClick={() => setPicked(stage.id)}
                onMouseEnter={() => setHovered(stage.id)}
                onMouseLeave={() => setHovered(null)}
                onFocus={() => setHovered(stage.id)}
                onBlur={() => setHovered(null)}
                className={cn(
                  'flex w-full flex-col gap-1.5 rounded-lg px-3 py-2.5 text-left transition-colors',
                  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-500',
                  on
                    ? 'ring-1 ring-primary-500 bg-primary-50 dark:bg-primary-950/40'
                    : 'ring-1 ring-neutral-200 dark:ring-neutral-700 hover:bg-neutral-50 dark:hover:bg-neutral-900'
                )}
              >
                <span className="flex items-center gap-2.5">
                  <span
                    className={cn(
                      'flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[11px] font-semibold ring-2',
                      on
                        ? 'bg-primary-50 text-primary-800 ring-primary-500 dark:bg-primary-950 dark:text-primary-200'
                        : 'bg-background text-neutral-500 ring-neutral-200 dark:ring-neutral-700'
                    )}
                  >
                    {index + 1}
                  </span>
                  <Text variant="body" weight="strong" className="min-w-0 flex-1">
                    {stage.label}
                  </Text>
                </span>
                {on ? (
                  <Text variant="subtext" theme="neutral" className="pl-8">
                    {stage.blurb}
                  </Text>
                ) : null}
              </button>
            </li>
          )
        })}
      </ol>

      <div className="flex min-w-0 flex-1 flex-col gap-3 rounded-xl p-4 ring-1 ring-neutral-200 dark:ring-neutral-700">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <Icon variant={CLOUD_ICON[cloud]} size={18} />
            <Text variant="base" weight="strong">
              Your test {CLOUD_CONNECT[cloud].accountNoun}
            </Text>
          </div>
          <Badge size="sm" variant="code">
            {region}
          </Badge>
        </div>

        <div className={cn('flex flex-col rounded-lg px-3.5 py-3 transition-colors', regionClass('runner'))}>
          <Text variant="body" weight="strong">
            Nuon runner
          </Text>
          <Text variant="subtext" theme="neutral">
            Starts on the machine your stack created and builds everything below from inside your account.
          </Text>
        </div>

        <div className={cn('flex flex-col gap-3 rounded-lg p-3.5 transition-colors', regionClass('sandbox'))}>
          <Text variant="body" weight="strong">
            Nuon sandbox
          </Text>
          <div className="flex flex-wrap gap-2">
            {SANDBOX_PARTS.map((part) => (
              <span
                key={part}
                className="rounded-md bg-background px-2 py-1 font-mono text-xs ring-1 ring-neutral-300 dark:ring-neutral-600"
              >
                {part}
              </span>
            ))}
          </div>
          <div
            className={cn('flex flex-col rounded-lg px-3.5 py-3 transition-colors', regionClass('components'))}
          >
            <Text variant="body" weight="strong">
              Your components
            </Text>
            <Text variant="subtext" theme="neutral">
              Terraform, Helm charts and images land here once the sandbox is up.
            </Text>
          </div>
        </div>
      </div>
    </div>
  )
}

export interface IProvisionStepView {
  cloud: TCloud
  appName: string
  region: string
  installId: string
  finishing?: boolean
  error?: string
  onFinish: () => void
  onBack?: () => void
}

// A static preview of what the install builds, in order. Live progress lives on
// the install's workflow page, which the primary opens.
export const ProvisionStepView = ({
  cloud,
  appName,
  region,
  installId,
  finishing,
  error,
  onFinish,
  onBack,
}: IProvisionStepView) => {
  const command = watchCommand(installId)
  return (
    <div className="flex flex-col gap-6">
      <Card className="!gap-5">
        <ProvisionAccountView stages={buildStages(cloud, appName)} cloud={cloud} region={region} />
        <div className="flex flex-col gap-3 border-t pt-5">
          <div className="flex flex-col gap-1">
            <Text variant="body" weight="strong">
              See it in action in the CLI
            </Text>
            <Text variant="subtext" theme="neutral" flex className="flex-wrap">
              Needs the Nuon CLI:
              <Badge size="sm" variant="code">
                {CLI_INSTALL}
              </Badge>
              then
              <Badge size="sm" variant="code">
                {CLI_LOGIN}
              </Badge>
            </Text>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Badge size="sm" variant="code">
              {command}
            </Badge>
            <CopyTextButton text={command} label="Copy" size="sm" />
          </div>
        </div>
      </Card>
      {error ? <Banner theme="error">{error}</Banner> : null}
      <NextButton label="Go to deploy workflow" onClick={onFinish} onBack={onBack} loading={finishing} />
    </div>
  )
}

const readString = (value: unknown) => (typeof value === 'string' ? value : '')

export const ProvisionStep = ({ sharedData, onGoBack }: IWizardStepComponentProps) => {
  const { orgId, journey } = useFirstRun()
  const cloud: TCloud = isCloud(sharedData.cloud) ? sharedData.cloud : 'aws'
  const installId = readString(sharedData.install_id)
  const workflowId = readString(sharedData.workflow_id)
  const appName = sharedData.path === 'own' ? readString(sharedData.app_name) : KITCHEN_SINK_LABEL
  const region = readString(sharedData.region) || defaultRegion(cloud)
  const [finishing, setFinishing] = useState(false)
  const [error, setError] = useState<string>()

  const finish = async () => {
    setFinishing(true)
    setError(undefined)
    try {
      await journey.complete()
      window.location.assign(
        workflowId
          ? `/${orgId}/installs/${installId}/history/${workflowId}`
          : `/${orgId}/installs/${installId}/history`
      )
    } catch (err) {
      setError((err as TAPIError)?.description || 'Unable to finish onboarding.')
      setFinishing(false)
    }
  }

  return (
    <ProvisionStepView
      cloud={cloud}
      appName={appName}
      region={region}
      installId={installId}
      finishing={finishing}
      error={error}
      onFinish={finish}
      onBack={onGoBack}
    />
  )
}
