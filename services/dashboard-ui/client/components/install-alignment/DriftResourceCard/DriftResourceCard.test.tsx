import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import {
  inputUpdateResource,
  manualOrigin,
  openDeploymentResource,
  stateDriftResource,
  unappliedResource,
} from '@/components/__fixtures__/install-alignment'
import type { TInstallResourceSyncStatusOrigin } from '@/types'
import { DriftResourceCard } from './DriftResourceCard'

afterEach(cleanup)

const originHref = (origin?: TInstallResourceSyncStatusOrigin) =>
  origin?.run
    ? `/runs/${origin.run.workflow_id}`
    : `/workflows/${origin?.workflow_id}`

const renderCard = (props: Parameters<typeof DriftResourceCard>[0]) =>
  render(
    <MemoryRouter>
      <DriftResourceCard {...props} />
    </MemoryRouter>
  )

test('a version change shows the deployed branch run, the reason, and deploy history', () => {
  renderCard({
    resource: unappliedResource(),
    historyHref: '/components/cmp-api/deploys',
    originHref,
  })

  const builds = screen.getByTestId('live-build')
  const [live] = within(builds).getAllByTestId('run-ref')
  expect(builds.textContent).toStartWith(
    'DeployedPR #412Pin the ledger module to 2.4.0·'
  )
  expect(
    within(live).queryByRole('link', { name: /Pull request 412/ })
  ).toBeNull()
  expect(
    within(live)
      .getByRole('link', { name: 'Pin the ledger module to 2.4.0' })
      .getAttribute('href')
  ).toBe('/runs/wf-branch-run-1')
  expect(screen.getAllByTestId('run-ref')).toHaveLength(1)
  const summary = screen.getByTestId('drift-summary')
  expect(summary.textContent).toStartWith(
    'The deployed version doesn’t match this install’s configuration.'
  )
  expect(summary.textContent).not.toContain('bld-')
  expect(
    within(summary)
      .getByRole('link', { name: 'View deploy history' })
      .getAttribute('href')
  ).toBe('/components/cmp-api/deploys')
})

test('a live version from outside a branch run names its workflow', () => {
  renderCard({
    resource: unappliedResource({
      applied_by: manualOrigin,
    }),
    originHref,
  })

  const builds = screen.getByTestId('live-build')
  expect(
    within(builds)
      .getByRole('link', { name: 'Manual deploy' })
      .getAttribute('href')
  ).toBe('/workflows/wf-manual-deploy')
  expect(within(builds).getAllByTestId('run-ref')).toHaveLength(1)
})

test('changed inputs on the live build keep the live run and show the reason', () => {
  renderCard({ resource: inputUpdateResource(), originHref })

  expect(screen.getByTestId('live-build')).toBeTruthy()
  expect(screen.getByTestId('drift-summary').textContent).toBe(
    'Saved input changes haven’t taken effect.'
  )
})

test('an open deployment shows its status and links to it instead of resolving', () => {
  renderCard({
    resource: openDeploymentResource(),
    deploymentHref: '/workflows/wf-open-ic-api',
    onResolve: () => {},
  })

  expect(screen.getByText('Failed — awaiting retry')).toBeTruthy()
  expect(
    screen.getByRole('link', { name: 'View deployment' }).getAttribute('href')
  ).toBe('/workflows/wf-open-ic-api')
  expect(screen.queryByRole('button', { name: /Resolve drift/ })).toBeNull()
})

test('state drift keeps the live run and links to the drift check, not deploy history', () => {
  renderCard({
    resource: stateDriftResource(),
    driftCheckHref: '/workflows/wf-plan-ic-api',
    historyHref: '/components/cmp-api/deploys',
    originHref,
  })

  expect(screen.getByTestId('live-build')).toBeTruthy()
  expect(screen.getByText('Drift detected')).toBeTruthy()
  expect(
    screen.getByRole('link', { name: 'View drift scan' }).getAttribute('href')
  ).toBe('/workflows/wf-plan-ic-api')
  expect(screen.queryByRole('link', { name: 'View deploy history' })).toBeNull()
})
