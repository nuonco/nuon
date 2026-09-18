import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Tooltip } from '@/components/common/Tooltip'
import type { TThemePreference } from '@/providers/theme-provider'
import { cn } from '@/utils/classnames'

export interface IThemeSwitcher {
  preference: TThemePreference
  onSelect: (preference: TThemePreference) => void
}

const OPTIONS: Array<{
  preference: TThemePreference
  label: string
  icon: TIconVariant
}> = [
  { preference: 'system', label: 'System', icon: 'DesktopIcon' },
  { preference: 'light', label: 'Light', icon: 'SunIcon' },
  { preference: 'dark', label: 'Dark', icon: 'MoonIcon' },
  { preference: 'classic', label: 'Classic', icon: 'PaletteIcon' },
  { preference: 'high-contrast', label: 'High contrast', icon: 'CircleHalfIcon' },
]

export const ThemeSwitcher = ({ preference, onSelect }: IThemeSwitcher) => {
  return (
    <div
      role="radiogroup"
      aria-label="Theme"
      className="flex items-center gap-0.5 rounded-md border p-0.5"
    >
      {OPTIONS.map((option) => {
        const isSelected = preference === option.preference
        return (
          <Tooltip
            key={option.preference}
            className="flex-auto"
            tipContent={option.label}
            position="bottom"
          >
            <button
              type="button"
              role="radio"
              aria-checked={isSelected}
              aria-label={option.label}
              onClick={() => onSelect(option.preference)}
              className={cn(
                'flex w-full cursor-pointer items-center justify-center rounded-sm p-1.5',
                'transition-colors duration-fastest ease-cubic',
                isSelected
                  ? 'bg-cool-grey-500/16 text-cool-grey-900 dark:text-white'
                  : 'text-cool-grey-600 hover:bg-cool-grey-500/8 dark:text-white/70'
              )}
            >
              <Icon variant={option.icon} size={16} />
            </button>
          </Tooltip>
        )
      })}
    </div>
  )
}
