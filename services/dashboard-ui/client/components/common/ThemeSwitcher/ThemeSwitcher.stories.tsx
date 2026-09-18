import { useState } from 'react'
import { ThemeSwitcher } from './ThemeSwitcher'
import type { TThemePreference } from '@/providers/theme-provider'

export default {
  title: 'Common / ThemeSwitcher',
}

const Demo = ({ initial }: { initial: TThemePreference }) => {
  const [preference, setPreference] = useState<TThemePreference>(initial)

  return (
    <div className="flex w-56 flex-col gap-2">
      <ThemeSwitcher preference={preference} onSelect={setPreference} />
    </div>
  )
}

export function System() {
  return <Demo initial="system" />
}

export function Light() {
  return <Demo initial="light" />
}

export function Dark() {
  return <Demo initial="dark" />
}

export function Classic() {
  return <Demo initial="classic" />
}

export function HighContrast() {
  return <Demo initial="high-contrast" />
}
