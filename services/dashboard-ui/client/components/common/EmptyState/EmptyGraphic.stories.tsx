export default {
  title: 'UI / Empty state / Empty graphic',
}

import { useEffect, useRef } from 'react'
import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { useTheme } from '@/hooks/use-theme'
import {
  THEME_PREFERENCES,
  type TThemePreference,
} from '@/providers/theme-provider'
import type { TEmptyVariant } from '@/types'
import { EmptyGraphic } from './EmptyGraphic'

const VARIANTS: TEmptyVariant[] = [
  '404',
  'actions',
  'app',
  'diagram',
  'history',
  'policy',
  'search',
  'table',
]

const THEME_LABELS: Record<TThemePreference, string> = {
  system: 'System',
  light: 'Light',
  dark: 'Dark',
  classic: 'Classic',
  'high-contrast': 'High contrast',
  monochrome: 'Monochrome',
}

export const Default = () => <EmptyGraphic />

export const Variants = () => (
  <div className="flex gap-4 items-center">
    <EmptyGraphic variant="404" />
    <EmptyGraphic variant="actions" />
    <EmptyGraphic variant="app" />
    <EmptyGraphic variant="diagram" />
    <EmptyGraphic variant="history" />
    <EmptyGraphic variant="policy" />
    <EmptyGraphic variant="search" />
    <EmptyGraphic variant="table" />
  </div>
)

export const Small = () => (
  <div className="flex gap-4 items-center">
    <EmptyGraphic variant="404" size="sm" />
    <EmptyGraphic variant="actions" size="sm" />
    <EmptyGraphic variant="app" size="sm" />
    <EmptyGraphic variant="diagram" size="sm" />
    <EmptyGraphic variant="history" size="sm" />
    <EmptyGraphic variant="policy" size="sm" />
    <EmptyGraphic variant="search" size="sm" />
    <EmptyGraphic variant="table" size="sm" />
  </div>
)

export const DarkModeOnly = () => <EmptyGraphic isDarkModeOnly />

export const Themes = () => {
  const { preference, setPreference } = useTheme()
  const initialPreference = useRef(preference)
  const setPreferenceRef = useRef(setPreference)
  setPreferenceRef.current = setPreference

  useEffect(() => {
    const initial = initialPreference.current
    return () => setPreferenceRef.current(initial)
  }, [])

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-wrap gap-2" role="group" aria-label="Theme">
        {THEME_PREFERENCES.map((theme) => (
          <Button
            key={theme}
            size="sm"
            variant={preference === theme ? 'primary' : 'secondary'}
            aria-pressed={preference === theme}
            onClick={() => setPreference(theme)}
          >
            {THEME_LABELS[theme]}
          </Button>
        ))}
      </div>
      <div className="grid grid-cols-2 gap-8 sm:grid-cols-4">
        {VARIANTS.map((variant) => (
          <div key={variant} className="flex flex-col items-center gap-3">
            <Text variant="label" theme="neutral">
              {variant}
            </Text>
            <EmptyGraphic variant={variant} />
            <EmptyGraphic variant={variant} size="sm" />
          </div>
        ))}
      </div>
    </div>
  )
}
