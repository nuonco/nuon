import { useRef, useState, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Code } from '@/components/common/Code'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { Text } from '@/components/common/Text'
import { githubAppInstallUrl } from '@/components/vcs-connections/ConnectGithub'
import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { useFirstRun } from '@/hooks/use-first-run'
import { getVCSConnectionRepos, getVCSConnections } from '@/lib'
import { trackEvent } from '@/lib/posthog-analytics'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import type { TAPIError, TVCSConnectionRepo } from '@/types'
import { cn } from '@/utils/classnames'
import { AppNameTakenError, setUpKitchenSink, setUpOwnApp } from './api'
import {
  APP_NAME_PATTERN,
  APP_NAME_RULE,
  CLI_SETUP,
  CLOUD_ICON,
  CLOUD_LABEL,
  EXAMPLE_APP_FACTS,
  EXAMPLE_CLOUDS,
  KITCHEN_SINK_REPO,
  KITCHEN_SINK_URL,
  defaultRegion,
  isCloud,
  type TCloud,
  type TExampleCloud,
} from './constants'
import { NextButton, TestCloudPicker } from './shared'

const OWN_APP_STEPS = [
  { icon: 'GitHub', title: 'Connect GitHub' },
  { icon: 'RobotIcon', title: 'Connect your app' },
  { icon: 'CloudIcon', title: 'Create the first install' },
] as const

// The repo the app template's config lives in: one named after the app, or the
// only repo the connection can see.
export const pickConfigRepo = (repos: TVCSConnectionRepo[] | undefined, appName: string) => {
  if (!repos?.length) return undefined
  return repos.find((repo) => repo.name === appName) ?? (repos.length === 1 ? repos[0] : undefined)
}

export type TGithubTile =
  | { status: 'disconnected'; error?: boolean }
  | { status: 'connected'; owner?: string; repoCount?: number }

const ExampleAppDrawer = () => {
  const [open, setOpen] = useState(false)
  return (
    <div className="flex flex-col rounded-md border border-dashed">
      <button
        type="button"
        aria-expanded={open}
        aria-controls="example-app-drawer"
        onClick={() => setOpen((prev) => !prev)}
        className="flex w-full items-center justify-between gap-3 px-4 py-2.5 text-left hover:bg-neutral-50 dark:hover:bg-neutral-900"
      >
        <span className="flex items-center gap-2">
          <Icon variant="GithubLogoIcon" size={16} theme="neutral" />
          <Text variant="subtext" weight="strong">
            See the example app repo
          </Text>
          <Code variant="inline">{KITCHEN_SINK_REPO}</Code>
        </span>
        <span className={cn('flex transition-transform duration-300', open && 'rotate-180')} aria-hidden>
          <Icon variant="CaretDownIcon" size={14} weight="bold" theme="neutral" />
        </span>
      </button>
      <div
        id="example-app-drawer"
        className={cn(
          'grid transition-[grid-template-rows,opacity,visibility] duration-300 ease-out',
          open ? 'visible grid-rows-[1fr] opacity-100' : 'invisible grid-rows-[0fr] opacity-0'
        )}
        aria-hidden={!open}
      >
        <div className="overflow-hidden">
          <div className="flex flex-col gap-3 border-t border-dashed px-4 py-3 md:flex-row md:items-start md:justify-between">
            <div className="flex flex-col gap-1.5">
              <Text variant="body" weight="strong">
                Kitchen Sink
              </Text>
              <ul className="flex flex-col gap-1">
                {EXAMPLE_APP_FACTS.map((fact) => (
                  <li key={fact} className="flex items-start gap-2">
                    <Icon
                      variant="CheckCircleIcon"
                      size={14}
                      weight="fill"
                      theme="success"
                      className="mt-0.5 shrink-0"
                    />
                    <Text variant="subtext" theme="neutral">
                      {fact}
                    </Text>
                  </li>
                ))}
              </ul>
            </div>
            <Button variant="secondary" size="sm" href={KITCHEN_SINK_URL} target="_blank" rel="noreferrer">
              <Icon variant="GithubLogoIcon" size={14} /> View on GitHub
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

const ExampleEscapeHatch = ({ onExit }: { onExit: () => void }) => (
  <div className="flex flex-wrap items-center justify-between gap-3 rounded-md border border-dashed px-4 py-3">
    <div className="flex items-center gap-2">
      <Icon variant="TireIcon" size={16} theme="neutral" />
      <Text variant="subtext" theme="neutral">
        Want to see an install work before touching your repo?
      </Text>
    </div>
    <Button variant="ghost" size="sm" onClick={onExit}>
      Use the example app <Icon variant="ArrowRightIcon" size={14} />
    </Button>
  </div>
)

const GithubTile = ({
  github,
  connectHref,
  onConnect,
  connecting,
  showErrors,
}: {
  github: TGithubTile
  connectHref?: string
  onConnect: () => void
  connecting: boolean
  showErrors: boolean
}) => {
  const connected = github.status === 'connected'
  const failed = github.status === 'disconnected' && github.error
  const missing = showErrors && !connected

  return (
    // Ring, not border, so the required/connected/error edge can change color
    // (the global border-color rule would paint a border grey regardless).
    <div
      className={cn(
        'flex flex-col gap-3 rounded-md p-4 ring-1 transition-shadow',
        connected
          ? 'ring-green-500 dark:ring-green-400'
          : missing || failed
            ? 'ring-red-500 dark:ring-red-400'
            : 'ring-neutral-200 dark:ring-neutral-700'
      )}
    >
      <div className="flex flex-col gap-2 items-start">
        {connected ? (
          <Text variant="body" theme="success" flex className="flex-wrap">
            <Icon variant="CheckCircleIcon" size={18} weight="fill" theme="success" />
            Connected as
            {github.owner ? (
              <Code variant="inline">{github.owner}</Code>
            ) : null}
            {github.repoCount !== undefined
              ? `· ${github.repoCount} ${github.repoCount === 1 ? 'repo' : 'repos'}`
              : null}
          </Text>
        ) : (
          <Button
            variant="secondary"
            size="lg"
            disabled={connecting || !connectHref}
            onClick={onConnect}
            tooltipProps={!connectHref ? { tipContent: 'GitHub App is not configured' } : undefined}
          >
            {connecting ? (
              'Connecting GitHub...'
            ) : (
              <>
                <Icon variant="GitHub" size={18} />
                Connect GitHub
              </>
            )}
          </Button>
        )}
        <Badge size="sm" theme={connected ? 'success' : 'brand'}>
          {connected ? 'Connected' : 'Required'}
        </Badge>
      </div>
      <Text variant="subtext" theme="neutral">
        Nuon integrates with your git workflow and syncs your app template.
      </Text>
      {failed ? (
        <Text variant="subtext" theme="error" flex>
          <Icon variant="WarningCircleIcon" size={14} weight="fill" />
          GitHub did not finish connecting. Try again.
        </Text>
      ) : missing ? (
        <Text variant="subtext" theme="error" flex>
          <Icon variant="WarningCircleIcon" size={14} weight="fill" />
          Connect GitHub to continue.
        </Text>
      ) : null}
    </div>
  )
}

export interface IStartStepView {
  expanded: boolean
  onExpand: () => void
  onBackToIntro: () => void
  onExitToExample: () => void
  appName: string
  onAppName: (name: string) => void
  // A server-side rejection of the name (e.g. already taken).
  appNameError?: string
  repoError?: string
  github: TGithubTile
  connectHref?: string
  onConnectGithub: () => void
  connectingGithub?: boolean
  cloud?: TCloud
  onCloud: (cloud: TCloud) => void
  showErrors: boolean
  onNext: () => void
  nextPending?: boolean
  // Set while something Next depends on is still loading.
  nextBlockedReason?: string
  onDeployExample: (cloud: TExampleCloud) => void
  examplePending?: TExampleCloud
  error?: string
}

export const StartStepView = ({
  expanded,
  onExpand,
  onBackToIntro,
  onExitToExample,
  appName,
  onAppName,
  appNameError,
  repoError,
  github,
  connectHref,
  onConnectGithub,
  connectingGithub = false,
  cloud,
  onCloud,
  showErrors,
  onNext,
  nextPending,
  nextBlockedReason,
  onDeployExample,
  examplePending,
  error,
}: IStartStepView) => {
  const named = appName.length > 0
  const nameInvalid = named && !APP_NAME_PATTERN.test(appName)
  const nameError = nameInvalid
    ? APP_NAME_RULE
    : appNameError ??
      repoError ??
      (showErrors && !named ? 'Name your app template to continue.' : undefined)

  let body: ReactNode
  if (expanded) {
    body = (
      <div className="flex flex-col gap-6 scroll-mt-6">
        <Card className="!gap-4">
          <Text variant="h2" role="heading" level={2}>
            Set up
          </Text>
          <div className="grid gap-4 md:grid-cols-2">
            <GithubTile
              github={github}
              connectHref={connectHref}
              onConnect={onConnectGithub}
              connecting={connectingGithub}
              showErrors={showErrors}
            />
            <div className="flex flex-col gap-3 rounded-md border p-4">
              <div className="flex items-center gap-2">
                <Icon variant="TerminalWindowIcon" size={20} theme="neutral" />
                <Text variant="body" weight="strong">
                  Install the CLI and log in
                </Text>
              </div>
              <CodeBlock language="bash" showCopy>
                {CLI_SETUP}
              </CodeBlock>
            </div>
          </div>
        </Card>

        <Card className="!gap-4">
          <div className="flex flex-col gap-1">
            <Text variant="h3" role="heading" level={3}>
              Name your app template
            </Text>
            <Text variant="body" theme="neutral">
              Nuon creates the app template and stubs out the config files it needs.
            </Text>
          </div>
          <div className="max-w-sm">
            <Input
              id="first-run-app-name"
              size="lg"
              placeholder="my-app"
              value={appName}
              onChange={(e) => onAppName(e.currentTarget.value)}
              labelProps={{ labelText: 'App template name' }}
              error={!!nameError}
              errorMessage={nameError}
              helperText={nameError ? undefined : APP_NAME_RULE}
              autoComplete="off"
              spellCheck={false}
            />
          </div>
          {/* Asked here so the stubbed runner, sandbox and permissions match the cloud the install will use. */}
          <TestCloudPicker value={cloud} onChange={onCloud} error={showErrors && !cloud} />
        </Card>
        <ExampleEscapeHatch onExit={onExitToExample} />
      </div>
    )
  } else {
    body = (
      <>
        <Card>
          <Text variant="h2" role="heading" level={2}>
            Deploy your app to a customer&apos;s cloud
          </Text>
          <ol className="grid gap-3 sm:grid-cols-3">
            {OWN_APP_STEPS.map((step, index) => (
              <li key={step.title} className="flex items-center gap-2.5 rounded-md border px-3.5 py-3">
                <Icon variant={step.icon} size={18} theme="brand" />
                <Text variant="body" weight="strong" className="min-w-0 flex-1">
                  {step.title}
                </Text>
                <Badge size="sm" theme="brand">
                  {index + 1}
                </Badge>
              </li>
            ))}
          </ol>
          <div>
            <Button variant="primary" size="lg" onClick={onExpand}>
              Start with your app <Icon variant="CaretRightIcon" weight="bold" />
            </Button>
          </div>
        </Card>
        <Card className="!gap-4">
          <div className="flex flex-col gap-2">
            <div className="flex items-center gap-2">
              <Icon variant="TireIcon" size={20} theme="neutral" />
              <Text variant="h3" role="heading" level={3}>
                Or kick the tires with our example app first
              </Text>
            </div>
            <Text variant="body" theme="neutral">
              Pre-wired with Terraform, Helm, images, manifests. Deploy it to your cloud account just like
              your customers would deploy your app.
            </Text>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            {EXAMPLE_CLOUDS.map((option) => (
              <Button
                key={option}
                variant="secondary"
                size="md"
                disabled={!!examplePending}
                onClick={() => onDeployExample(option)}
              >
                {examplePending === option ? (
                  <Icon variant="Loading" size={18} />
                ) : (
                  <Icon variant={CLOUD_ICON[option]} size={18} />
                )}
                Deploy to {CLOUD_LABEL[option]}
              </Button>
            ))}
          </div>
          <ExampleAppDrawer />
        </Card>
      </>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <Text variant="h1" role="heading" level={1}>
        Create your first app template
      </Text>
      {error ? <Banner theme="error">{error}</Banner> : null}
      {body}
      {expanded ? (
        <NextButton
          label="Next"
          onClick={onNext}
          onBack={onBackToIntro}
          loading={nextPending}
          disabled={nameInvalid || !!nextBlockedReason}
          disabledReason={nameInvalid ? APP_NAME_RULE : nextBlockedReason}
        />
      ) : (
        <NextButton onBack={onBackToIntro} showNext={false} />
      )}
    </div>
  )
}

const readString = (value: unknown) => (typeof value === 'string' ? value : '')

export const StartStep = ({ sharedData, setSharedData, onAdvance }: IWizardStepComponentProps) => {
  const { orgId, journey, vcsConnectionId: callbackConnectionId, vcsError, choosePath, backToIntro } =
    useFirstRun()
  const { githubAppName } = useConfig()
  const { user } = useAuth()
  const scrollRef = useRef<HTMLDivElement>(null)

  const path = sharedData.path === 'own' ? 'own' : 'example'
  const expanded = typeof sharedData.expandOwn === 'boolean' ? sharedData.expandOwn : path === 'own'
  const appName = readString(sharedData.app_name)
  const testCloud = isCloud(sharedData.testCloud) ? sharedData.testCloud : undefined

  const [showErrors, setShowErrors] = useState(false)
  const [appNameError, setAppNameError] = useState<string>()
  const [repoError, setRepoError] = useState<string>()
  const [error, setError] = useState<string>()
  const [nextPending, setNextPending] = useState(false)
  const [examplePending, setExamplePending] = useState<TExampleCloud>()
  const [connectingGithub, setConnectingGithub] = useState(false)

  const { data: connections } = useQuery({
    queryKey: ['first-run-vcs-connections', orgId, callbackConnectionId],
    queryFn: () => getVCSConnections({ orgId }),
    enabled: expanded,
  })
  // The connection GitHub just returned with, else any the org already has.
  const connectionId = callbackConnectionId ?? connections?.[0]?.id
  const connection = connections?.find((c) => c.id === connectionId)

  const { data: repos, isLoading: reposLoading } = useQuery({
    queryKey: ['first-run-vcs-repos', orgId, connectionId],
    queryFn: () => getVCSConnectionRepos({ orgId, connectionId: connectionId! }),
    enabled: expanded && !!connectionId,
  })

  const github: TGithubTile = connectionId
    ? {
        status: 'connected',
        owner: connection?.github_account_name,
        repoCount: repos?.total_count ?? repos?.repositories?.length,
      }
    : { status: 'disconnected', error: vcsError }

  const connectHref = githubAppName
    ? githubAppInstallUrl({ githubAppName, orgId, onboarding: true })
    : undefined

  const setOwn = (values: Record<string, unknown>) => {
    Object.entries(values).forEach(([key, value]) => setSharedData(key, value))
  }

  const expand = () => {
    const cloud = testCloud ?? 'aws'
    setOwn({ path: 'own', expandOwn: true })
    choosePath('own', cloud)
    requestAnimationFrame(() => scrollRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' }))
  }

  const exitToExample = () => {
    const cloud = EXAMPLE_CLOUDS[0]
    setOwn({ path: 'example', expandOwn: false, cloud, region: defaultRegion(cloud) })
    choosePath('example', cloud)
  }

  const onCloud = (cloud: TCloud) => {
    setOwn({ testCloud: cloud, cloud, region: defaultRegion(cloud) })
    choosePath('own', cloud)
  }

  // Leaving for GitHub drops in-memory state, so the draft rides on the journey.
  const connectGithub = async () => {
    if (!connectHref) return
    setConnectingGithub(true)
    try {
      await journey.saveStep('start', {
        path: 'own',
        app_name: appName,
        cloud: testCloud ?? '',
      })
    } catch {
      // The draft is a convenience; connecting still works without it.
    }
    window.location.assign(connectHref)
  }

  const next = async () => {
    setAppNameError(undefined)
    setRepoError(undefined)
    setError(undefined)
    if (!appName || !connectionId || !testCloud) {
      setShowErrors(true)
      scrollRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' })
      return
    }
    if (!APP_NAME_PATTERN.test(appName) || reposLoading) return
    const repo = pickConfigRepo(repos?.repositories, appName)
    if (!repo) {
      setRepoError(`No repo named ${appName} in your GitHub connection.`)
      return
    }

    setNextPending(true)
    const base = { path: 'own', cloud: testCloud, app_name: appName, repo: repo.full_name }
    try {
      const { appId, branchId } = await setUpOwnApp({
        orgId,
        name: appName,
        repo: repo.full_name,
        vcsConnectionId: connectionId,
        saved: {
          appId: readString(sharedData.app_id) || undefined,
          branchId: readString(sharedData.app_branch_id) || undefined,
        },
        onProgress: ({ appId: id, branchId: branch }) =>
          journey.saveStep('start', { ...base, app_id: id, app_branch_id: branch ?? '' }),
      })
      await journey.saveStep(
        'start',
        { ...base, app_id: appId, app_branch_id: branchId },
        { complete: true }
      )
      // Connect may be marked done from an earlier pass on the example path.
      await journey.saveStep('connect', {}, { complete: false })
      trackEvent({ event: 'app_create', status: 'ok', user, props: { appId, path: 'own', cloud: testCloud } })
      setOwn({ ...base, app_id: appId, app_branch_id: branchId, region: defaultRegion(testCloud) })
      choosePath('own', testCloud)
      onAdvance()
    } catch (err) {
      trackEvent({
        event: 'app_create',
        status: 'error',
        user,
        props: {
          path: 'own',
          cloud: testCloud,
          err: err instanceof AppNameTakenError ? 'name_taken' : (err as TAPIError)?.error,
        },
      })
      if (err instanceof AppNameTakenError) {
        setAppNameError('An app template with this name exists')
      } else {
        setError((err as TAPIError)?.description || (err as Error)?.message || 'Unable to create the app template.')
      }
    } finally {
      setNextPending(false)
    }
  }

  const deployExample = async (cloud: TExampleCloud) => {
    setError(undefined)
    setExamplePending(cloud)
    choosePath('example', cloud)
    try {
      const { appId, appName: exampleName, branchId } = await setUpKitchenSink({ orgId, cloud })
      const values = {
        path: 'example',
        cloud,
        app_name: exampleName,
        app_id: appId,
        app_branch_id: branchId,
        repo: KITCHEN_SINK_REPO,
      }
      await journey.saveStep('start', values, { complete: true })
      // The example path has no Connect step, so it is done by definition.
      await journey.saveStep('connect', {}, { complete: true })
      trackEvent({ event: 'app_create', status: 'ok', user, props: { appId, path: 'example', cloud } })
      setOwn({ ...values, expandOwn: false, region: defaultRegion(cloud) })
      onAdvance()
    } catch (err) {
      trackEvent({
        event: 'app_create',
        status: 'error',
        user,
        props: { path: 'example', cloud, err: (err as TAPIError)?.error },
      })
      setError((err as TAPIError)?.description || 'Unable to set up the example app.')
    } finally {
      setExamplePending(undefined)
    }
  }

  return (
    <div ref={scrollRef}>
      <StartStepView
        expanded={expanded}
        onExpand={expand}
        onBackToIntro={backToIntro}
        onExitToExample={exitToExample}
        appName={appName}
        onAppName={(name) => {
          setAppNameError(undefined)
          setRepoError(undefined)
          setSharedData('app_name', name)
        }}
        appNameError={appNameError}
        repoError={repoError}
        github={github}
        connectHref={connectHref}
        onConnectGithub={connectGithub}
        connectingGithub={connectingGithub}
        cloud={testCloud}
        onCloud={onCloud}
        showErrors={showErrors}
        onNext={next}
        nextPending={nextPending}
        nextBlockedReason={reposLoading ? 'Loading your GitHub repos' : undefined}
        onDeployExample={deployExample}
        examplePending={examplePending}
        error={error}
      />
    </div>
  )
}
