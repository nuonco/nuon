import { useTheme } from '@/hooks/use-theme'
import { ThemeSwitcher } from './ThemeSwitcher'

export const ThemeSwitcherContainer = () => {
  const { preference, setPreference } = useTheme()

  return <ThemeSwitcher preference={preference} onSelect={setPreference} />
}
