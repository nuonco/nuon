import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { UserDropdown } from './UserDropdown'

afterEach(cleanup)

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
