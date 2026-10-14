import { useEffect, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Status } from '@/components/common/Status'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { AwaitAzureDetails } from '@/components/workflows/step-details/stack-details/AwaitAzureDetails'
import { AwaitGCPDetails } from '@/components/workflows/step-details/stack-details/AwaitGCPDetails'
import { InstallAppConfigProvider } from '@/providers/install-app-config-provider'
import { InstallProvider } from '@/providers/install-provider'
import { useAuth } from '@/hooks/use-auth'
import { useFirstRun } from '@/hooks/use-first-run'
import { getInstallStack, getInstallWorkflows, getWorkflowSteps } from '@/lib'
import { trackEvent } from '@/lib/posthog-analytics'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import type { TInstallStack } from '@/types'
import {
  CLOUD_CONNECT,
  CLOUD_ICON,
  CLOUD_LABEL,
  DOCS_STACKS,
  KITCHEN_SINK_LABEL,
  STACK_METHODS,
  STACK_RENDER_LABEL,
  defaultRegion,
  isCloud,
  type TCloud,
} from './constants'
import { waitItems, type IWaitItem } from './progress'
import { FirstRunCloudRegion } from './shared'
import { ByocReferences, WaitProgress } from './WaitProgress'

const STACK_POLL_MS = 2000

export type TStackPhase = 'generating' | 'ready' | 'launched' | 'done' | 'error'

export const stackStatus = (stack?: TInstallStack | null) =>
  stack?.versions?.at(0)?.composite_status?.status as string | undefined

export const phaseFromStatus = (status: string | undefined, launched: boolean): TStackPhase => {
  if (status === 'provisioning' || status === 'active') return 'done'
  if (status === 'error' || status === 'failed') return 'error'
  if (status === 'awaiting-user-run') return launched ? 'launched' : 'ready'
  return 'generating'
}

export interface IStackStepView {
  cloud: TCloud
  appName: string
  region: string
  phase: TStackPhase
  quickLinkUrl?: string
  errorDescription?: string
  details?: React.ReactNode
  progress?: IWaitItem[]
  progressSince?: number
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
  errorDescription,
  details,
  progress,
  progressSince,
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
  const failed = phase === 'error'
  const linkLaunch = cloud === 'aws' && !!quickLinkUrl

  const statusTheme = failed ? 'error' : ready || done ? 'success' : 'neutral'
  const regionName = (
    <FirstRunCloudRegion cloud={cloud} region={region} variant="subtext" theme={statusTheme} />
  )
  const status = generating ? (
    <>
      Generating the {connect.artifactNoun} for {regionName}. Usually under a minute.
    </>
  ) : ready ? (
    <>
      {connect.artifactNoun} ready for {regionName}. From launch to a healthy runner is about 11 minutes. This
      page updates on its own.
    </>
  ) : launched ? (
    `${connect.waitingHint} This page updates on its own.`
  ) : failed ? (
    `Nuon could not generate the ${connect.artifactNoun}. The deploy workflow has the details.`
  ) : (
    `${connect.stackLabel} created. Test ${connect.accountNoun} connected.`
  )

  const launchLabel = generating
    ? connect.generating
    : ready
      ? connect.launch
      : launched
        ? `Waiting for the ${connect.stackLabel}...`
        : failed
          ? `${connect.stackLabel} failed`
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
              Test {connect.accountNoun} ·{' '}
              <FirstRunCloudRegion cloud={cloud} region={region} variant="body" theme="neutral" />
            </Text>
          </div>
        </div>
        <Status
          status={failed ? 'failed' : ready || done ? 'success' : generating ? 'provisioning' : 'waiting'}
        >
          {generating ? 'Generating' : ready ? 'Ready' : done ? 'Created' : failed ? 'Failed' : 'Waiting'}
        </Status>
      </Card>

      {generating && progress ? (
        <>
          <WaitProgress title={connect.generating} items={progress} since={progressSince} />
          <ByocReferences />
        </>
      ) : null}

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
            {failed ? (
              <Icon variant="WarningCircleIcon" size={16} theme="error" weight="fill" />
            ) : ready || done ? (
              <Icon variant="CheckCircleIcon" size={16} theme="success" weight="fill" />
            ) : (
              <Icon variant="Loading" size={16} />
            )}
            <Text variant="subtext" theme={statusTheme}>
              {status}
            </Text>
          </div>
          <Link href={DOCS_STACKS} isExternal textVariant="subtext">
            All formats and CLI snippets
          </Link>
        </div>
      </Card>

      {failed && errorDescription ? <Banner theme="error">{errorDescription}</Banner> : null}

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
          {launched || failed ? (
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
  const { user } = useAuth()
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

  // The create response has no workflow ID, so find the install's provision workflow.
  const { data: workflow } = useQuery({
    queryKey: ['first-run-provision-workflow', orgId, installId],
    queryFn: () =>
      getInstallWorkflows({ orgId, installId, type: 'provision', planonly: false, limit: 1 }).then(
        (res) => res.data?.[0] ?? null
      ),
    enabled: !!installId,
    refetchInterval: (query) => (query.state.data ? false : STACK_POLL_MS),
  })
  const workflowId = workflow?.id ?? ''
  const { data: steps } = useQuery({
    queryKey: ['first-run-provision-steps', orgId, workflowId],
    queryFn: () => getWorkflowSteps({ orgId, workflowId }),
    enabled: !!workflowId && phase === 'generating',
    refetchInterval: STACK_POLL_MS,
  })
  const progress = waitItems(
    [
      { label: 'Create the install', done: true },
      { label: 'Plan the install workflow', done: !!steps?.length },
      { label: 'Generate the install state', step: 'generate install state' },
      { label: "Create the runner's identity", step: 'provision runner service account' },
      { label: STACK_RENDER_LABEL[cloud], step: 'generate install stack' },
      { label: `Publish the ${CLOUD_CONNECT[cloud].artifactNoun}`, done: phase !== 'generating' },
    ],
    steps
  )
  const progressSince = workflow?.created_at ? Date.parse(workflow.created_at) : undefined
  const version = stack?.versions?.at(0)

  const advance = () => {
    if (advanced.current) return
    advanced.current = true
    journey.saveStep('stack', {}, { complete: true }).catch(() => {})
    onAdvance()
  }
  const advanceRef = useRef(advance)
  advanceRef.current = advance

  useEffect(() => {
    if (phase === 'done') advanceRef.current()
  }, [phase])

  const showDetails = !!stack && (phase === 'launched' || phase === 'ready') && cloud !== 'aws'
  const details = showDetails ? (
    <InstallProvider installId={installId}>
      <InstallAppConfigProvider>
        <Card className="!gap-4">
          {cloud === 'gcp' ? <AwaitGCPDetails stack={stack} /> : <AwaitAzureDetails stack={stack} />}
        </Card>
      </InstallAppConfigProvider>
    </InstallProvider>
  ) : null

  return (
    <StackStepView
      cloud={cloud}
      appName={appName}
      region={region}
      phase={phase}
      quickLinkUrl={version?.quick_link_url}
      errorDescription={version?.composite_status?.status_human_description}
      details={details}
      progress={progress}
      progressSince={progressSince}
      onLaunch={() => {
        setLaunched(true)
        trackEvent({ event: 'install_stack_launch', status: 'ok', user, props: { installId, cloud } })
      }}
      onContinue={advance}
      onBack={onGoBack}
    />
  )
}
