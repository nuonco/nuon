import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import {
  inputUpdateResource,
  manualOrigin,
  openDeploymentResource,
  unappliedResource,
} from '@/components/__fixtures__/install-alignment'
import type { TInstallAlignmentOrigin } from '@/types'
import { DriftResourceCard } from './DriftResourceCard'

afterEach(cleanup)

const originHref = (origin?: TInstallAlignmentOrigin) =>
  origin?.run
    ? `/runs/${origin.run.workflow_id}`
    : `/workflows/${origin?.workflow_id}`

const renderCard = (props: Parameters<typeof DriftResourceCard>[0]) =>
  render(
    <MemoryRouter>
      <DriftResourceCard {...props} />
    </MemoryRouter>
  )

test('a version change shows the running branch run, the blocking attempt, and earlier attempts', () => {
  renderCard({
    resource: unappliedResource(),
    explanationHref: '/workflows/wf-attempt-ic-api',
    historyHref: '/components/cmp-api/deploys',
    originHref,
    repo: 'acme/payments',
  })

  const builds = screen.getByTestId('drift-builds')
  const [live] = within(builds).getAllByTestId('run-ref')
  expect(builds.textContent).toStartWith(
    'RunningPR #412Pin the ledger module to 2.4.0· deployed'
  )
  expect(
    within(live)
      .getByRole('link', { name: /Pull request 412/ })
      .getAttribute('href')
  ).toBe('https://github.com/acme/payments/pull/412')
  expect(
    within(live)
      .getByRole('link', { name: 'Pin the ledger module to 2.4.0' })
      .getAttribute('href')
  ).toBe('/runs/wf-branch-run-1')
  expect(builds.textContent).not.toContain('bld-')
  const runs = screen.getAllByTestId('run-ref')
  expect(runs).toHaveLength(2)
  expect(
    within(runs[1])
      .getByRole('link', { name: /Tag v1.4.2/ })
      .getAttribute('href')
  ).toBe('https://github.com/acme/payments/releases/tag/v1.4.2')
  expect(
    within(runs[1])
      .getByRole('link', { name: 'Raise the checkout worker retry budget' })
      .getAttribute('href')
  ).toBe('/runs/wf-branch-run-6')
  expect(screen.queryByTestId('drift-consequence')).toBeNull()
  expect(screen.getByText('Deployment failed')).toBeTruthy()
  expect(screen.getByText('The latest deployment failed.')).toBeTruthy()
  expect(
    screen.getByRole('link', { name: 'View workflow' }).getAttribute('href')
  ).toBe('/workflows/wf-attempt-ic-api')
  const history = screen.getByTestId('drift-history')
  expect(history.textContent).toContain(
    '2 earlier attempts since the live deploy'
  )
  expect(
    within(history)
      .getByRole('link', { name: 'View deploy history' })
      .getAttribute('href')
  ).toBe('/components/cmp-api/deploys')
  expect(screen.queryByText(/branch run/)).toBeNull()
  expect(screen.queryByText(/Last attempt/)).toBeNull()
})

test('a live version from outside a branch run names its workflow', () => {
  renderCard({
    resource: unappliedResource({
      applied_by: manualOrigin,
      expected_by: undefined,
    }),
    originHref,
  })

  const builds = screen.getByTestId('drift-builds')
  expect(
    within(builds)
      .getByRole('link', { name: 'Manual deploy' })
      .getAttribute('href')
  ).toBe('/workflows/wf-manual-deploy')
  expect(within(builds).getAllByTestId('run-ref')).toHaveLength(1)
})

test('changed inputs on the live build show the reason instead of a build change', () => {
  renderCard({
    resource: inputUpdateResource(),
    explanationHref: '/workflows/input-update?panel=deploy',
    historyHref: '/components/cmp-api/deploys',
  })

  expect(screen.queryByTestId('drift-builds')).toBeNull()
  expect(screen.getByTestId('drift-consequence').textContent).toBe(
    'Saved input changes haven’t taken effect.'
  )
  expect(screen.getByText('Workflow cancelled')).toBeTruthy()
  expect(
    screen.getByRole('link', { name: 'View workflow' }).getAttribute('href')
  ).toBe('/workflows/input-update?panel=deploy')
  expect(screen.queryByTestId('drift-history')).toBeNull()
})

test('an open deployment links only to itself and counts the superseded failure as earlier', () => {
  renderCard({
    resource: openDeploymentResource({ earlier_attempts: 0 }),
    deploymentHref: '/workflows/wf-open-ic-api',
    explanationHref: '/workflows/wf-open-ic-api',
  })

  expect(
    screen.getByRole('link', { name: 'View deployment' }).getAttribute('href')
  ).toBe('/workflows/wf-open-ic-api')
  expect(screen.queryByRole('link', { name: 'View workflow' })).toBeNull()
  expect(screen.getByText('Failed — awaiting retry')).toBeTruthy()
  expect(screen.getByTestId('drift-history').textContent).toContain(
    '1 earlier attempt since'
  )
})

test('unattributed changes explain the missing evidence without inventing a workflow', () => {
  renderCard({
    resource: unappliedResource({
      expected_build: undefined,
      applied_build: undefined,
      reason: 'This component hasn’t been deployed yet.',
      explanation: undefined,
      blocking_attempt: undefined,
      earlier_attempts: 0,
      plan: undefined,
    }),
  })
  expect(screen.getByTestId('drift-consequence').textContent).toBe(
    'This component hasn’t been deployed yet.'
  )
  expect(
    screen.getByText(
      'We couldn’t determine why these changes haven’t taken effect.'
    )
  ).toBeTruthy()
  expect(screen.queryByRole('link')).toBeNull()
  expect(screen.queryByText('Workflow cancelled')).toBeNull()
})
