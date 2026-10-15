export default {
  title: 'Features / Onboarding / First run / Clickable flow',
}

import { useContext, useEffect, useMemo, useRef, useState } from 'react'
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
  type TPath,
} from './constants'
import { DeployStepView, type TDeployPhase } from './DeployStep'
import { IntroScreen } from './IntroScreen'
import type { IWaitItem, TWaitState } from './progress'
import { ProvisionStepView } from './ProvisionStep'
import { StackStepView, type TStackPhase } from './StackStep'
import { StartStepView } from './StartStep'
import { buildFirstRunSteps } from './steps'

// Each simulated wait ticks one checklist item forward at this pace.
const TICK_MS = 1500
const OWN_APP = 'acme-api'
const OWN_REPO = 'acme/acme-api'

type TFlow = { path: TPath; cloud: TCloud; step: string } | null

// Ladle rewrites its own URL and remounts the story, so the place in the flow
// is kept in session storage: a remount or reload opens the same screen.
const FLOW_KEY = 'first-run-clickable-flow'

const readFlow = (): TFlow => {
  try {
    const value = sessionStorage.getItem(FLOW_KEY)
    if (!value) return null
    const [path, cloud, step] = value.split('.')
    if ((path === 'own' || path === 'example') && isCloud(cloud) && step) return { path, cloud, step }
  } catch {
    // Storage blocked: the flow starts at the intro.
  }
  return null
}

const writeFlow = (flow: TFlow) => {
  try {
    if (flow) sessionStorage.setItem(FLOW_KEY, `${flow.path}.${flow.cloud}.${flow.step}`)
    else sessionStorage.removeItem(FLOW_KEY)
  } catch {
    // Storage blocked: the flow still works, it just restarts on reload.
  }
}

// Steps 0..n-1 of a wait, ticking one at a time while the screen is open.
const useTicker = (total: number, running: boolean) => {
  const [count, setCount] = useState(0)
  useEffect(() => {
    if (!running) return
    const timer = setInterval(() => setCount((c) => Math.min(c + 1, total)), TICK_MS)
    return () => clearInterval(timer)
  }, [running, total])
  return count
}

const waitStates = (labels: string[], done: number): IWaitItem[] =>
  labels.map((label, i) => ({
    label,
    state: (i < done ? 'done' : i === done ? 'active' : 'pending') as TWaitState,
  }))

const appNameFor = (path: TPath) => (path === 'own' ? OWN_APP : 'kitchen-sink')
const repoFor = (path: TPath) => (path === 'own' ? OWN_REPO : KITCHEN_SINK_REPO)

const MockDeploy = ({ path, cloud }: { path: TPath; cloud: TCloud }) =>
  function MockDeployStep({ onAdvance, onGoBack }: IWizardStepComponentProps) {
    const [phase, setPhase] = useState<TDeployPhase>('idle')
    const [region, setRegion] = useState(defaultRegion(cloud))
    const [since, setSince] = useState<number>()
    const labels = [
      'Start a branch run on main',
      `Fetch the latest commit from ${repoFor(path)}`,
      'Read and validate the app config',
      'Build the components and sandbox',
      'Create the install',
    ]
    const done = useTicker(labels.length, phase === 'waiting-config')
    // onAdvance changes identity as the wizard re-renders; advancing twice
    // would skip the stack step.
    const advanced = useRef(false)
    useEffect(() => {
      if (phase !== 'waiting-config' || done < labels.length || advanced.current) return
      advanced.current = true
      onAdvance()
    }, [phase, done, labels.length, onAdvance])
    return (
      <DeployStepView
        path={path}
        appName={appNameFor(path)}
        repo={repoFor(path)}
        cloud={cloud}
        region={region}
        onRegion={setRegion}
        phase={phase}
        installCreated={false}
        missingInputs={[]}
        progress={waitStates(labels, done)}
        progressSince={since}
        onGetHelp={() => {}}
        onCreate={() => {
          setSince(Date.now())
          setPhase('waiting-config')
        }}
        onBack={onGoBack}
      />
    )
  }

const MockStack = ({ path, cloud }: { path: TPath; cloud: TCloud }) =>
  function MockStackStep({ onAdvance, onGoBack }: IWizardStepComponentProps) {
    const [since] = useState(() => Date.now())
    const [launched, setLaunched] = useState(false)
    const labels = [
      'Create the install',
      'Plan the install workflow',
      'Generate the install state',
      "Create the runner's identity",
      STACK_RENDER_LABEL[cloud],
      `Publish the ${CLOUD_CONNECT[cloud].artifactNoun}`,
    ]
    const done = useTicker(labels.length, true)
    const phase: TStackPhase = done < labels.length ? 'generating' : launched ? 'launched' : 'ready'
    return (
      <StackStepView
        cloud={cloud}
        appName={path === 'own' ? OWN_APP : KITCHEN_SINK_LABEL}
        region={defaultRegion(cloud)}
        phase={phase}
        quickLinkUrl={cloud === 'aws' ? '#' : undefined}
        progress={waitStates(labels, done)}
        progressSince={since}
        onGetHelp={() => {}}
        onLaunch={() => setLaunched(true)}
        onContinue={onAdvance}
        onBack={onGoBack}
      />
    )
  }

const MockConnect = ({ cloud }: { cloud: TCloud }) =>
  function MockConnectStep({ onAdvance, onGoBack }: IWizardStepComponentProps) {
    const [detected, setDetected] = useState(false)
    const [confirmSkip, setConfirmSkip] = useState(false)
    useEffect(() => {
      const timer = setTimeout(() => setDetected(true), 6000)
      return () => clearTimeout(timer)
    }, [])
    return (
      <ConnectStepView
        appName={OWN_APP}
        repo={OWN_REPO}
        cloud={cloud}
        prompt="/goal Set up this application on Nuon so a customer can run it in their own cloud account."
        promptFailed={false}
        detected={detected}
        sha={detected ? 'a1b2c3d' : undefined}
        confirmSkip={confirmSkip}
        onContinue={() => (detected ? onAdvance() : setConfirmSkip(true))}
        onContinueAnyway={onAdvance}
        onKeepWaiting={() => setConfirmSkip(false)}
        onExampleExit={() => {}}
        onGetHelp={() => {}}
        onBack={onGoBack}
      />
    )
  }

const MockProvision = ({ path, cloud }: { path: TPath; cloud: TCloud }) =>
  function MockProvisionStep({ onGoBack }: IWizardStepComponentProps) {
    return (
      <ProvisionStepView
        cloud={cloud}
        appName={path === 'own' ? OWN_APP : KITCHEN_SINK_LABEL}
        region={defaultRegion(cloud)}
        installId="inlk3x9q2m7v4w8p1z6r5t0y2c"
        onFinish={() => window.alert('Finish opens the deploy workflow in the dashboard.')}
        onBack={onGoBack}
      />
    )
  }

const MockStart = ({ onChoose, onBackToIntro }: { onChoose: (path: TPath, cloud: TCloud) => void; onBackToIntro: () => void }) =>
  function MockStartStep() {
    const [expanded, setExpanded] = useState(false)
    const [appName, setAppName] = useState(OWN_APP)
    const [cloud, setCloud] = useState<TCloud>()
    return (
      <StartStepView
        expanded={expanded}
        onExpand={() => setExpanded(true)}
        onBackToIntro={onBackToIntro}
        onExitToExample={() => setExpanded(false)}
        appName={appName}
        onAppName={setAppName}
        github={{ status: 'connected', owner: 'acme', repoCount: 12 }}
        onConnectGithub={() => {}}
        cloud={cloud}
        onCloud={setCloud}
        showErrors={false}
        onNext={(values) => onChoose('own', isCloud(values.cloud) ? values.cloud : 'aws')}
        onDeployExample={(exampleCloud) => onChoose('example', exampleCloud)}
      />
    )
  }

const StepUrlSync = ({ path, cloud }: { path: TPath; cloud: TCloud }) => {
  const wizard = useContext(WizardContext)
  const id = wizard?.steps[wizard.currentStepIndex]?.id
  useEffect(() => {
    if (id) writeFlow({ path, cloud, step: id })
  }, [id, path, cloud])
  return null
}

const ClickableFlow = () => {
  const [flow, setFlow] = useState<TFlow>(() => readFlow())
  const [startKey, setStartKey] = useState(0)

  const backToIntro = () => {
    setFlow(null)
    writeFlow(null)
  }

  const path = flow?.path ?? 'own'
  const cloud = flow?.cloud ?? 'aws'

  // Built once per path and cloud: new component functions would remount the
  // step and reset its simulated wait.
  const steps = useMemo(() => {
    const choose = (nextPath: TPath, nextCloud: TCloud) => {
      setStartKey((k) => k + 1)
      setFlow({ path: nextPath, cloud: nextCloud, step: nextPath === 'own' ? 'connect' : 'deploy' })
    }
    return buildFirstRunSteps(path, cloud).map((def) => {
      const component =
        def.id === 'start'
          ? MockStart({ onChoose: choose, onBackToIntro: backToIntro })
          : def.id === 'connect'
            ? MockConnect({ cloud })
            : def.id === 'deploy'
              ? MockDeploy({ path, cloud })
              : def.id === 'stack'
                ? MockStack({ path, cloud })
                : MockProvision({ path, cloud })
      return { ...def, component }
    })
  }, [path, cloud, startKey])

  if (!flow) {
    return <IntroScreen onStart={() => setFlow({ path: 'own', cloud: 'aws', step: 'start' })} />
  }

  const index = Math.max(
    0,
    steps.findIndex((def) => def.id === flow.step)
  )

  return (
    <OnboardingWizardProvider
      key={`${path}-${cloud}-${startKey}`}
      steps={steps}
      initialStepIndex={index}
      initialSharedData={{}}
      onComplete={backToIntro}
    >
      <StepUrlSync path={path} cloud={cloud} />
      <OnboardingWizardLayout onSkip={backToIntro} />
    </OnboardingWizardProvider>
  )
}

export const ClickThrough = () => (
  <CopyEditor>
    <ClickableFlow />
  </CopyEditor>
)
ClickThrough.meta = { fullBleed: true }
