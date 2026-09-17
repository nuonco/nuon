import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { SwitcherMenu } from './SwitcherMenu'

afterEach(cleanup)

const renderMenu = ({
  hasMore = false,
  onLoadMore = () => {},
}: {
  hasMore?: boolean
  onLoadMore?: () => void
} = {}) =>
  render(
    <MemoryRouter>
      <SwitcherMenu
        items={[
          { id: 'br_main', label: 'main', href: '/branches/br_main' },
          { id: 'br_release', label: 'release', href: '/branches/br_release' },
        ]}
        selectedId="br_main"
        search=""
        onSearchChange={() => {}}
        onLoadMore={onLoadMore}
        searchLabel="Search branches"
        searchPlaceholder="Search branches..."
        emptyTitle="No branches found"
        errorTitle="Branches failed to load"
        hasMore={hasMore}
      />
    </MemoryRouter>
  )

test('renders searchable selectable resources', () => {
  renderMenu()

  expect(screen.getByRole('searchbox', { name: 'Search branches' })).toBeTruthy()
  expect(screen.getByRole('menuitemcheckbox', { name: 'main' })).toHaveAttribute(
    'aria-checked',
    'true'
  )
  expect(screen.queryByRole('menuitem', { name: 'Load more' })).toBeNull()
})

test('shows load more only when another page exists', () => {
  let loads = 0
  renderMenu({ hasMore: true, onLoadMore: () => loads++ })

  fireEvent.click(screen.getByRole('menuitem', { name: 'Load more' }))
  expect(loads).toBe(1)
})

test('moves from search to the first result with ArrowDown', () => {
  renderMenu()
  const search = screen.getByRole('searchbox', { name: 'Search branches' })
  search.focus()

  fireEvent.keyDown(search, { key: 'ArrowDown' })

  expect(screen.getByRole('menuitemcheckbox', { name: 'main' })).toHaveFocus()
})
