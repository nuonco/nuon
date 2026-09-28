export default {
  title: 'Onboarding/First run/Start',
}

import { useState } from 'react'
import type { TCloud } from './constants'
import { StartStepView, type IStartStepView, type TGithubTile } from './StartStep'
import { StoryFrame } from './StoryFrame'

const Harness = ({
  path = 'own',
  ...overrides
}: Partial<IStartStepView> & { path?: 'own' | 'example' }) => {
  const [appName, setAppName] = useState(overrides.appName ?? '')
  const [cloud, setCloud] = useState<TCloud | undefined>(overrides.cloud)
  return (
    <StoryFrame step="start" path={path}>
      <StartStepView
        expanded={path === 'own'}
        onExpand={() => {}}
        onBackToIntro={() => {}}
        onExitToExample={() => {}}
        github={{ status: 'disconnected' }}
        connectHref="https://github.com/apps/nuon-dev/installations/new"
        onConnectGithub={() => {}}
        showErrors={false}
        onNext={() => {}}
        onDeployExample={() => {}}
        {...overrides}
        appName={appName}
        onAppName={setAppName}
        cloud={cloud}
        onCloud={setCloud}
      />
    </StoryFrame>
  )
}

const connected: TGithubTile = { status: 'connected', owner: 'acme', repoCount: 3 }

export const ChoosePath = () => <Harness path="example" />
ChoosePath.meta = { fullBleed: true }

export const OwnAppDisconnected = () => <Harness />
OwnAppDisconnected.meta = { fullBleed: true }

export const OwnAppConnected = () => <Harness github={connected} appName="acme-api" cloud="aws" />
OwnAppConnected.meta = { fullBleed: true }

export const MissingFields = () => <Harness showErrors />
MissingFields.meta = { fullBleed: true }

export const InvalidName = () => <Harness github={connected} appName="My App" />
InvalidName.meta = { fullBleed: true }

export const NoMatchingRepo = () => (
  <Harness github={connected} appName="acme-api" cloud="gcp" repoError="No repo named acme-api in your GitHub connection." />
)
NoMatchingRepo.meta = { fullBleed: true }

export const NameTaken = () => (
  <Harness github={connected} appName="acme-api" cloud="azure" appNameError="An app template with this name exists" />
)
NameTaken.meta = { fullBleed: true }

export const GithubCallbackFailed = () => <Harness github={{ status: 'disconnected', error: true }} />
GithubCallbackFailed.meta = { fullBleed: true }

export const CreatingExample = () => <Harness path="example" examplePending="aws" />
CreatingExample.meta = { fullBleed: true }
