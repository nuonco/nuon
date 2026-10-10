import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { InstallBranchBanner } from './InstallBranchBanner'
import { PendingCommits } from './PendingCommits'

afterEach(cleanup)

test('hides branch lag when current or provenance is unknown', () => {
  const { rerender } = render(<InstallBranchBanner commitsBehind={3} />)
  expect(screen.getByText(/3 commits behind/)).toBeTruthy()
  for (const commitsBehind of [0, undefined, null]) {
    rerender(<InstallBranchBanner commitsBehind={commitsBehind} />)
    expect(screen.queryByText(/commits? behind/)).toBeNull()
  }
})

test('shows the count and tracked branch with singular copy', () => {
  render(
    <MemoryRouter>
      <InstallBranchBanner
        commitsBehind={1}
        branchName="release"
        branchHref="/org-1/apps/app-1/branches/br-1"
      />
    </MemoryRouter>
  )
  expect(screen.getByText('1 commit behind')).toBeTruthy()
  expect(
    screen.getByRole('link', { name: 'release' }).getAttribute('href')
  ).toBe('/org-1/apps/app-1/branches/br-1')
})

test('reports a failed refresh rather than showing a stale count', () => {
  render(
    <InstallBranchBanner
      commitsBehind={3}
      error={{ error: 'Unavailable', description: '', user_error: false }}
      onRetry={() => {}}
    />
  )
  expect(
    screen.getByText('Unable to load branch tracking for this install.')
  ).toBeTruthy()
  expect(screen.getByRole('button', { name: 'Try again' })).toBeTruthy()
  expect(screen.queryByText(/commits behind/)).toBeNull()
})

test('defaults to active branch runs and lets the user review terminal history', () => {
  const commits = [
    { workflow_id: 'pending', run_status: 'pending' },
    { workflow_id: 'running', run_status: 'running', awaiting_approval: true },
    { workflow_id: 'success', run_status: 'success' },
    { workflow_id: 'failed', run_status: 'failed' },
    {
      workflow_id: 'cancelled',
      run_status: 'cancelled',
      awaiting_approval: true,
    },
    { workflow_id: 'not-attempted', run_status: 'not-attempted' },
    { workflow_id: 'unknown' },
  ]
  render(
    <MemoryRouter>
      <PendingCommits
        commits={commits}
        hrefFor={(commit) => `/runs/${commit?.workflow_id}`}
      />
    </MemoryRouter>
  )
  expect(
    screen
      .getAllByRole('link', { name: 'View branch run' })
      .map((link) => link.getAttribute('href'))
  ).toEqual(['/runs/pending', '/runs/running'])
  expect(screen.getAllByText('Awaiting approval')).toHaveLength(1)
  expect(
    screen
      .getByRole('button', { name: 'Active runs' })
      .getAttribute('aria-pressed')
  ).toBe('true')

  fireEvent.click(screen.getByRole('button', { name: 'All commits' }))
  expect(
    screen
      .getAllByRole('link', { name: 'View branch run' })
      .map((link) => link.getAttribute('href'))
  ).toEqual([
    '/runs/pending',
    '/runs/running',
    '/runs/success',
    '/runs/failed',
    '/runs/cancelled',
    '/runs/not-attempted',
    '/runs/unknown',
  ])
  fireEvent.click(screen.getByRole('button', { name: 'Active runs' }))
  expect(screen.getAllByRole('link', { name: 'View branch run' })).toHaveLength(
    2
  )
})

test('explains when no active runs exist in the available commit history', () => {
  render(
    <MemoryRouter>
      <PendingCommits
        commits={[{ workflow_id: 'cancelled', run_status: 'cancelled' }]}
        commitsBehind={60}
        hrefFor={(commit) => `/runs/${commit?.workflow_id}`}
      />
    </MemoryRouter>
  )
  expect(screen.getByText('No active runs')).toBeTruthy()
  expect(screen.getByText(/No active runs in the newest commits/)).toBeTruthy()
  expect(screen.queryByRole('link', { name: 'View branch run' })).toBeNull()
  fireEvent.click(screen.getByRole('button', { name: 'All commits' }))
  expect(screen.queryByText('No active runs')).toBeNull()
  expect(
    screen.getByRole('link', { name: 'View branch run' }).getAttribute('href')
  ).toBe('/runs/cancelled')
})
