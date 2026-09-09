import {
  createContext,
  useCallback,
  useContext,
  useState,
  useSyncExternalStore,
  type ReactNode,
} from 'react'

export type TThemePreference = 'light' | 'dark' | 'high-contrast' | 'system'
export type TCollectionView = 'table' | 'cards'
export type TDiffPreferenceView = 'unified' | 'split'

export interface IUserPreferences {
  theme: TThemePreference
  collectionView: TCollectionView
  diffWrap: boolean
  planSectionsOpen: boolean
  diffView: TDiffPreferenceView
}

interface IUserPreferencesRecord {
  version: 1
  preferences: IUserPreferences
}

interface IUserPreferencesStore {
  getSnapshot: () => IUserPreferences
  subscribe: (listener: () => void) => () => void
  setPreference: <K extends keyof IUserPreferences>(
    key: K,
    value: IUserPreferences[K]
  ) => void
  resetPreferences: () => void
}

export const USER_PREFERENCES_STORAGE_KEY = 'nuon-lite-preferences'
export const LEGACY_THEME_STORAGE_KEY = 'nuon-lite-theme'
export const LEGACY_TABLE_VIEW_STORAGE_KEY = 'nuon-lite-table-view'

export const DEFAULT_USER_PREFERENCES: IUserPreferences = Object.freeze({
  theme: 'system',
  collectionView: 'table',
  diffWrap: false,
  planSectionsOpen: false,
  diffView: 'unified',
})

const themePreferences: TThemePreference[] = [
  'light',
  'dark',
  'high-contrast',
  'system',
]
const collectionViews: TCollectionView[] = ['table', 'cards']
const diffViews: TDiffPreferenceView[] = ['unified', 'split']

const isThemePreference = (value: unknown): value is TThemePreference =>
  themePreferences.includes(value as TThemePreference)

const isCollectionView = (value: unknown): value is TCollectionView =>
  collectionViews.includes(value as TCollectionView)

const isDiffView = (value: unknown): value is TDiffPreferenceView =>
  diffViews.includes(value as TDiffPreferenceView)

const validatedPreferences = (value: unknown): IUserPreferences => {
  const candidate =
    value && typeof value === 'object'
      ? (value as Partial<IUserPreferences>)
      : {}

  return {
    theme: isThemePreference(candidate.theme)
      ? candidate.theme
      : DEFAULT_USER_PREFERENCES.theme,
    collectionView: isCollectionView(candidate.collectionView)
      ? candidate.collectionView
      : DEFAULT_USER_PREFERENCES.collectionView,
    diffWrap:
      typeof candidate.diffWrap === 'boolean'
        ? candidate.diffWrap
        : DEFAULT_USER_PREFERENCES.diffWrap,
    planSectionsOpen:
      typeof candidate.planSectionsOpen === 'boolean'
        ? candidate.planSectionsOpen
        : DEFAULT_USER_PREFERENCES.planSectionsOpen,
    diffView: isDiffView(candidate.diffView)
      ? candidate.diffView
      : DEFAULT_USER_PREFERENCES.diffView,
  }
}

const readRecord = (): IUserPreferences | null => {
  if (typeof window === 'undefined') return null

  try {
    const raw = window.localStorage.getItem(USER_PREFERENCES_STORAGE_KEY)
    if (!raw) return null
    const record = JSON.parse(raw) as Partial<IUserPreferencesRecord>
    if (record.version !== 1) return null
    return validatedPreferences(record.preferences)
  } catch {
    return null
  }
}

const writeRecord = (preferences: IUserPreferences) => {
  if (typeof window === 'undefined') return

  try {
    const record: IUserPreferencesRecord = { version: 1, preferences }
    window.localStorage.setItem(
      USER_PREFERENCES_STORAGE_KEY,
      JSON.stringify(record)
    )
  } catch {}
}

const migrateLegacyPreferences = (): IUserPreferences => {
  const stored = readRecord()
  if (stored) {
    writeRecord(stored)
    return stored
  }
  if (typeof window === 'undefined') return DEFAULT_USER_PREFERENCES

  let theme = DEFAULT_USER_PREFERENCES.theme
  let collectionView = DEFAULT_USER_PREFERENCES.collectionView

  try {
    const legacyTheme = window.localStorage.getItem(LEGACY_THEME_STORAGE_KEY)
    if (isThemePreference(legacyTheme)) theme = legacyTheme
  } catch {}

  try {
    const legacyView = window.sessionStorage.getItem(
      LEGACY_TABLE_VIEW_STORAGE_KEY
    )
    if (isCollectionView(legacyView)) collectionView = legacyView
  } catch {}

  const preferences = {
    ...DEFAULT_USER_PREFERENCES,
    theme,
    collectionView,
  }
  writeRecord(preferences)

  try {
    window.localStorage.removeItem(LEGACY_THEME_STORAGE_KEY)
  } catch {}
  try {
    window.sessionStorage.removeItem(LEGACY_TABLE_VIEW_STORAGE_KEY)
  } catch {}

  return preferences
}

export const createUserPreferencesStore = (): IUserPreferencesStore => {
  let snapshot = migrateLegacyPreferences()
  const listeners = new Set<() => void>()

  const emit = () => listeners.forEach((listener) => listener())

  const setSnapshot = (preferences: IUserPreferences) => {
    snapshot = preferences
    writeRecord(preferences)
    emit()
  }

  return {
    getSnapshot: () => snapshot,
    subscribe: (listener) => {
      const handleStorage = (event: StorageEvent) => {
        if (event.key !== USER_PREFERENCES_STORAGE_KEY) return
        const next = readRecord() ?? DEFAULT_USER_PREFERENCES
        if (JSON.stringify(next) === JSON.stringify(snapshot)) return
        snapshot = next
        emit()
      }

      listeners.add(listener)
      window.addEventListener('storage', handleStorage)
      return () => {
        listeners.delete(listener)
        window.removeEventListener('storage', handleStorage)
      }
    },
    setPreference: (key, value) => setSnapshot({ ...snapshot, [key]: value }),
    resetPreferences: () => setSnapshot({ ...DEFAULT_USER_PREFERENCES }),
  }
}

const UserPreferencesContext = createContext<IUserPreferencesStore | null>(null)

export const UserPreferencesProvider = ({
  children,
}: {
  children: ReactNode
}) => {
  const [store] = useState(createUserPreferencesStore)

  return (
    <UserPreferencesContext.Provider value={store}>
      {children}
    </UserPreferencesContext.Provider>
  )
}

export const useUserPreferences = () => {
  const store = useContext(UserPreferencesContext)
  if (!store) {
    throw new Error(
      'useUserPreferences must be used within a UserPreferencesProvider'
    )
  }

  const preferences = useSyncExternalStore(
    store.subscribe,
    store.getSnapshot,
    () => DEFAULT_USER_PREFERENCES
  )
  const setPreference = useCallback(
    <K extends keyof IUserPreferences>(key: K, value: IUserPreferences[K]) =>
      store.setPreference(key, value),
    [store]
  )
  const resetPreferences = useCallback(() => store.resetPreferences(), [store])

  return { preferences, setPreference, resetPreferences }
}
