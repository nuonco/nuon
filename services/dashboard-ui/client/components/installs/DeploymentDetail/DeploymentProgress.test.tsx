import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { ResourceScopeSummary, StepContext } from './DeploymentProgress'

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

test('category summaries stay on one line of labels', () => {
  render(
    <ResourceScopeSummary
      run={{
        status: 'error',
        outcomes: [
          {
            category: 'Sandbox',
            name: 'Sandbox',
            status: 'warn',
            detail: 'Partial rollout',
          },
          {
            category: 'Secrets',
            name: 'Secrets',
            status: 'error',
            detail: 'Failed',
          },
          {
            category: 'Components',
            name: 'api',
            status: 'not-started',
            detail: 'Not started',
          },
        ],
      }}
    />
  )
  expect(
    screen.getByRole('group', { name: 'Sandbox: Partial rollout' })
  ).toBeTruthy()
  expect(screen.getByRole('group', { name: 'Secrets: Failed' })).toBeTruthy()
  expect(
    screen.getByRole('group', { name: 'Components: Not started' })
  ).toBeTruthy()
  expect(screen.queryByRole('listitem')).toBeNull()
})

test('cancelled and skipped categories are labelled explicitly', () => {
  render(
    <ResourceScopeSummary
      run={{
        status: 'cancelled',
        outcomes: [
          {
            category: 'Secrets',
            name: 'Secrets',
            status: 'cancelled',
            detail: 'Cancelled',
          },
          {
            category: 'Components',
            name: 'api',
            status: 'user-skipped',
            detail: 'Skipped by operator',
          },
          {
            category: 'Sandbox',
            name: 'Sandbox',
            status: 'auto-skipped',
            detail: 'Apply skipped by policy',
          },
        ],
      }}
    />
  )
  expect(screen.getByRole('group', { name: 'Secrets: Cancelled' })).toBeTruthy()
  expect(
    screen.getByRole('group', { name: 'Components: Skipped' })
  ).toBeTruthy()
  expect(
    screen.getByRole('group', { name: 'Sandbox: Skipped by policy' })
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
