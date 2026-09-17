import { createContext, useCallback, useEffect, useMemo, useState } from 'react'
import {
  useUserPreferences,
  type TThemePreference,
} from './user-preferences-provider'

export type { TThemePreference } from './user-preferences-provider'
export type TTheme = 'light' | 'dark' | 'high-contrast'

interface IThemeContext {
  preference: TThemePreference
  theme: TTheme
  setPreference: (preference: TThemePreference) => void
}

export const ThemeContext = createContext<IThemeContext | undefined>(undefined)

const darkQuery = () => window.matchMedia('(prefers-color-scheme: dark)')

export const ThemeProvider = ({ children }: { children: React.ReactNode }) => {
  const { preferences, setPreference: setUserPreference } = useUserPreferences()
  const preference = preferences.theme
  const [systemTheme, setSystemTheme] = useState<TTheme>(() =>
    darkQuery().matches ? 'dark' : 'light'
  )

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

  const setPreference = useCallback(
    (next: TThemePreference) => {
      setUserPreference('theme', next)
    },
    [setUserPreference]
  )

  const theme = preference === 'system' ? systemTheme : preference

  const value = useMemo(
    () => ({ preference, theme, setPreference }),
    [preference, theme, setPreference]
  )

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}
