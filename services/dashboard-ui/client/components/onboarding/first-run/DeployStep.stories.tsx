export default {
  title: 'Onboarding/First run/Deploy',
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

export const WaitingForConfig = () => <Harness phase="waiting-config" />
WaitingForConfig.meta = { fullBleed: true }

export const SyncFailed = () => (
  <Harness phase="waiting-config" syncError="workflow stopped: no such file or directory" />
)
SyncFailed.meta = { fullBleed: true }

export const MissingInputDefaults = () => <Harness cloud="azure" missingInputs={['db_password', 'api_domain']} />
MissingInputDefaults.meta = { fullBleed: true }
