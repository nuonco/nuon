import { Icon } from '@/components/common/Icon'
import { cn } from '@/utils/classnames'
import { useTheme } from '../../hooks/use-theme'
import type { TThemePreference } from '../../providers/theme-provider'

const OPTIONS: Array<{
  preference: TThemePreference
  label: string
  icon: 'SunIcon' | 'MoonIcon' | 'CircleHalfIcon' | 'DesktopIcon'
}> = [
  { preference: 'light', label: 'Light', icon: 'SunIcon' },
  { preference: 'dark', label: 'Dark', icon: 'MoonIcon' },
  { preference: 'high-contrast', label: 'High contrast', icon: 'CircleHalfIcon' },
  { preference: 'system', label: 'System', icon: 'DesktopIcon' },
]

export const ThemeSwitcher = () => {
  const { preference, setPreference } = useTheme()

  return (
    <div
      role="radiogroup"
      aria-label="Theme"
      className="inline-flex items-center gap-1 rounded-lg border border-divider bg-surface-02 p-1"
    >
      {OPTIONS.map((option) => {
        const isSelected = preference === option.preference
        return (
          <button
            key={option.preference}
            type="button"
            role="radio"
            aria-checked={isSelected}
            aria-label={option.label}
            onClick={() => setPreference(option.preference)}
            className={cn(
              'flex cursor-pointer items-center gap-1.5 rounded-md px-2.5 py-1.5 text-sm transition-colors',
              isSelected
                ? 'bg-surface-default text-primary shadow-sm'
                : 'text-tertiary hover:text-secondary'
            )}
          >
            <Icon variant={option.icon} size={16} />
            {option.label}
          </button>
        )
      })}
    </div>
  )
}
