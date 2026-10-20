import { afterEach, expect, mock, test } from 'bun:test'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { APP_NAME_RULE } from './constants'
import { DeployStepView } from './DeployStep'
import { StartStepView, type IStartStepView } from './StartStep'

afterEach(cleanup)

const start = (overrides: Partial<IStartStepView> = {}) => {
  const onNext = mock(() => {})
  render(
    <StartStepView
      expanded
      onExpand={() => {}}
      onBackToIntro={() => {}}
      onExitToExample={() => {}}
      appName=""
      onAppName={() => {}}
      github={{ status: 'connected', owner: 'acme', repoCount: 1 }}
      onConnectGithub={() => {}}
      onCloud={() => {}}
      showErrors={false}
      onDeployExample={() => {}}
      {...overrides}
      onNext={onNext}
    />
  )
  return { onNext }
}

test('an invalid app name shows the rule and disables Next', async () => {
  start({ appName: 'My App' })
  const name = screen.getByLabelText('App template name')
  await waitFor(() => {
    expect(name).toHaveAttribute('aria-invalid', 'true')
  })
  expect(document.querySelector('#first-run-app-name-description')).toHaveTextContent(APP_NAME_RULE)
  expect(screen.getByRole('button', { name: /^Next/ })).toHaveAttribute('aria-disabled', 'true')
  expect(name).toHaveValue('My App')
})

test('Next stays enabled until a name is typed, then reveals empty-field errors', async () => {
  const { onNext } = start()
  const next = screen.getByRole('button', { name: /^Next/ })
  expect(next).toBeEnabled()
  fireEvent.click(next)
  await waitFor(() => {
    expect(screen.getByText('Name your app template to continue.')).toBeInTheDocument()
    expect(screen.getByText('Select a test cloud to continue.')).toBeInTheDocument()
  })
  expect(onNext).not.toHaveBeenCalled()
})

test('a valid name and cloud submit those values', async () => {
  const { onNext } = start({ appName: 'acme-api', cloud: 'aws' })
  fireEvent.click(screen.getByRole('button', { name: /^Next/ }))
  await waitFor(() => {
    expect(onNext).toHaveBeenCalledWith({ appName: 'acme-api', cloud: 'aws' })
  })
})

test('a valid name on the path picker does not crash', () => {
  start({ expanded: false, appName: 'kitchen-sink', cloud: 'aws' })
  expect(screen.getByRole('heading', { name: 'Create your first app template' })).toBeInTheDocument()
})

test('a taken name is shown on the name field', async () => {
  start({
    appName: 'acme-api',
    cloud: 'aws',
    github: { status: 'connected', owner: 'acme', repoCount: 1 },
    appNameError: 'An app template with this name exists',
  })
  await waitFor(() => {
    expect(screen.getByText('An app template with this name exists')).toBeInTheDocument()
  })
})

test('deploy submits the region and auto-approve choice', async () => {
  const onCreate = mock(() => {})
  render(
    <DeployStepView
      path="own"
      appName="acme-api"
      repo="acme/acme-api"
      cloud="aws"
      region="us-east-1"
      phase="idle"
      installCreated={false}
      missingInputs={[]}
      onCreate={onCreate}
    />
  )
  fireEvent.click(screen.getByRole('button', { name: 'Create install' }))
  await waitFor(() => {
    expect(onCreate).toHaveBeenCalledWith({ region: 'us-east-1', autoApprove: true })
  })
})

test('deploy shows live progress and a BYOC reference while waiting for the app config', () => {
  render(
    <DeployStepView
      path="example"
      appName="kitchen-sink"
      repo="nuonco/kitchen-sink"
      cloud="aws"
      region="us-east-1"
      phase="waiting-config"
      installCreated={false}
      missingInputs={[]}
      progressSince={Date.now() - 5_000}
      progress={[
        { label: 'Start a branch run on main', state: 'done' },
        { label: 'Read and validate the app config', state: 'active' },
      ]}
      onCreate={() => {}}
    />
  )
  expect(screen.getByText('Setting up your test install')).toBeInTheDocument()
  expect(screen.getByText('Read and validate the app config')).toBeInTheDocument()
  expect(screen.getByText('0:05')).toBeInTheDocument()
  expect(screen.getByText('Datadog')).toBeInTheDocument()
  expect(screen.getByText('See all on Awesome BYOC')).toBeInTheDocument()
})

test('a failed Kitchen Sink sync offers a retry and help instead of a push', () => {
  const onRetry = mock(() => {})
  const onGetHelp = mock(() => {})
  render(
    <DeployStepView
      path="example"
      appName="kitchen-sink"
      repo="nuonco/kitchen-sink"
      cloud="aws"
      region="us-east-1"
      phase="waiting-config"
      installCreated={false}
      missingInputs={[]}
      syncError="Component builds failed."
      onRetry={onRetry}
      onGetHelp={onGetHelp}
      onCreate={() => {}}
    />
  )
  expect(screen.getByText('The Kitchen Sink app config did not sync')).toBeInTheDocument()
  expect(screen.queryByText(/Push a fix/)).toBeNull()
  fireEvent.click(screen.getByRole('button', { name: /Run the sync again/ }))
  fireEvent.click(screen.getByRole('button', { name: /Get help/ }))
  expect(onRetry).toHaveBeenCalledTimes(1)
  expect(onGetHelp).toHaveBeenCalledTimes(1)
})

test('a wait past the slow threshold says so and offers help', () => {
  render(
    <DeployStepView
      path="example"
      appName="kitchen-sink"
      repo="nuonco/kitchen-sink"
      cloud="aws"
      region="us-east-1"
      phase="waiting-config"
      installCreated={false}
      missingInputs={[]}
      progressSince={Date.now() - 11 * 60_000}
      progress={[{ label: 'Build the components and sandbox', state: 'active' }]}
      onGetHelp={() => {}}
      onCreate={() => {}}
    />
  )
  expect(screen.getByText(/taking longer than usual/)).toBeInTheDocument()
  expect(screen.getByRole('button', { name: /Get help/ })).toBeInTheDocument()
})

test('a normal wait shows no slow note', () => {
  render(
    <DeployStepView
      path="example"
      appName="kitchen-sink"
      repo="nuonco/kitchen-sink"
      cloud="aws"
      region="us-east-1"
      phase="waiting-config"
      installCreated={false}
      missingInputs={[]}
      progressSince={Date.now() - 60_000}
      progress={[{ label: 'Build the components and sandbox', state: 'active' }]}
      onGetHelp={() => {}}
      onCreate={() => {}}
    />
  )
  expect(screen.queryByText(/taking longer than usual/)).toBeNull()
})
