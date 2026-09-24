import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import {
  BranchActivityFeed,
  type TBranchActivityItem,
} from './BranchActivityFeed'

afterEach(cleanup)

const baseItem = (
  overrides: Partial<TBranchActivityItem> = {}
): TBranchActivityItem => ({
  appId: 'app-1',
  appName: 'acme-payments',
  branchId: 'branch-1',
  branchName: 'main',
  runId: 'run-1',
  runStatus: 'success',
  runCreatedAt: '2026-09-01T10:00:00Z',
  ...overrides,
})

test('renders feed items', () => {
  render(
    <BranchActivityFeed
      items={[
        baseItem({ appName: 'acme-payments', branchName: 'main' }),
        baseItem({
          appId: 'app-2',
          appName: 'acme-portal',
          branchId: 'branch-2',
          branchName: 'release',
          runId: 'run-2',
          runStatus: 'failed',
        }),
      ]}
    />
  )

  expect(screen.getByText('acme-payments')).toBeTruthy()
  expect(screen.getByText('acme-portal')).toBeTruthy()
})

test('renders separate update cards for multiple runs on one branch', () => {
  const { container } = render(
    <BranchActivityFeed
      items={[
        baseItem({ runId: 'run-1', commitMessage: 'first update' }),
        baseItem({ runId: 'run-2', commitMessage: 'second update' }),
      ]}
    />
  )

  expect(screen.getAllByText('acme-payments')).toHaveLength(2)
  expect(container.querySelectorAll('[data-run-id]')).toHaveLength(2)
})

test('renders a view run button per update with a run href', () => {
  render(
    <MemoryRouter>
      <BranchActivityFeed
        items={[
          baseItem({
            runId: 'run-1',
            runHref: '/org-1/apps/app-1/branches/branch-1/runs/run-1',
            commitMessage: 'first update',
          }),
          baseItem({
            runId: 'run-2',
            runHref: '/org-1/apps/app-1/branches/branch-1/runs/run-2',
            commitMessage: 'second update',
          }),
        ]}
      />
    </MemoryRouter>
  )

  expect(screen.getAllByRole('link', { name: 'View run' })).toHaveLength(2)
})

test('shows empty state when no items', () => {
  render(<BranchActivityFeed items={[]} />)
  expect(screen.getByText('No branch activity yet')).toBeTruthy()
})

test('filters items by failed status', () => {
  render(
    <BranchActivityFeed
      items={[
        baseItem({ appName: 'acme-payments', runStatus: 'success' }),
        baseItem({
          appId: 'app-2',
          appName: 'acme-portal',
          branchId: 'branch-2',
          runId: 'run-2',
          runStatus: 'failed',
        }),
      ]}
    />
  )

  fireEvent.click(screen.getByRole('button', { name: /failed/i }))

  expect(screen.queryByText('acme-payments')).toBeNull()
  expect(screen.getByText('acme-portal')).toBeTruthy()
})

test('shows no matching branches empty state when filter yields no results', () => {
  render(<BranchActivityFeed items={[baseItem({ runStatus: 'success' })]} />)

  fireEvent.click(screen.getByRole('button', { name: /failed/i }))

  expect(screen.getByText('No matching branches')).toBeTruthy()
})

test('shows loading skeleton when isLoading is true', () => {
  const { container } = render(<BranchActivityFeed items={[]} isLoading />)
  expect(container.querySelector('.animate-pulse')).toBeTruthy()
})

test('expands to show installs updated by a run', () => {
  render(
    <BranchActivityFeed
      items={[
        baseItem({
          planGroups: [{ name: 'canary', installs: 1, hasSelector: false }],
          updatedInstalls: [
            { id: 'install-1', name: 'staging-example', group: 'canary' },
          ],
        }),
      ]}
    />
  )

  expect(screen.getByText('canary')).toBeTruthy()
  fireEvent.click(screen.getByRole('button', { expanded: false }))
  expect(screen.getAllByText('staging-example').length).toBeGreaterThan(0)
})

test('renders plan dots without an expander when no installs were updated', () => {
  render(
    <BranchActivityFeed
      items={[
        baseItem({
          planGroups: [{ name: 'enterprise', installs: 0, hasSelector: false }],
        }),
      ]}
    />
  )

  expect(screen.getByText('enterprise')).toBeTruthy()
  expect(screen.queryByRole('button', { expanded: false })).toBeNull()
})

test('filter buttons are rendered', () => {
  render(<BranchActivityFeed items={[]} />)
  expect(screen.getByRole('button', { name: /all/i })).toBeTruthy()
  expect(
    screen.getByRole('button', { name: /awaiting approval/i })
  ).toBeTruthy()
  expect(screen.getByRole('button', { name: /failed/i })).toBeTruthy()
  expect(screen.getByRole('button', { name: /in progress/i })).toBeTruthy()
})
