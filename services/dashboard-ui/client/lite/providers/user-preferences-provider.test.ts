import { afterEach, beforeEach, expect, test } from 'bun:test'
import {
  DEFAULT_USER_PREFERENCES,
  LEGACY_TABLE_VIEW_STORAGE_KEY,
  LEGACY_THEME_STORAGE_KEY,
  USER_PREFERENCES_STORAGE_KEY,
  createUserPreferencesStore,
} from './user-preferences-provider'

beforeEach(() => {
  localStorage.clear()
  sessionStorage.clear()
})

afterEach(() => {
  localStorage.clear()
  sessionStorage.clear()
})

test('writes and returns the complete default record', () => {
  const store = createUserPreferencesStore()

  expect(store.getSnapshot()).toEqual(DEFAULT_USER_PREFERENCES)
  expect(
    JSON.parse(localStorage.getItem(USER_PREFERENCES_STORAGE_KEY) ?? '')
  ).toEqual({
    version: 1,
    preferences: DEFAULT_USER_PREFERENCES,
  })
})

test('recovers invalid fields independently', () => {
  localStorage.setItem(
    USER_PREFERENCES_STORAGE_KEY,
    JSON.stringify({
      version: 1,
      preferences: {
        theme: 'dark',
        collectionView: 'rows',
        diffWrap: true,
        planSectionsOpen: 'yes',
        diffView: 'split',
      },
    })
  )

  expect(createUserPreferencesStore().getSnapshot()).toEqual({
    ...DEFAULT_USER_PREFERENCES,
    theme: 'dark',
    diffWrap: true,
    diffView: 'split',
  })
})

test('migrates valid legacy theme and table preferences', () => {
  localStorage.setItem(LEGACY_THEME_STORAGE_KEY, 'high-contrast')
  sessionStorage.setItem(LEGACY_TABLE_VIEW_STORAGE_KEY, 'cards')

  const store = createUserPreferencesStore()

  expect(store.getSnapshot()).toEqual({
    ...DEFAULT_USER_PREFERENCES,
    theme: 'high-contrast',
    collectionView: 'cards',
  })
  expect(localStorage.getItem(LEGACY_THEME_STORAGE_KEY)).toBeNull()
  expect(sessionStorage.getItem(LEGACY_TABLE_VIEW_STORAGE_KEY)).toBeNull()
})

test('notifies same-tab subscribers when one preference changes', () => {
  const store = createUserPreferencesStore()
  let notifications = 0
  const unsubscribe = store.subscribe(() => {
    notifications += 1
  })

  store.setPreference('diffView', 'split')

  expect(store.getSnapshot().diffView).toBe('split')
  expect(notifications).toBe(1)
  expect(
    JSON.parse(localStorage.getItem(USER_PREFERENCES_STORAGE_KEY) ?? '')
      .preferences
  ).toEqual({
    ...DEFAULT_USER_PREFERENCES,
    diffView: 'split',
  })

  unsubscribe()
})

test('synchronizes a changed record from another tab', () => {
  const store = createUserPreferencesStore()
  const preferences = {
    ...DEFAULT_USER_PREFERENCES,
    collectionView: 'cards' as const,
  }
  localStorage.setItem(
    USER_PREFERENCES_STORAGE_KEY,
    JSON.stringify({ version: 1, preferences })
  )

  const unsubscribe = store.subscribe(() => {})
  window.dispatchEvent(
    new StorageEvent('storage', {
      key: USER_PREFERENCES_STORAGE_KEY,
      newValue: localStorage.getItem(USER_PREFERENCES_STORAGE_KEY),
    })
  )

  expect(store.getSnapshot()).toEqual(preferences)
  unsubscribe()
})

test('reset restores every default', () => {
  const store = createUserPreferencesStore()
  store.setPreference('theme', 'dark')
  store.setPreference('collectionView', 'cards')
  store.setPreference('diffWrap', true)
  store.setPreference('planSectionsOpen', true)
  store.setPreference('diffView', 'split')

  store.resetPreferences()

  expect(store.getSnapshot()).toEqual(DEFAULT_USER_PREFERENCES)
})
