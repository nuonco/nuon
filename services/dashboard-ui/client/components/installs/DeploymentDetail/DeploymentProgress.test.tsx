import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { StepContext } from './DeploymentProgress'

afterEach(cleanup)

test('a queued deployment without steps shows that steps are being prepared', () => {
  render(
    <StepContext
      run={{ status: 'pending', activity: 'queued', steps: [], outcomes: [] }}
    />
  )
  expect(screen.getByText('Preparing steps')).toBeTruthy()
  expect(screen.getByText('Waiting for steps')).toBeTruthy()
})

test('a plan being checked is labelled under the current step', () => {
  render(
    <StepContext
      run={{
        status: 'in-progress',
        activity: '',
        steps: [
          {
            id: 'plan',
            name: 'sync and plan api',
            status: { status: 'checking-plan' },
          },
        ],
        outcomes: [],
      }}
    />
  )
  expect(screen.getByText('sync and plan api')).toBeTruthy()
  expect(screen.getByText('Checking plan')).toBeTruthy()
})
