import { useEffect, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { AwaitAzureDetailsComponent } from '@/components/workflows/step-details/stack-details/AwaitAzureDetails'
import { AwaitGCPDetailsComponent } from '@/components/workflows/step-details/stack-details/AwaitGCPDetails'
import { useFirstRun } from '@/hooks/use-first-run'
import { getInstallStack } from '@/lib'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import type { TInstallStack } from '@/types'
import {
  CLOUD_CONNECT,
  CLOUD_ICON,
  CLOUD_LABEL,
  DOCS_STACKS,
  KITCHEN_SINK_LABEL,
  STACK_METHODS,
  defaultRegion,
  isCloud,
  type TCloud,
} from './constants'

const STACK_POLL_MS = 3000

// The stack version's lifecycle, as the step shows it. `launched` is local: the
// user opened the link or copied the commands and the stack has not reported back.
export type TStackPhase = 'generating' | 'ready' | 'launched' | 'done'

export const stackStatus = (stack?: TInstallStack | null) =>
  stack?.versions?.at(0)?.composite_status?.status as string | undefined

export const phaseFromStatus = (status: string | undefined, launched: boolean): TStackPhase => {
  if (status === 'provisioning' || status === 'active') return 'done'
  if (status === 'awaiting-user-run') return launched ? 'launched' : 'ready'
  return 'generating'
}

export interface IStackStepView {
  cloud: TCloud
  appName: string
  region: string
  phase: TStackPhase
  quickLinkUrl?: string
  details?: React.ReactNode
  onLaunch: () => void
  onContinue: () => void
  onBack?: () => void
}

export const StackStepView = ({
  cloud,
  appName,
  region,
  phase,
  quickLinkUrl,
  details,
  onLaunch,
  onContinue,
  onBack,
}: IStackStepView) => {
  const connect = CLOUD_CONNECT[cloud]
  const methods = STACK_METHODS[cloud]
  const generating = phase === 'generating'
  const ready = phase === 'ready'
  const launched = phase === 'launched'
  const done = phase === 'done'
  const linkLaunch = cloud === 'aws' && !!quickLinkUrl

  const status = generating
    ? `Generating the ${connect.artifactNoun} for ${region}. About 30 seconds.`
    : ready
      ? `${connect.artifactNoun} ready for ${region}. From launch to a healthy runner is about 11 minutes. This page updates on its own.`
      : launched
        ? `${connect.waitingHint} This page updates on its own.`
        : `${connect.stackLabel} created. Test ${connect.accountNoun} connected.`

  const launchLabel = generating
    ? connect.generating
    : ready
      ? connect.launch
      : launched
        ? `Waiting for the ${connect.stackLabel}...`
        : `${connect.stackLabel} created`

  return (
    <div className="flex flex-col gap-6">
      <Card className="!gap-0 !p-4 !flex-row items-center justify-between">
        <div className="flex items-center gap-3">
          <Icon variant={CLOUD_ICON[cloud]} size={24} />
          <div className="flex flex-col">
            <Text variant="base" weight="strong">
              {connect.stackLabel} for {appName}
            </Text>
            <Text variant="body" theme="neutral">
              Test {connect.accountNoun} · {region}
            </Text>
          </div>
        </div>
        <Badge size="sm" theme={ready || done ? 'success' : 'brand'}>
          {generating ? 'Generating' : ready ? 'Ready' : done ? 'Created' : 'Waiting'}
        </Badge>
      </Card>

      <Card className="!gap-5">
        <div className="flex flex-col gap-1">
          <Text variant="h3" role="heading" level={3}>
            How your customers create this install
          </Text>
          <Text variant="body" theme="neutral">
            {cloud === 'gcp'
              ? 'On Google Cloud, Nuon renders the install stack in Terraform.'
              : `Nuon renders the install stack in Terraform and in ${CLOUD_LABEL[cloud]}'s native format.`}{' '}
            Your customer creates it with their own credentials; that is how access is granted. You are about
            to do it the way they would.
          </Text>
        </div>
        <ul className="flex flex-col divide-y rounded-md border">
          {methods.map((method, index) => (
            <li key={method.name} className="flex flex-wrap items-baseline gap-x-3 gap-y-1 px-4 py-3">
              <Text variant="body" weight="strong">
                {method.name}
              </Text>
              {index === 0 ? (
                <Badge size="sm" theme="brand">
                  This install
                </Badge>
              ) : null}
              <Text variant="subtext" theme="neutral">
                {method.how}
              </Text>
            </li>
          ))}
        </ul>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            {ready || done ? (
              <Icon variant="CheckCircleIcon" size={16} theme="success" weight="fill" />
            ) : (
              <Icon variant="Loading" size={16} />
            )}
            <Text variant="subtext" theme={ready || done ? 'success' : 'neutral'}>
              {status}
            </Text>
          </div>
          <Link href={DOCS_STACKS} isExternal textVariant="subtext">
            All formats and CLI snippets
          </Link>
        </div>
      </Card>

      {details}

      <div className="flex flex-wrap items-center justify-between gap-3">
        {onBack && (generating || ready) ? (
          <Button variant="secondary" size="lg" onClick={onBack}>
            <Icon variant="CaretLeftIcon" weight="bold" /> Back
          </Button>
        ) : (
          <span />
        )}
        <div className="flex flex-wrap items-center gap-3">
          {launched ? (
            <Button variant="secondary" size="lg" onClick={onContinue}>
              Continue <Icon variant="CaretRightIcon" weight="bold" />
            </Button>
          ) : null}
          {linkLaunch && ready ? (
            <Button
              variant="primary"
              size="lg"
              href={quickLinkUrl}
              target="_blank"
              rel="noreferrer"
              onClick={onLaunch}
            >
              {launchLabel} <Icon variant="ArrowSquareOutIcon" size={14} />
            </Button>
          ) : (
            <Button
              variant="primary"
              size="lg"
              disabled={!ready}
              onClick={onLaunch}
              tooltipProps={
                generating
                  ? { tipContent: `Cannot launch until Nuon finishes generating the ${connect.artifactNoun}` }
                  : undefined
              }
            >
              {generating || launched ? <Icon variant="Loading" size={16} /> : null}
              {launchLabel}
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}

const readString = (value: unknown) => (typeof value === 'string' ? value : '')

export const StackStep = ({ sharedData, onAdvance, onGoBack }: IWizardStepComponentProps) => {
  const { orgId, journey } = useFirstRun()
  const cloud: TCloud = isCloud(sharedData.cloud) ? sharedData.cloud : 'aws'
  const installId = readString(sharedData.install_id)
  const region = readString(sharedData.region) || defaultRegion(cloud)
  const appName =
    sharedData.path === 'own' ? readString(sharedData.app_name) : KITCHEN_SINK_LABEL
  const [launched, setLaunched] = useState(false)
  const advanced = useRef(false)

  const { data: stack } = useQuery({
    queryKey: ['first-run-install-stack', orgId, installId],
    queryFn: () => getInstallStack({ orgId, installId }),
    enabled: !!installId,
    refetchInterval: STACK_POLL_MS,
  })
  const phase = phaseFromStatus(stackStatus(stack), launched)
  const version = stack?.versions?.at(0)

  const advance = () => {
    if (advanced.current) return
    advanced.current = true
    journey.saveStep('stack', {}, { complete: true }).catch(() => {})
    onAdvance()
  }
  const advanceRef = useRef(advance)
  advanceRef.current = advance

  // Phone-home moves the user on without a click.
  useEffect(() => {
    if (phase === 'done') advanceRef.current()
  }, [phase])

  const showDetails = !!stack && (phase === 'launched' || phase === 'ready') && cloud !== 'aws'
  const details = showDetails ? (
    <Card className="!gap-4">
      {cloud === 'gcp' ? (
        <AwaitGCPDetailsComponent
          stack={stack}
          orgId={orgId}
          installId={installId}
          gcpRegion={region}
        />
      ) : (
        <AwaitAzureDetailsComponent
          stack={stack}
          orgId={orgId}
          installId={installId}
          azureLocation={region}
        />
      )}
    </Card>
  ) : null

  return (
    <StackStepView
      cloud={cloud}
      appName={appName}
      region={region}
      phase={phase}
      quickLinkUrl={version?.quick_link_url}
      details={details}
      onLaunch={() => setLaunched(true)}
      onContinue={advance}
      onBack={onGoBack}
    />
  )
}
