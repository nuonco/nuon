import { expect, test } from 'bun:test'
import { render, screen, fireEvent } from '@testing-library/react'
import { BranchActivityFeed, type TBranchActivityItem } from './BranchActivityFeed'

const baseItem = (overrides: Partial<TBranchActivityItem> = {}): TBranchActivityItem => ({
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
    />,
  )

  expect(screen.getByText('acme-payments')).toBeTruthy()
  expect(screen.getByText('acme-portal')).toBeTruthy()
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
    />,
  )

  fireEvent.click(screen.getByRole('button', { name: /failed/i }))

  expect(screen.queryByText('acme-payments')).toBeNull()
  expect(screen.getByText('acme-portal')).toBeTruthy()
})

test('shows no matching branches empty state when filter yields no results', () => {
  render(
    <BranchActivityFeed
      items={[baseItem({ runStatus: 'success' })]}
    />,
  )

  fireEvent.click(screen.getByRole('button', { name: /failed/i }))

  expect(screen.getByText('No matching branches')).toBeTruthy()
})

test('shows loading skeleton when isLoading is true', () => {
  const { container } = render(<BranchActivityFeed items={[]} isLoading />)
  expect(container.querySelector('.animate-pulse')).toBeTruthy()
})

test('filter buttons are rendered', () => {
  render(<BranchActivityFeed items={[]} />)
  expect(screen.getByRole('button', { name: /all/i })).toBeTruthy()
  expect(screen.getByRole('button', { name: /awaiting approval/i })).toBeTruthy()
  expect(screen.getByRole('button', { name: /failed/i })).toBeTruthy()
  expect(screen.getByRole('button', { name: /in progress/i })).toBeTruthy()
})
