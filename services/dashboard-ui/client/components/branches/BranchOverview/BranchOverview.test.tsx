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
  render(<BranchOverview {...base} previewMode="build-only" showInstalls={false} />)

  expect(
    screen.getByText('This preview will not update any install.')
  ).toBeTruthy()
})

test('tells a plan-only preview that it plans the selected install', () => {
  render(<BranchOverview {...base} previewMode="plan-only" showInstalls={false} />)

  expect(
    screen.getByText(
      'This preview will plan the selected install. It will not roll out to the branch.'
    )
  ).toBeTruthy()
  expect(
    screen.queryByText('This preview will not update any install.')
  ).toBeNull()
})

test('omits the preview banner for a rollout run', () => {
  render(<BranchOverview {...base} showInstalls={false} />)

  expect(
    screen.queryByText('This preview will not update any install.')
  ).toBeNull()
  expect(
    screen.queryByText(/will plan the selected install/)
  ).toBeNull()
})
