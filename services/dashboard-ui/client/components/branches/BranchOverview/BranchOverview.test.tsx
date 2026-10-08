import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { BranchOverview } from './BranchOverview'

afterEach(cleanup)

const base = {
  hasPlan: true,
  groups: [],
  rolloutHref: '/rollout',
  groupHref: (id: string) => `/groups/${id}`,
  rollout: {
    id: 'run_1',
    href: '/runs/run_1',
    source: { kind: 'manual' as const },
    title: 'Manual run',
    status: 'success',
  },
}

test('tells a build-and-validate preview that it will not update installs', () => {
  render(
    <BranchOverview {...base} previewMode="build-only" showInstalls={false} />
  )

  expect(
    screen.getByText('This preview will not update any install.')
  ).toBeTruthy()
})

test('tells a plan-only preview that it plans the selected install', () => {
  render(
    <BranchOverview {...base} previewMode="plan-only" showInstalls={false} />
  )

  expect(
    screen.getByText(
      'This preview will plan the selected install. It will not roll out to the branch.'
    )
  ).toBeTruthy()
  expect(
    screen.queryByText('This preview will not update any install.')
  ).toBeNull()
})

test('shows a tag next to the commit', () => {
  render(
    <BranchOverview
      {...base}
      showInstalls={false}
      rollout={{
        ...base.rollout,
        sha: 'a1b2c3d4e5f6',
        source: { kind: 'tag', tag: 'v1.4.2' },
      }}
    />
  )

  expect(screen.getAllByText('v1.4.2').length).toBeGreaterThan(0)
})

test('shows a pull request and its label next to the commit', () => {
  render(
    <BranchOverview
      {...base}
      showInstalls={false}
      rollout={{
        ...base.rollout,
        sha: 'a1b2c3d4e5f6',
        source: {
          kind: 'pull-request',
          number: 482,
          label: 'deploy',
        },
      }}
    />
  )

  expect(screen.getAllByText('PR #482').length).toBeGreaterThan(0)
  expect(screen.getAllByText('deploy').length).toBeGreaterThan(0)
})

test('omits the preview banner for a rollout run', () => {
  render(<BranchOverview {...base} showInstalls={false} />)

  expect(
    screen.queryByText('This preview will not update any install.')
  ).toBeNull()
  expect(screen.queryByText(/will plan the selected install/)).toBeNull()
})
