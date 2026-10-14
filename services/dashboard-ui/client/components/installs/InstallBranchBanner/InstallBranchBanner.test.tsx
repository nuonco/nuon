import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { InstallBranchBanner } from './InstallBranchBanner'
import { SuccessfulBranchRuns } from './SuccessfulBranchRuns'

afterEach(cleanup)

test('hides branch lag when current or provenance is unknown', () => {
  const { rerender } = render(<InstallBranchBanner runsBehind={3} />)
  expect(screen.getByText(/3 app branch runs behind/)).toBeTruthy()
  for (const runsBehind of [0, undefined, null]) {
    rerender(<InstallBranchBanner runsBehind={runsBehind} />)
    expect(screen.queryByText(/app branch runs? behind/)).toBeNull()
  }
})

test('shows the count and tracked branch with singular copy', () => {
  render(
    <MemoryRouter>
      <InstallBranchBanner
        runsBehind={1}
        branchName="release"
        branchHref="/org-1/apps/app-1/branches/br-1"
      />
    </MemoryRouter>
  )
  expect(screen.getByText('1 app branch run behind')).toBeTruthy()
  expect(
    screen.getByRole('link', { name: 'release' }).getAttribute('href')
  ).toBe('/org-1/apps/app-1/branches/br-1')
})

test('reports a failed refresh rather than showing a stale count', () => {
  render(
    <InstallBranchBanner
      runsBehind={3}
      error={{ error: 'Unavailable', description: '', user_error: false }}
      onRetry={() => {}}
    />
  )
  expect(
    screen.getByText('Unable to load branch tracking for this install.')
  ).toBeTruthy()
  expect(screen.getByRole('button', { name: 'Try again' })).toBeTruthy()
  expect(screen.queryByText(/app branch runs behind/)).toBeNull()
})

test('shows each successful run with its commit and keeps repeated commits distinct', () => {
  render(
    <MemoryRouter>
      <SuccessfulBranchRuns
        selectedCommit={{ sha: 'abcdef1234', message: 'Selected config' }}
        repo="acme/payments"
        runs={[
          {
            run_id: 'newer',
            workflow_id: 'newer',
            sha: '1234567abc',
            message: 'Update worker',
            run_status: 'success',
            created_at: '2026-10-14T12:00:00Z',
            metadata: { trigger: 'push' },
          },
          {
            run_id: 'older',
            workflow_id: 'older',
            sha: '1234567abc',
            message: 'Update worker',
            run_status: 'success',
            created_at: '2026-10-14T11:00:00Z',
            metadata: { trigger: 'manual' },
          },
        ]}
        getRunHref={(run) => `/runs/${run?.workflow_id}`}
      />
    </MemoryRouter>
  )
  expect(
    screen
      .getAllByRole('link', { name: /^(Commit pushed|Manual run)$/ })
      .map((link) => link.getAttribute('href'))
  ).toEqual(['/runs/newer', '/runs/older'])
  expect(screen.getByText('Selected commit')).toBeTruthy()
  expect(screen.getAllByText('Update worker')).toHaveLength(2)
  expect(screen.getAllByText('Completed')).toHaveLength(2)
  expect(screen.getByText('manual')).toBeTruthy()
  expect(
    screen
      .getAllByRole('link', { name: 'Update worker' })
      .map((link) => link.getAttribute('href'))
  ).toEqual(['/runs/newer', '/runs/older'])
  expect(
    screen
      .getAllByRole('link', { name: /1234567/ })
      .map((link) => link.getAttribute('href'))
  ).toEqual([
    'https://github.com/acme/payments/commit/1234567abc',
    'https://github.com/acme/payments/commit/1234567abc',
  ])
})

test('reports when the successful run history is capped', () => {
  render(
    <MemoryRouter>
      <SuccessfulBranchRuns
        runs={[
          {
            workflow_id: 'success',
            run_status: 'success',
            created_at: '2026-10-14T12:00:00Z',
          },
        ]}
        runsBehind={60}
        getRunHref={(run) => `/runs/${run?.workflow_id}`}
      />
    </MemoryRouter>
  )
  expect(
    screen.getByText(/Showing the newest 1 of 60 successful app branch runs/)
  ).toBeTruthy()
})

test('explains how successful runs appear when history is empty', () => {
  render(<SuccessfulBranchRuns runs={[]} getRunHref={() => undefined} />)
  expect(screen.getByText('No successful runs')).toBeTruthy()
  expect(screen.getByText(/after an app branch run succeeds/)).toBeTruthy()
  expect(screen.queryByRole('link')).toBeNull()
})
