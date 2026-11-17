import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import type { TAppBranch } from '@/types'
import { BranchStep } from './BranchStep'

const noop = () => {}

const branches: TAppBranch[] = [
  { id: 'branch-1', name: 'main', configs: [] } as unknown as TAppBranch,
]

afterEach(cleanup)

test('flag off shows the skip option and its exact text', () => {
  render(
    <BranchStep
      branches={branches}
      selected={null}
      allowSkip
      onSelect={noop}
      onSkip={noop}
    />
  )

  expect(screen.getByRole('button', { name: 'Skip app branch' })).toBeTruthy()
  expect(
    screen.getByText(
      "This install will use the latest sync config and won't belong to an app branch."
    )
  ).toBeTruthy()
})

test('when app sync is disabled there is no skip option', () => {
  render(
    <BranchStep
      branches={branches}
      selected={null}
      allowSkip={false}
      onSelect={noop}
      onSkip={noop}
    />
  )

  expect(screen.queryByRole('button', { name: 'Skip app branch' })).toBeNull()
  expect(
    screen.queryByText(
      "This install will use the latest sync config and won't belong to an app branch."
    )
  ).toBeNull()
})
