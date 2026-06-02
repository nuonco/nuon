import type { TEnvAccentColor, TInstall, TOrg } from '@/types'

export type TEnvAccent = {
  /** The label value that matched (e.g., "production"). */
  value: string
  /** The accent color (maps 1:1 to TTheme tokens). */
  color: TEnvAccentColor
  /** The label key used (e.g., "env"). */
  labelKey: string
}

/**
 * Resolve the environment accent for an install given the org's config.
 *
 * Returns null when:
 *   - the org has no config
 *   - the config has no label_key
 *   - the install doesn't have that label
 *   - the install's label value isn't in the configured map
 *
 * This means installs that aren't tagged with a known env stay visually
 * untouched (no accent), so the indicator never fires by accident.
 */
export function resolveEnvAccent(
  install?: Pick<TInstall, 'labels'> | null,
  org?: Pick<TOrg, 'env_accent_config'> | null,
): TEnvAccent | null {
  const cfg = org?.env_accent_config
  const labelKey = cfg?.label_key
  if (!cfg || !labelKey || !cfg.values) return null

  const labels = install?.labels
  if (!labels) return null

  const value = labels[labelKey]
  if (!value) return null

  const color = cfg.values[value]
  if (!color) return null

  return { value, color, labelKey }
}
