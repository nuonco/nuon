import { useEffect, useState, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Code } from '@/components/common/Code'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { useAuth } from '@/hooks/use-auth'
import { useFirstRun } from '@/hooks/use-first-run'
import { trackEvent } from '@/lib/posthog-analytics'
import { openSupportChat } from '@/lib/pylon-chat'
import type { IWizardStepComponentProps } from '@/providers/onboarding-wizard-provider'
import { cn } from '@/utils/classnames'
import { findPushRun, getBranchRuns, pushRunSha } from './api'
import {
  CLOUD_LABEL,
  CONTACT_MESSAGE,
  DOCS_CONFIG_FILES,
  DOCS_LSP,
  DOCS_MCP,
  DOCS_RUNNERS,
  DOCS_SANDBOXES,
  EXAMPLE_CLOUDS,
  GIT_PUSH,
  KITCHEN_SINK_URL,
  LSP_NEOVIM_SETUP,
  MCP_ADD_CLAUDE,
  PROMPT_URL,
  VSCODE_EXTENSION,
  appFileStubs,
  defaultRegion,
  isCloud,
  type TCloud,
} from './constants'
import { CopyTextButton, InlineLink, NextButton } from './shared'

const PUSH_POLL_MS = 5000

export const fetchAgentPrompt = async (): Promise<string> => {
  const res = await fetch(PROMPT_URL, { credentials: 'omit' })
  const type = res.headers.get('content-type') ?? ''
  if (!res.ok || type.includes('text/html')) throw new Error(`prompt returned ${res.status}`)
  return res.text()
}

const PushListener = ({
  detected,
  sha,
  cloud,
}: {
  detected: boolean
  sha?: string
  cloud: TCloud
}) => (
  <div
    className={cn(
      'flex items-start justify-between gap-3 rounded-md bg-background p-4 ring-1 transition-shadow',
      detected ? 'ring-green-500 dark:ring-green-400' : 'ring-neutral-200 dark:ring-neutral-700'
    )}
  >
    <div className="flex min-w-0 flex-1 items-start gap-3">
      {detected ? (
        <Icon variant="CheckCircleIcon" size={20} weight="fill" theme="success" />
      ) : (
        <Icon variant="Loading" size={20} />
      )}
      <div className="flex min-w-0 flex-col gap-0.5">
        <Text variant="body" weight="strong">
          {detected ? 'Synced from main' : 'The prompt ends with pushing your app config'}
        </Text>
        {detected ? (
          <Text variant="subtext" theme="neutral" flex className="flex-wrap">
            Commit
            {sha ? (
              <Code variant="inline">{sha}</Code>
            ) : null}
            updated the default app branch. Building your components now.
          </Text>
        ) : (
          <Text variant="subtext" theme="neutral">
            Continue now and the next steps create the install and provision the{' '}
            <InlineLink href={DOCS_RUNNERS}>runner</InlineLink> and{' '}
            <InlineLink href={DOCS_SANDBOXES}>Nuon sandbox</InlineLink> in your {CLOUD_LABEL[cloud]}{' '}
            test account. Your app deploys when the push lands. Or wait for the push and watch it all
            deploy in one workflow.
          </Text>
        )}
      </div>
    </div>
    <Badge size="sm" theme={detected ? 'success' : 'brand'} className="mt-0.5 shrink-0">
      {detected ? 'Synced' : 'Watching'}
    </Badge>
  </div>
)

const FileStubRows = ({ appName, cloud, repo }: { appName: string; cloud: TCloud; repo: string }) => {
  const [open, setOpen] = useState<string[]>([])
  const toggle = (name: string) =>
    setOpen((prev) => (prev.includes(name) ? prev.filter((n) => n !== name) : [...prev, name]))

  return (
    <ul className="flex flex-col rounded-md border divide-y">
      {appFileStubs({ appName, cloud, repo }).map((file) => {
        const isOpen = open.includes(file.name)
        return (
          <li key={file.name} className="flex flex-col">
            <button
              type="button"
              aria-expanded={isOpen}
              onClick={() => toggle(file.name)}
              className="flex w-full items-center justify-between gap-3 px-4 py-3 text-left cursor-pointer hover:bg-cool-grey-500/8"
            >
              <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
                <Text as="span" variant="body" family="mono" weight="strong">
                  {file.name}
                </Text>
                <Badge size="sm" theme={file.required ? 'brand' : 'neutral'}>
                  {file.badge}
                </Badge>
                <Text as="span" variant="subtext" theme="neutral">
                  {file.purpose}
                </Text>
              </div>
              <span className={cn('shrink-0 transition-transform', isOpen && 'rotate-90')} aria-hidden>
                <Icon variant="CaretRightIcon" size={16} theme="neutral" />
              </span>
            </button>
            {isOpen ? (
              <div className="border-t p-3">
                <CodeBlock language="toml" wrapLongLines>
                  {file.snippet}
                </CodeBlock>
              </div>
            ) : null}
          </li>
        )
      })}
    </ul>
  )
}

const ManualSetup = ({ appName, repo, cloud }: { appName: string; repo: string; cloud: TCloud }) => {
  const steps: { title: string; body: ReactNode; detail?: ReactNode }[] = [
    {
      title: 'Put the config at the root of the repo',
      body: (
        <>
          Top level of{' '}
          <Code variant="inline">{repo}</Code>
          , the same layout as <InlineLink href={KITCHEN_SINK_URL}>nuonco/kitchen-sink</InlineLink>
        </>
      ),
    },
    {
      title: 'Fill in the app config templates with your values',
      body: (
        <>
          Point each component at a repo and branch, pick a sandbox, scope roles.{' '}
          <InlineLink href={DOCS_CONFIG_FILES}>Configuration files</InlineLink>
        </>
      ),
      detail: <FileStubRows appName={appName} cloud={cloud} repo={repo} />,
    },
    {
      title: 'Push to main',
      body: <>Every push syncs the default app branch.</>,
      detail: (
        <CodeBlock language="bash" showCopy>
          {GIT_PUSH}
        </CodeBlock>
      ),
    },
  ]

  return (
    <div className="flex flex-col gap-4">
      <ol className="flex flex-col gap-4">
        {steps.map((step, index) => (
          <li key={step.title} className="flex gap-3">
            <Badge size="sm" theme="brand" className="mt-0.5 shrink-0">
              {index + 1}
            </Badge>
            <div className="flex min-w-0 flex-1 flex-col gap-1.5">
              <Text variant="body" weight="strong">
                {step.title}
              </Text>
              <Text variant="subtext" theme="neutral">
                {step.body}
              </Text>
              {step.detail}
            </div>
          </li>
        ))}
      </ol>
      <Text variant="subtext" theme="neutral" flex className="flex-wrap border-t pt-3">
        Editing TOML by hand? The Nuon language server adds autocomplete and validation:
        <Link href={VSCODE_EXTENSION} isExternal textVariant="subtext">
          VS Code extension
        </Link>
        <span aria-hidden>·</span>
        <Link href={LSP_NEOVIM_SETUP} isExternal textVariant="subtext">
          Neovim setup
        </Link>
        <span aria-hidden>·</span>
        <Link href={DOCS_LSP} isExternal textVariant="subtext">
          Language server docs
        </Link>
      </Text>
    </div>
  )
}

type TFootnote = 'mcp' | 'deps' | 'manual'
const FOOTNOTE_LINK =
  'cursor-pointer text-cool-grey-500 underline decoration-dotted underline-offset-2 hover:text-foreground dark:text-cool-grey-400'

const Footnotes = ({
  appName,
  repo,
  cloud,
  onExampleExit,
}: {
  appName: string
  repo: string
  cloud: TCloud
  onExampleExit: () => void
}) => {
  const [open, setOpen] = useState<TFootnote | null>(null)
  const toggle = (key: TFootnote) => setOpen((prev) => (prev === key ? null : key))
  const panel = (key: TFootnote, label: string) => (
    <button
      type="button"
      aria-expanded={open === key}
      aria-controls={`footnote-${key}`}
      onClick={() => toggle(key)}
      className={FOOTNOTE_LINK}
    >
      {label}
    </button>
  )

  return (
    <div className="flex flex-col gap-3">
      <Text variant="subtext" theme="neutral" flex className="flex-wrap gap-x-2">
        <span>Optional:</span>
        {panel('mcp', 'MCP setup')}
        <span aria-hidden>·</span>
        {panel('deps', 'Dependencies')}
        <span aria-hidden>·</span>
        {panel('manual', 'Manual steps')}
        <span aria-hidden>·</span>
        <button type="button" onClick={onExampleExit} className={FOOTNOTE_LINK}>
          Use the example app instead
        </button>
      </Text>
      {open === 'manual' ? (
        <div id="footnote-manual" className="rounded-md border bg-background p-4">
          <ManualSetup appName={appName} repo={repo} cloud={cloud} />
        </div>
      ) : null}
      {open === 'mcp' ? (
        <div id="footnote-mcp" className="flex flex-col gap-1.5 rounded-md border bg-background p-4">
          <Text variant="subtext" weight="strong">
            Give your agent the Nuon Model Context Protocol (MCP) server
          </Text>
          <Text variant="subtext" theme="neutral">
            Live access to your org while it works: apps, builds, installs and logs. In Claude Code:
          </Text>
          <CodeBlock language="bash" showCopy wrapLongLines className="!pr-14">
            {MCP_ADD_CLAUDE}
          </CodeBlock>
          <Link href={DOCS_MCP} isExternal textVariant="subtext">
            docs.nuon.co/guides/agents/setup
          </Link>
        </div>
      ) : null}
      {open === 'deps' ? (
        <div id="footnote-deps" className="flex flex-col gap-1.5 rounded-md border bg-background p-4">
          <Text variant="subtext" weight="strong">
            Dependencies
          </Text>
          <Text variant="subtext" theme="neutral">
            Databases like Postgres run as components in the customer&apos;s account. Third-party services
            like Clerk or SendGrid stay external; their keys arrive as install inputs.
          </Text>
        </div>
      ) : null}
    </div>
  )
}

const AgentSetup = ({
  prompt,
  promptFailed,
  onCopyPrompt,
}: {
  prompt?: string
  promptFailed: boolean
  onCopyPrompt?: () => void
}) => {
  const [head, ...rest] = (prompt ?? '').split(' ')
  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-1.5">
        <Text as="h3" variant="h3" weight="strong" flex>
          <Icon variant="RobotIcon" size={20} />
          Have your agent write the config
        </Text>
        <Text variant="body" theme="neutral">
          Paste this prompt in the same directory as your app. Then, you&apos;re one step from a test install
          as if it were a customer&apos;s cloud.
        </Text>
      </div>
      <div className="flex flex-col gap-4 rounded-md border bg-background p-4 sm:flex-row sm:items-center">
        <div className="line-clamp-1 min-w-0 flex-1">
          {prompt ? (
            <>
              <Text as="span" variant="body" family="mono" weight="strong" theme="brand">
                {head}
              </Text>
              <Text as="span" variant="body" family="mono" theme="neutral">
                {rest.length ? ` ${rest.join(' ')}` : ''}
              </Text>
            </>
          ) : (
            <Text as="span" variant="body" family="mono" theme="neutral">
              {promptFailed ? 'The prompt did not load. Open the full prompt instead.' : 'Loading the prompt...'}
            </Text>
          )}
        </div>
        <div className="flex shrink-0 flex-wrap items-center gap-2">
          <CopyTextButton
            text={prompt ?? ''}
            label="Copy prompt"
            size="lg"
            variant="primary"
            disabled={!prompt}
            disabledReason={promptFailed ? 'Cannot copy the prompt: it did not load' : 'Loading the prompt'}
            onCopy={onCopyPrompt}
          />
          <Button variant="secondary" size="lg" href={PROMPT_URL} target="_blank" rel="noreferrer">
            See full prompt <Icon variant="ArrowSquareOutIcon" size={14} />
          </Button>
        </div>
      </div>
    </div>
  )
}

export interface IConnectStepView {
  appName: string
  repo: string
  cloud: TCloud
  prompt?: string
  promptFailed: boolean
  detected: boolean
  sha?: string
  confirmSkip: boolean
  onContinue: () => void
  onContinueAnyway: () => void
  onKeepWaiting: () => void
  onExampleExit: () => void
  onGetHelp: () => void
  onCopyPrompt?: () => void
  onBack?: () => void
}

export const ConnectStepView = ({
  appName,
  repo,
  cloud,
  prompt,
  promptFailed,
  detected,
  sha,
  confirmSkip,
  onContinue,
  onContinueAnyway,
  onKeepWaiting,
  onExampleExit,
  onGetHelp,
  onCopyPrompt,
  onBack,
}: IConnectStepView) => (
  <div className="flex flex-col gap-6">
    <Card className="!gap-10 !p-5 !border-0 !shadow-none bg-primary-50 dark:bg-primary-950/40 ring-1 ring-primary-200 dark:ring-primary-800">
      <AgentSetup prompt={prompt} promptFailed={promptFailed} onCopyPrompt={onCopyPrompt} />
      <PushListener detected={detected} sha={sha} cloud={cloud} />
    </Card>
    <Footnotes appName={appName} repo={repo} cloud={cloud} onExampleExit={onExampleExit} />
    {confirmSkip && !detected ? (
      <Banner theme="warn">
        <div className="flex flex-col gap-2">
          <Text weight="strong">Nuon does not have your app config yet</Text>
          <Text variant="subtext">
            If you continue now, you would only deploy Nuon infrastructure that your BYOC install runs on:{' '}
            <InlineLink href={DOCS_RUNNERS}>Nuon Runner</InlineLink> and{' '}
            <InlineLink href={DOCS_SANDBOXES}>Nuon Sandbox</InlineLink>. This means you&apos;ll do a separate
            deploy of your app afterward.
          </Text>
          <div className="flex flex-wrap items-center gap-3 pt-1">
            <Button variant="secondary" size="sm" onClick={onContinueAnyway}>
              Continue anyway
            </Button>
            <Button variant="ghost" size="sm" onClick={onKeepWaiting}>
              Keep waiting
            </Button>
          </div>
        </div>
      </Banner>
    ) : null}
    <NextButton
      label="Set up your first install"
      onClick={onContinue}
      onBack={onBack}
      size="lg"
      secondary={
        <Button variant="secondary" size="lg" onClick={onGetHelp}>
          <Icon variant="ChatCircleIcon" size={16} /> Get help
        </Button>
      }
    />
  </div>
)

const readString = (value: unknown) => (typeof value === 'string' ? value : '')

export const ConnectStep = ({ sharedData, setSharedData, onAdvance, onGoBack }: IWizardStepComponentProps) => {
  const { orgId, journey, choosePath } = useFirstRun()
  const { user } = useAuth()
  const appName = readString(sharedData.app_name)
  const repo = readString(sharedData.repo)
  const appId = readString(sharedData.app_id)
  const branchId = readString(sharedData.app_branch_id)
  const cloud: TCloud = isCloud(sharedData.cloud) ? sharedData.cloud : 'aws'
  const [confirmSkip, setConfirmSkip] = useState(false)
  const [pushSeen, setPushSeen] = useState(false)

  const { data: prompt, isError: promptFailed } = useQuery({
    queryKey: ['first-run-agent-prompt'],
    queryFn: fetchAgentPrompt,
    staleTime: Infinity,
    retry: 1,
  })

  const { data: runs } = useQuery({
    queryKey: ['first-run-branch-runs', orgId, appId, branchId],
    queryFn: () => getBranchRuns({ orgId, appId, branchId }),
    enabled: !!appId && !!branchId,
    refetchInterval: pushSeen ? false : PUSH_POLL_MS,
  })
  const pushRun = findPushRun(runs)
  const detected = !!pushRun

  useEffect(() => {
    if (detected) setPushSeen(true)
  }, [detected])

  const advance = () => {
    journey.saveStep('connect', {}, { complete: true }).catch(() => {})
    onAdvance()
  }

  const exitToExample = () => {
    const example = EXAMPLE_CLOUDS[0]
    setSharedData('expandOwn', false)
    setSharedData('path', 'example')
    setSharedData('cloud', example)
    setSharedData('region', defaultRegion(example))
    choosePath('example', example)
    onGoBack?.()
  }

  return (
    <ConnectStepView
      appName={appName}
      repo={repo}
      cloud={cloud}
      prompt={prompt}
      promptFailed={promptFailed}
      detected={detected}
      sha={pushRunSha(pushRun)}
      confirmSkip={confirmSkip}
      onContinue={() => (detected ? advance() : setConfirmSkip(true))}
      onContinueAnyway={advance}
      onKeepWaiting={() => setConfirmSkip(false)}
      onExampleExit={exitToExample}
      onGetHelp={() => {
        openSupportChat(CONTACT_MESSAGE)
        trackEvent({ event: 'support_chat_open', status: 'ok', user, props: { source: 'onboarding_connect' } })
      }}
      onCopyPrompt={() =>
        trackEvent({ event: 'agent_prompt_copy', status: 'ok', user, props: { appId } })
      }
      onBack={onGoBack}
    />
  )
}
