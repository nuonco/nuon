import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { ResourceScopeSummary, StepContext } from './DeploymentProgress'

afterEach(cleanup)

test('a partially applied resource is labelled as a partial rollout', () => {
  render(
    <ResourceScopeSummary
      run={{
        status: 'failed-pending-retry',
        activity: '',
        steps: [],
        outcomes: [
          {
            category: 'Sandbox',
            name: 'Sandbox',
            status: 'warn',
            detail: 'Applied; remaining steps not completed',
            applied: true,
          },
        ],
      }}
    />
  )
  expect(
    screen.getByRole('group', { name: 'Sandbox: Partial rollout' })
  ).toBeTruthy()
})

test('a queued deployment without steps shows that steps are being prepared', () => {
  render(
    <StepContext
      run={{ status: 'pending', activity: 'queued', steps: [], outcomes: [] }}
    />
  )
  expect(screen.getByText('Preparing steps')).toBeTruthy()
  expect(screen.getByText('Waiting for steps')).toBeTruthy()
})

test('cancelled and skipped categories are labelled explicitly', () => {
  render(
    <ResourceScopeSummary
      run={{
        status: 'cancelled',
        activity: '',
        steps: [],
        outcomes: [
          {
            category: 'Sandbox',
            name: 'Sandbox',
            status: 'cancelled',
            detail: 'Cancelled',
          },
          {
            category: 'Components',
            name: 'api',
            status: 'user-skipped',
            detail: 'Skipped by operator',
          },
        ],
      }}
    />
  )
  expect(screen.getByRole('group', { name: 'Sandbox: Cancelled' })).toBeTruthy()
  expect(
    screen.getByRole('group', { name: 'Components: Skipped' })
  ).toBeTruthy()
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
