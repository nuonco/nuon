import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { InstallBranchBanner } from './InstallBranchBanner'

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
