export default {
  title: 'Features / Onboarding / First run / Deploy',
}

import { DeployStepView, type IDeployStepView } from './DeployStep'
import { StoryFrame } from './StoryFrame'

const Harness = ({ path = 'own', ...overrides }: Partial<IDeployStepView>) => {
  const cloud = overrides.cloud ?? 'aws'
  const region =
    overrides.region ?? (cloud === 'gcp' ? 'us-central1' : cloud === 'azure' ? 'eastus' : 'us-east-1')
  return (
    <StoryFrame step="deploy" path={path} cloud={cloud}>
      <DeployStepView
        path={path}
        appName={path === 'own' ? 'acme-api' : 'kitchen-sink'}
        repo={path === 'own' ? 'acme/acme-api' : 'nuonco/kitchen-sink'}
        cloud={cloud}
        phase="idle"
        installCreated={false}
        missingInputs={[]}
        onCreate={() => {}}
        onBack={() => {}}
        {...overrides}
        region={region}
      />
    </StoryFrame>
  )
}

export const OwnApp = () => <Harness />
OwnApp.meta = { fullBleed: true }

export const ExampleApp = () => <Harness path="example" cloud="gcp" />
ExampleApp.meta = { fullBleed: true }

export const WaitingForConfig = () => (
  <Harness
    phase="waiting-config"
    progressSince={Date.now() - 14_000}
    progress={[
      { label: 'Start a branch run on main', state: 'done' },
      { label: 'Fetch the latest commit from acme/acme-api', state: 'done' },
      { label: 'Read and validate the app config', state: 'done' },
      { label: 'Build the components and sandbox', state: 'active' },
      { label: 'Create the install', state: 'pending' },
    ]}
  />
)
WaitingForConfig.meta = { fullBleed: true }

export const SyncFailed = () => (
  <Harness
    phase="waiting-config"
    syncError="workflow stopped: no such file or directory"
    onRetry={() => {}}
    onGetHelp={() => {}}
  />
)
SyncFailed.meta = { fullBleed: true }

export const ExampleSyncFailed = () => (
  <Harness
    path="example"
    phase="waiting-config"
    syncError="Component builds failed."
    onRetry={() => {}}
    onGetHelp={() => {}}
    progressSince={Date.now() - 4 * 60_000}
    progress={[
      { label: 'Start a branch run on main', state: 'done' },
      { label: 'Fetch the latest commit from nuonco/kitchen-sink', state: 'done' },
      { label: 'Read and validate the app config', state: 'done' },
      { label: 'Build the components and sandbox', state: 'error' },
      { label: 'Create the install', state: 'pending' },
    ]}
  />
)
ExampleSyncFailed.meta = { fullBleed: true }

export const SlowWait = () => (
  <Harness
    path="example"
    phase="waiting-config"
    onGetHelp={() => {}}
    progressSince={Date.now() - 11 * 60_000}
    progress={[
      { label: 'Start a branch run on main', state: 'done' },
      { label: 'Fetch the latest commit from nuonco/kitchen-sink', state: 'done' },
      { label: 'Read and validate the app config', state: 'done' },
      { label: 'Build the components and sandbox', state: 'active' },
      { label: 'Create the install', state: 'pending' },
    ]}
  />
)
SlowWait.meta = { fullBleed: true }

export const MissingInputDefaults = () => <Harness cloud="azure" missingInputs={['db_password', 'api_domain']} />
MissingInputDefaults.meta = { fullBleed: true }

export const KitchenSinkAws = () => <Harness path="example" />
KitchenSinkAws.meta = { fullBleed: true }

export const KitchenSinkWaitingForConfig = () => (
  <Harness
    path="example"
    phase="waiting-config"
    onGetHelp={() => {}}
    progressSince={Date.now() - 74_000}
    progress={[
      { label: 'Start a branch run on main', state: 'done' },
      { label: 'Fetch the latest commit from nuonco/kitchen-sink', state: 'done' },
      { label: 'Read and validate the app config', state: 'done' },
      { label: 'Build the components and sandbox', state: 'active' },
      { label: 'Create the install', state: 'pending' },
    ]}
  />
)
KitchenSinkWaitingForConfig.meta = { fullBleed: true }
