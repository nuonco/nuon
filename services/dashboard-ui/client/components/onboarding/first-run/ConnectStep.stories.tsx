export default {
  title: 'Onboarding/First run/Connect',
}

import { useState } from 'react'
import { ConnectStepView, type IConnectStepView } from './ConnectStep'
import { StoryFrame } from './StoryFrame'

const PROMPT =
  '/goal Set up this application on Nuon (nuon.co) so a customer can run it in their own cloud account. This is sample text standing in for the file served at nuon.co/loop.md.'

const Harness = (overrides: Partial<IConnectStepView>) => {
  const [confirmSkip, setConfirmSkip] = useState(overrides.confirmSkip ?? false)
  return (
    <StoryFrame step="connect">
      <ConnectStepView
        appName="acme-api"
        repo="acme/acme-api"
        cloud="aws"
        prompt={PROMPT}
        promptFailed={false}
        detected={false}
        onContinue={() => setConfirmSkip(true)}
        onContinueAnyway={() => {}}
        onKeepWaiting={() => setConfirmSkip(false)}
        onExampleExit={() => {}}
        onGetHelp={() => {}}
        onBack={() => {}}
        {...overrides}
        confirmSkip={confirmSkip}
      />
    </StoryFrame>
  )
}

export const WaitingForPush = () => <Harness />
WaitingForPush.meta = { fullBleed: true }

export const PushWarning = () => <Harness confirmSkip />
PushWarning.meta = { fullBleed: true }

export const PushDetected = () => <Harness detected sha="a1b2c3d" />
PushDetected.meta = { fullBleed: true }

export const PromptUnavailable = () => <Harness prompt={undefined} promptFailed cloud="gcp" />
PromptUnavailable.meta = { fullBleed: true }
