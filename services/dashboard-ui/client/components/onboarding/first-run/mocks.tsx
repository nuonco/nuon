import { createContext, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { OnboardingWizardLayout } from '@/components/onboarding/OnboardingWizard/OnboardingWizard'
import {
  OnboardingWizardProvider,
  WizardContext,
  type IWizardStepComponentProps,
} from '@/providers/onboarding-wizard-provider'
import { CopyEditor } from '../CopyEditor'
import { ConnectStepView } from './ConnectStep'
import {
  CLOUD_CONNECT,
  KITCHEN_SINK_LABEL,
  KITCHEN_SINK_REPO,
  STACK_RENDER_LABEL,
  defaultRegion,
  isCloud,
  type TCloud,
  type TExampleCloud,
  type TPath,
} from './constants'
import { DeployStepView, type TDeployPhase } from './DeployStep'
import { IntroScreen } from './IntroScreen'
import type { IWaitItem, TWaitState } from './progress'
import { ProvisionStepView } from './ProvisionStep'
import { StackStepView, type TStackPhase } from './StackStep'
import { StartStepView, type TGithubTile } from './StartStep'
import { buildFirstRunSteps } from './steps'

// Story-only: runs the real step views with mocked data so every button in
// Ladle does what the product does, or says what it would do.

const TICK_MS = 1500
const OWN_APP = 'acme-api'
const OWN_REPO = 'acme/acme-api'
const PROMPT =
  '/goal Set up this application on Nuon (nuon.co) so a customer can run it in their own cloud account. This is sample text standing in for the file served at nuon.co/loop.md.'

export type TMockStep = 'intro' | 'start' | 'connect' | 'deploy' | 'stack' | 'provision' | 'done'

// Seeds for the first screen the story opens on. Screens reached by clicking
// start from their normal state.
export interface IMockScenario {
  github?: TGithubTile
  appName?: string
  testCloud?: TCloud
  showErrors?: boolean
  appNameError?: string
  repoError?: string
  creating?: TExampleCloud
  expanded?: boolean
  confirmSkip?: boolean
  detected?: boolean
  promptFailed?: boolean
  deploy?: 'idle' | 'waiting' | 'slow' | 'sync-failed' | 'missing-inputs'
  stack?: 'generating' | 'ready' | 'launched' | 'generation-failed' | 'provision-failed'
}

export interface IMockFlowStart {
  step: TMockStep
  path?: TPath
  cloud?: TCloud
  scenario?: IMockScenario
}

type TMock = { path: TPath; cloud: TCloud }

// What the user already did, so Back then Next returns to the same state
// instead of replaying it. Progress past Start is kept per path and cloud.
type TStartMemory = { appName?: string; testCloud?: TCloud; github?: TGithubTile; expanded?: boolean }
type TRunMemory = { detected?: boolean; installCreated?: boolean; stackDone?: boolean }
type TFlowMemory = { start: TStartMemory; runs: Record<string, TRunMemory> }
type TMemory = {
  start: () => TStartMemory
  run: (mock: TMock) => TRunMemory
  setStart: (patch: TStartMemory) => void
  setRun: (mock: TMock, patch: TRunMemory) => void
}
const emptyMemory = (): TFlowMemory => ({ start: {}, runs: {} })
const runKeyFor = ({ path, cloud }: TMock) => `${path}.${cloud}`

const appNameFor = ({ path }: TMock) => (path === 'own' ? OWN_APP : 'kitchen-sink')
const repoFor = ({ path }: TMock) => (path === 'own' ? OWN_REPO : KITCHEN_SINK_REPO)
const labelFor = ({ path }: TMock) => (path === 'own' ? OWN_APP : KITCHEN_SINK_LABEL)

const waitStates = (labels: string[], done: number, failedAt?: number): IWaitItem[] =>
  labels.map((label, i) => ({
    label,
    state: (i === failedAt ? 'error' : i < done ? 'done' : i === done ? 'active' : 'pending') as TWaitState,
  }))

const useTicker = (total: number, running: boolean, start = 0) => {
  const [count, setCount] = useState(start)
  useEffect(() => {
    if (!running) return
    const timer = setInterval(() => setCount((c) => Math.min(c + 1, total)), TICK_MS)
    return () => clearInterval(timer)
  }, [running, total])
  return [count, setCount] as const
}

const NoticeContext = createContext<(text: string) => void>(() => {})

const NoticeHost = ({ children }: { children: ReactNode }) => {
  const [notice, setNotice] = useState<string>()
  useEffect(() => {
    if (!notice) return
    const timer = setTimeout(() => setNotice(undefined), 3500)
    return () => clearTimeout(timer)
  }, [notice])
  return (
    <NoticeContext.Provider value={setNotice}>
      {children}
      {notice ? (
        <div className="fixed bottom-24 left-1/2 z-50 -translate-x-1/2 rounded-md border bg-background px-4 py-2.5 shadow-lg">
          <Text variant="subtext">
            <strong>In the product:</strong> {notice}
          </Text>
        </div>
      ) : null}
    </NoticeContext.Provider>
  )
}

const useNotice = () => useContext(NoticeContext)

type TNav = {
  choose: (path: TPath, cloud: TCloud, step?: TMockStep) => void
  toIntro: () => void
  finish: () => void
}

const MockStart = (nav: TNav, seed: IMockScenario | undefined, mock: TMock, memory: TMemory) =>
  function MockStartStep() {
    const notice = useNotice()
    const kept = seed ? {} : memory.start()
    // Reached by clicking, Start opens on the path choice like the product, or
    // on the form you left; a story opened on the own-app form starts expanded.
    const [expanded, setExpandedState] = useState(
      seed ? (seed.expanded ?? mock.path === 'own') : (kept.expanded ?? false)
    )
    const [appName, setAppNameState] = useState(seed?.appName ?? kept.appName ?? '')
    const [cloud, setCloudState] = useState<TCloud | undefined>(seed?.testCloud ?? kept.testCloud)
    const [github, setGithubState] = useState<TGithubTile>(
      seed?.github ?? kept.github ?? { status: 'disconnected' }
    )
    const setExpanded = (value: boolean) => {
      setExpandedState(value)
      memory.setStart({ expanded: value })
    }
    const setAppName = (value: string) => {
      setAppNameState(value)
      memory.setStart({ appName: value })
    }
    const setCloud = (value: TCloud) => {
      setCloudState(value)
      memory.setStart({ testCloud: value })
    }
    const setGithub = (value: TGithubTile) => {
      setGithubState(value)
      memory.setStart({ github: value })
    }
    const [connecting, setConnecting] = useState(false)
    const [showErrors, setShowErrors] = useState(seed?.showErrors ?? false)
    const [appNameError, setAppNameError] = useState(seed?.appNameError)
    const [repoError, setRepoError] = useState(seed?.repoError)
    const [nextPending, setNextPending] = useState(false)
    const [pending, setPending] = useState<TExampleCloud | undefined>(seed?.creating)

    useEffect(() => {
      if (!seed?.creating) return
      const exampleCloud = seed.creating
      const timer = setTimeout(() => nav.choose('example', exampleCloud), 2000)
      return () => clearTimeout(timer)
    }, [])

    return (
      <StartStepView
        expanded={expanded}
        onExpand={() => setExpanded(true)}
        onBackToIntro={nav.toIntro}
        onExitToExample={() => setExpanded(false)}
        appName={appName}
        onAppName={(name) => {
          setAppName(name)
          setAppNameError(undefined)
          setRepoError(undefined)
        }}
        appNameError={appNameError}
        repoError={repoError}
        github={github}
        connectHref="https://github.com/apps/nuon/installations/new"
        connectingGithub={connecting}
        onConnectGithub={() => {
          setConnecting(true)
          notice('opens GitHub to install the Nuon app, then returns here connected.')
          setTimeout(() => {
            setConnecting(false)
            setGithub({ status: 'connected', owner: 'acme', repoCount: 3 })
          }, 1500)
        }}
        cloud={cloud}
        onCloud={(next) => {
          setCloud(next)
          setRepoError(undefined)
        }}
        showErrors={showErrors}
        nextPending={nextPending}
        onNext={(values) => {
          setShowErrors(true)
          if (appNameError || repoError) return
          setNextPending(true)
          setTimeout(() => nav.choose('own', isCloud(values.cloud) ? values.cloud : 'aws', 'connect'), 900)
        }}
        examplePending={pending}
        onDeployExample={(exampleCloud) => {
          setPending(exampleCloud)
          setTimeout(() => nav.choose('example', exampleCloud), 1200)
        }}
      />
    )
  }

const MockConnect = (nav: TNav, seed: IMockScenario | undefined, mock: TMock, memory: TMemory) =>
  function MockConnectStep({ onAdvance, onGoBack }: IWizardStepComponentProps) {
    const notice = useNotice()
    const [detected, setDetected] = useState(seed?.detected ?? (!seed && !!memory.run(mock).detected))
    const [confirmSkip, setConfirmSkip] = useState(seed?.confirmSkip ?? false)
    useEffect(() => {
      if (detected) return
      const timer = setTimeout(() => {
        setDetected(true)
        memory.setRun(mock, { detected: true })
      }, 10_000)
      return () => clearTimeout(timer)
    }, [detected])
    return (
      <ConnectStepView
        appName={OWN_APP}
        repo={OWN_REPO}
        cloud={mock.cloud}
        prompt={seed?.promptFailed ? undefined : PROMPT}
        promptFailed={seed?.promptFailed ?? false}
        detected={detected}
        sha={detected ? 'a1b2c3d' : undefined}
        confirmSkip={confirmSkip}
        onContinue={() => (detected ? onAdvance() : setConfirmSkip(true))}
        onContinueAnyway={onAdvance}
        onKeepWaiting={() => setConfirmSkip(false)}
        onExampleExit={() => nav.choose('example', 'aws', 'start')}
        onGetHelp={() => notice('opens support chat with a message about the app config.')}
        onCopyPrompt={() => notice('the prompt is copied for your coding agent.')}
        onBack={onGoBack}
      />
    )
  }

const MockDeploy = (seed: IMockScenario | undefined, mock: TMock, memory: TMemory) =>
  function MockDeployStep({ onAdvance, onGoBack, sharedData, setSharedData }: IWizardStepComponentProps) {
    const notice = useNotice()
    const scenario = seed?.deploy ?? 'idle'
    const installCreated =
      sharedData.install_id === 'created' || (!seed && !!memory.run(mock).installCreated)
    const [phase, setPhase] = useState<TDeployPhase>(
      scenario === 'idle' || scenario === 'missing-inputs' ? 'idle' : 'waiting-config'
    )
    const [missing, setMissing] = useState<string[]>([])
    const [since, setSince] = useState<number | undefined>(
      scenario === 'slow' ? Date.now() - 11 * 60_000 : scenario === 'idle' || scenario === 'missing-inputs' ? undefined : Date.now() - 40_000
    )
    const [failed, setFailed] = useState(scenario === 'sync-failed')
    const [retrying, setRetrying] = useState(false)
    const labels = [
      'Start a branch run on main',
      `Fetch the latest commit from ${repoFor(mock)}`,
      'Read and validate the app config',
      'Build the components and sandbox',
      'Create the install',
    ]
    const buildsIndex = 3
    const stalled = scenario === 'slow'
    const [done, setDone] = useTicker(
      stalled ? buildsIndex : labels.length,
      phase === 'waiting-config' && !failed,
      scenario === 'idle' || scenario === 'missing-inputs' ? 0 : buildsIndex
    )
    const advanced = useRef(false)
    useEffect(() => {
      if (phase !== 'waiting-config' || done < labels.length || advanced.current) return
      advanced.current = true
      setSharedData('install_id', 'created')
      memory.setRun(mock, { installCreated: true })
      onAdvance()
    }, [phase, done, labels.length, onAdvance, setSharedData])

    return (
      <DeployStepView
        path={mock.path}
        appName={appNameFor(mock)}
        repo={repoFor(mock)}
        cloud={mock.cloud}
        region={defaultRegion(mock.cloud)}
        phase={phase}
        installCreated={installCreated}
        missingInputs={missing}
        syncError={failed ? 'Component builds failed: one or more component builds failed.' : undefined}
        progress={waitStates(labels, done, failed ? buildsIndex : undefined)}
        progressSince={since}
        retrying={retrying}
        onRetry={() => {
          setRetrying(true)
          setTimeout(() => {
            setRetrying(false)
            setFailed(false)
            setDone(buildsIndex)
          }, 900)
        }}
        onGetHelp={() => notice('opens support chat about the stuck app config.')}
        onCreate={() => {
          if (installCreated) {
            onAdvance()
            return
          }
          if (scenario === 'missing-inputs') {
            setMissing(['db_password', 'api_domain'])
            return
          }
          setSince(Date.now())
          setPhase('waiting-config')
        }}
        onBack={onGoBack}
      />
    )
  }

const MockStack = (seed: IMockScenario | undefined, mock: TMock, memory: TMemory) =>
  function MockStackStep({ onAdvance, onGoBack }: IWizardStepComponentProps) {
    const notice = useNotice()
    const stackDone = !seed && !!memory.run(mock).stackDone
    const scenario = seed?.stack ?? (stackDone ? 'ready' : 'generating')
    // Stories opened mid-wait show time already spent; reached by clicking, the clock starts now.
    const [since] = useState(() =>
      seed ? Date.now() - (scenario === 'generating' ? 8_000 : 50_000) : Date.now()
    )
    const [launched, setLaunched] = useState(scenario === 'launched')
    const advance = () => {
      memory.setRun(mock, { stackDone: true })
      onAdvance()
    }
    const labels = [
      'Create your test install',
      STACK_RENDER_LABEL[mock.cloud],
      `${CLOUD_CONNECT[mock.cloud].artifactNoun} ready`,
    ]
    const ready = scenario !== 'generating'
    const [done] = useTicker(labels.length, scenario === 'generating', ready ? labels.length : 1)
    const failed =
      scenario === 'generation-failed'
        ? 'Stack template rendering failed: missing input vpc_id.'
        : scenario === 'provision-failed'
          ? 'error while generating steps'
          : undefined
    const phase: TStackPhase = failed
      ? 'error'
      : stackDone
        ? 'done'
        : done < labels.length
          ? 'generating'
          : launched
            ? 'launched'
            : 'ready'

    useEffect(() => {
      if (phase !== 'launched') return
      const timer = setTimeout(advance, 8000)
      return () => clearTimeout(timer)
    }, [phase])

    return (
      <StackStepView
        cloud={mock.cloud}
        appName={labelFor(mock)}
        region={defaultRegion(mock.cloud)}
        phase={phase}
        quickLinkUrl={mock.cloud === 'aws' ? 'https://console.aws.amazon.com/cloudformation/home' : undefined}
        errorDescription={failed}
        progress={waitStates(labels, done)}
        progressSince={since}
        onGetHelp={() => notice('opens support chat about the install stack.')}
        onLaunch={() => {
          setLaunched(true)
          notice(
            mock.cloud === 'aws'
              ? 'the CloudFormation console opens in a new tab. This mock moves on about 8 seconds after launch.'
              : 'you apply the stack from your terminal. This mock moves on about 8 seconds after launch.'
          )
        }}
        onContinue={advance}
        onBack={onGoBack}
      />
    )
  }

const MockProvision = (nav: TNav, mock: TMock) =>
  function MockProvisionStep({ onGoBack }: IWizardStepComponentProps) {
    return (
      <ProvisionStepView
        cloud={mock.cloud}
        appName={labelFor(mock)}
        region={defaultRegion(mock.cloud)}
        installId="inlk3x9q2m7v4w8p1z6r5t0y2c"
        onFinish={nav.finish}
        onBack={onGoBack}
      />
    )
  }

const FlowDone = ({ onRestart }: { onRestart: () => void }) => (
  <div className="flex min-h-screen items-center justify-center bg-background p-6">
    <Card className="!gap-4 max-w-md">
      <div className="flex items-center gap-2">
        <Icon variant="CheckCircleIcon" size={20} theme="success" weight="fill" />
        <Text variant="base" weight="strong">
          End of onboarding
        </Text>
      </div>
      <Text variant="body" theme="neutral">
        In the product, this lands on the install&apos;s deploy workflow in the dashboard (or the dashboard home
        if you skipped).
      </Text>
      <div>
        <Button variant="primary" onClick={onRestart}>
          Start over
        </Button>
      </div>
    </Card>
  </div>
)

type TFlowState = { step: TMockStep; path: TPath; cloud: TCloud; seeded: boolean }

const FLOW_KEY = 'first-run-clickable-flow'

type TStored = { state: TFlowState; memory: TFlowMemory }

// Ladle rewrites its own URL and remounts the story, so a persisted flow keeps
// its place and what you did in session storage.
const readStored = (): TStored | undefined => {
  try {
    const parsed = JSON.parse(sessionStorage.getItem(FLOW_KEY) ?? 'null')
    const { step, path, cloud } = parsed?.state ?? {}
    if ((path === 'own' || path === 'example') && isCloud(cloud) && step) {
      return {
        state: { step, path, cloud, seeded: false },
        memory: { ...emptyMemory(), ...parsed.memory },
      }
    }
  } catch {
    // Storage blocked or an older format: start fresh.
  }
  return undefined
}

const store = (stored: TStored) => {
  try {
    sessionStorage.setItem(FLOW_KEY, JSON.stringify(stored))
  } catch {
    // Storage blocked: the flow restarts on reload.
  }
}

const StepTracker = ({ state, onStep }: { state: TFlowState; onStep: (step: TMockStep) => void }) => {
  const wizard = useContext(WizardContext)
  const id = wizard?.steps[wizard.currentStepIndex]?.id as TMockStep | undefined
  useEffect(() => {
    if (id && id !== state.step) onStep(id)
  }, [id, state.step, onStep])
  return null
}

export const MockFlow = ({ start, persist = false }: { start: IMockFlowStart; persist?: boolean }) => {
  const initial: TFlowState = {
    step: start.step,
    path: start.path ?? 'own',
    cloud: start.cloud ?? 'aws',
    seeded: true,
  }
  const [stored] = useState(() => (persist ? readStored() : undefined))
  const [state, setState] = useState<TFlowState>(stored?.state ?? initial)
  const [runKey, setRunKey] = useState(0)
  const memoryRef = useRef<TFlowMemory>(stored?.memory ?? emptyMemory())
  const stateRef = useRef(state)

  const update = (next: TFlowState) => {
    stateRef.current = next
    setState(next)
    if (persist) store({ state: next, memory: memoryRef.current })
  }

  const memory = useMemo<TMemory>(() => {
    const save = () => {
      if (persist) store({ state: stateRef.current, memory: memoryRef.current })
    }
    return {
      start: () => memoryRef.current.start,
      run: (mock) => memoryRef.current.runs[runKeyFor(mock)] ?? {},
      setStart: (patch) => {
        memoryRef.current.start = { ...memoryRef.current.start, ...patch }
        save()
      },
      setRun: (mock, patch) => {
        const key = runKeyFor(mock)
        memoryRef.current.runs[key] = { ...memoryRef.current.runs[key], ...patch }
        save()
      },
    }
  }, [persist])

  const nav: TNav = {
    choose: (path, cloud, step) => {
      setRunKey((k) => k + 1)
      update({ path, cloud, step: step ?? (path === 'own' ? 'connect' : 'deploy'), seeded: false })
    },
    toIntro: () => update({ ...state, step: 'intro', seeded: false }),
    finish: () => update({ ...state, step: 'done', seeded: false }),
  }
  const navRef = useRef(nav)
  navRef.current = nav

  const seed = state.seeded ? (start.scenario ?? {}) : undefined
  const mock: TMock = { path: state.path, cloud: state.cloud }

  // Built once per run: new component functions would remount the step and
  // reset its simulated wait.
  const steps = useMemo(() => {
    const stableNav: TNav = {
      choose: (...args) => navRef.current.choose(...args),
      toIntro: () => navRef.current.toIntro(),
      finish: () => navRef.current.finish(),
    }
    const seedFor = (id: string) => (seed && id === state.step ? seed : undefined)
    return buildFirstRunSteps(mock.path, mock.cloud).map((def) => ({
      ...def,
      component:
        def.id === 'start'
          ? MockStart(stableNav, seedFor('start'), mock, memory)
          : def.id === 'connect'
            ? MockConnect(stableNav, seedFor('connect'), mock, memory)
            : def.id === 'deploy'
              ? MockDeploy(seedFor('deploy'), mock, memory)
              : def.id === 'stack'
                ? MockStack(seedFor('stack'), mock, memory)
                : MockProvision(stableNav, mock),
    }))
    // Rebuilt only when the run changes, never while a step is on screen.
  }, [runKey, mock.path, mock.cloud])

  if (state.step === 'intro') {
    return (
      <IntroScreen onStart={() => update({ ...state, step: 'start', path: 'own', seeded: false })} />
    )
  }
  if (state.step === 'done') {
    return (
      <FlowDone
        onRestart={() => {
          memoryRef.current = emptyMemory()
          setRunKey((k) => k + 1)
          update({ ...initial, step: 'intro', seeded: false })
        }}
      />
    )
  }

  const index = Math.max(
    0,
    steps.findIndex((def) => def.id === state.step)
  )

  return (
    <OnboardingWizardProvider
      key={`${runKey}-${mock.path}-${mock.cloud}`}
      steps={steps}
      initialStepIndex={index}
      initialSharedData={{}}
      onComplete={nav.finish}
    >
      <StepTracker state={state} onStep={(step) => update({ ...state, step })} />
      <OnboardingWizardLayout onSkip={nav.finish} />
    </OnboardingWizardProvider>
  )
}

export const MockStory = (props: { start: IMockFlowStart; persist?: boolean }) => (
  <CopyEditor>
    <NoticeHost>
      <MockFlow {...props} />
    </NoticeHost>
  </CopyEditor>
)
