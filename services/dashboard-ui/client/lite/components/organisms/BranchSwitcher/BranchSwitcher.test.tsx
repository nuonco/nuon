import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { BranchSwitcher } from './BranchSwitcher'

afterEach(cleanup)

test('opens a searchable branch menu with the current branch selected', () => {
  render(
    <MemoryRouter>
      <BranchSwitcher
        branches={[
          { id: 'br_main', name: 'main' },
          { id: 'br_release', name: 'release' },
        ]}
        currentBranch={{ id: 'br_main', name: 'main' }}
        search=""
        onSearchChange={() => {}}
        onLoadMore={() => {}}
        getBranchHref={(branchId) => `/branches/${branchId}/activity`}
      />
    </MemoryRouter>
  )

  fireEvent.click(screen.getByRole('button', { name: 'main' }))

  expect(screen.getByRole('searchbox', { name: 'Search branches' })).toBeTruthy()
  expect(screen.getByRole('menuitemcheckbox', { name: 'main' })).toHaveAttribute(
    'aria-checked',
    'true'
  )
  expect(
    screen.getByRole('menuitemcheckbox', { name: 'release' })
  ).toHaveAttribute('href', '/branches/br_release/activity')
})
