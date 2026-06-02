import type { TEnvAccentColor, TTheme } from '@/types'

/**
 * Map an env accent color to the existing TTheme tokens so Badge / Text /
 * etc. components can use them directly without bespoke styling.
 */
export const ENV_ACCENT_THEME: Record<TEnvAccentColor, TTheme> = {
  error: 'error',
  warn: 'warn',
  success: 'success',
  info: 'info',
  brand: 'brand',
  neutral: 'neutral',
}

/**
 * Background-color class for each accent. Used by the page accent stripe
 * — a solid bar is much more visible than a hairline border and can't be
 * clobbered by other CSS rules on the surrounding layout.
 */
export const ENV_ACCENT_BG: Record<TEnvAccentColor, string> = {
  error: 'bg-red-500',
  warn: 'bg-orange-500',
  success: 'bg-green-500',
  info: 'bg-blue-500',
  brand: 'bg-primary-500',
  neutral: 'bg-cool-grey-400',
}
