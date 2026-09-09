import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import {
  DEFAULT_USER_PREFERENCES,
  USER_PREFERENCES_STORAGE_KEY,
  UserPreferencesProvider,
} from '../../providers/user-preferences-provider'
import { DiffSection } from './DiffSection'
import { DiffSections } from './DiffSections'

const section = (
  <DiffSection title="Example resource" operation="update" before="" after="" />
)

const storedPreferences = () =>
  JSON.parse(localStorage.getItem(USER_PREFERENCES_STORAGE_KEY) ?? '')
    .preferences

afterEach(() => {
  cleanup()
  localStorage.clear()
  sessionStorage.clear()
})

test('uses saved defaults without persisting page controls', () => {
  localStorage.setItem(
    USER_PREFERENCES_STORAGE_KEY,
    JSON.stringify({
      version: 1,
      preferences: {
        ...DEFAULT_USER_PREFERENCES,
        planSectionsOpen: true,
        diffView: 'split',
      },
    })
  )

  render(
    <UserPreferencesProvider>
      <DiffSections>{section}</DiffSections>
    </UserPreferencesProvider>
  )

  expect(
    screen.getByRole('button', { name: /^Example resource/ })
  ).toHaveAttribute('aria-expanded', 'true')
  fireEvent.click(screen.getByRole('button', { name: 'Unified view' }))
  expect(storedPreferences().diffView).toBe('split')

  fireEvent.click(screen.getByRole('button', { name: /^Example resource/ }))
  expect(storedPreferences().planSectionsOpen).toBe(true)

  cleanup()
  render(
    <UserPreferencesProvider>
      <DiffSections>{section}</DiffSections>
    </UserPreferencesProvider>
  )
  expect(screen.getByRole('button', { name: 'Unified view' })).toBeTruthy()
})

test('keeps explicit defaults local to the mounted diff', () => {
  render(
    <UserPreferencesProvider>
      <DiffSections defaultOpen defaultView="split">
        {section}
      </DiffSections>
    </UserPreferencesProvider>
  )

  expect(
    screen.getByRole('button', { name: /^Example resource/ })
  ).toHaveAttribute('aria-expanded', 'true')
  fireEvent.click(screen.getByRole('button', { name: 'Unified view' }))

  expect(storedPreferences()).toEqual(DEFAULT_USER_PREFERENCES)
})
