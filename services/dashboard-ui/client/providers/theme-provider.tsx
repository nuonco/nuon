import {
  createContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { useStoredViewMode } from '@/hooks/use-stored-view-mode'

export type TThemePreference =
  | 'system'
  | 'light'
  | 'dark'
  | 'classic'
  | 'high-contrast'
export type TResolvedTheme = Exclude<TThemePreference, 'system'>
export type TColorScheme = 'light' | 'dark'

export const THEME_STORAGE_KEY = 'nuon-theme'

export const THEME_PREFERENCES: readonly TThemePreference[] = [
  'system',
  'light',
  'dark',
  'classic',
  'high-contrast',
]

const DARK_THEMES: readonly TResolvedTheme[] = [
  'dark',
  'classic',
  'high-contrast',
]

export interface IThemeContext {
  preference: TThemePreference
  theme: TResolvedTheme
  colorScheme: TColorScheme
  setPreference: (preference: TThemePreference) => void
}

export const ThemeContext = createContext<IThemeContext | undefined>(undefined)

const darkQuery = () => window.matchMedia('(prefers-color-scheme: dark)')

export const systemColorScheme = (): TColorScheme => {
  if (typeof window === 'undefined') return 'light'
  return darkQuery().matches ? 'dark' : 'light'
}

export const ThemeProvider = ({ children }: { children: ReactNode }) => {
  const [preference, setStoredPreference] = useStoredViewMode<TThemePreference>(
    THEME_STORAGE_KEY,
    THEME_PREFERENCES,
    'system'
  )
  const [systemTheme, setSystemTheme] = useState<TColorScheme>(systemColorScheme)

  useEffect(() => {
    const query = darkQuery()
    const onChange = (event: MediaQueryListEvent) =>
      setSystemTheme(event.matches ? 'dark' : 'light')
    query.addEventListener('change', onChange)
    return () => query.removeEventListener('change', onChange)
  }, [])

  useEffect(() => {
    const root = document.documentElement
    if (preference === 'system') {
      root.removeAttribute('data-theme')
    } else {
      root.setAttribute('data-theme', preference)
    }
  }, [preference])

  const theme = preference === 'system' ? systemTheme : preference
  const colorScheme: TColorScheme = DARK_THEMES.includes(theme)
    ? 'dark'
    : 'light'

  const value = useMemo(
    () => ({
      preference,
      theme,
      colorScheme,
      setPreference: setStoredPreference,
    }),
    [preference, theme, colorScheme, setStoredPreference]
  )

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}
