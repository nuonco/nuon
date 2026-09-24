import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { UserPreferencesProvider } from '../../providers/user-preferences-provider'
import { SurfacesProvider } from '../../providers/surfaces-provider'
import { OrgSwitcherMenu } from './OrgSwitcherMenu/OrgSwitcherMenu'
import { UserDropdown } from './UserDropdown'
import { UserDropdownContainer } from './UserDropdownContainer'
import { SurfaceHost } from './surfaces'

afterEach(() => {
  cleanup()
  localStorage.clear()
  sessionStorage.clear()
})

test('opens a sign-out-only account menu', () => {
  render(
    <UserDropdown
      user={{ name: 'Alex Morgan', email: 'alex@example.com' }}
      signOutHref="https://auth.example.com/logout"
    />
  )

  fireEvent.click(screen.getByRole('button', { name: /Alex Morgan/ }))

  const signOut = screen.getByRole('menuitem', { name: 'Sign out' })
  expect(signOut.getAttribute('href')).toBe('https://auth.example.com/logout')
  expect(signOut.getAttribute('target')).toBe('_self')
  expect(screen.getAllByRole('menuitem')).toHaveLength(1)
})

test('labels the compact trigger as a user menu', () => {
  render(
    <UserDropdown
      user={{ name: 'Alex Morgan', email: 'alex@example.com' }}
      signOutHref="https://auth.example.com/logout"
      compact
    />
  )

  expect(screen.getByRole('button', { name: 'Open user menu' })).toBeTruthy()
})

test('opens the organization switcher as a nested menu', () => {
  render(
    <MemoryRouter>
      <UserDropdown
        user={{ name: 'Alex Morgan', email: 'alex@example.com' }}
        signOutHref="https://auth.example.com/logout"
        org={{ id: 'org_alpha', name: 'alpha', status: 'active' }}
        orgSwitcher={
          <OrgSwitcherMenu
            orgs={[
              { id: 'org_alpha', name: 'alpha' },
              { id: 'org_beta', name: 'beta' },
            ]}
            currentOrgId="org_alpha"
            search=""
            onSearchChange={() => {}}
            onLoadMore={() => {}}
          />
        }
      />
    </MemoryRouter>
  )

  fireEvent.click(screen.getByRole('button', { name: /Alex Morgan/ }))
  fireEvent.click(screen.getByRole('menuitem', { name: /alpha/ }))

  expect(
    screen.getByRole('searchbox', { name: 'Search organizations' })
  ).toBeTruthy()
  expect(
    screen.getByRole('menuitemcheckbox', { name: /beta/ })
  ).toHaveAttribute('href', '/org_beta')
})

test('opens preferences and saves changes from the user menu', () => {
  render(
    <MemoryRouter>
      <UserPreferencesProvider>
        <SurfacesProvider>
          <SurfaceHost scope="test">
            <UserDropdownContainer
              user={{ name: 'Alex Morgan', email: 'alex@example.com' }}
              signOutHref="https://auth.example.com/logout"
            />
          </SurfaceHost>
        </SurfacesProvider>
      </UserPreferencesProvider>
    </MemoryRouter>
  )

  fireEvent.click(screen.getByRole('button', { name: /Alex Morgan/ }))
  fireEvent.click(screen.getByRole('menuitem', { name: 'Preferences' }))

  expect(screen.getByRole('dialog', { name: 'Preferences' })).toBeTruthy()
  fireEvent.click(screen.getByRole('radio', { name: /Cards/ }))

  expect(
    JSON.parse(localStorage.getItem('nuon-lite-preferences') ?? '').preferences
      .collectionView
  ).toBe('cards')

  fireEvent.click(screen.getByRole('button', { name: 'Reset preferences' }))

  expect(screen.getByRole('radio', { name: /Table/ })).toBeChecked()
})
