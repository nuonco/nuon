import { Badge, type IBadge } from '@/components/common/Badge'
import type { TEnvAccent } from '@/utils/env-accent'
import { ENV_ACCENT_THEME } from '@/utils/env-accent-classes'

interface IEnvAccentBadge extends Omit<IBadge, 'theme' | 'children'> {
  accent: TEnvAccent
  /** Show the matched label value, defaults to true. */
  showLabel?: boolean
}

/**
 * Filled badge used in the status bar and (compactly) inside the install
 * page header.
 */
export const EnvAccentBadge = ({
  accent,
  showLabel = true,
  size = 'sm',
  variant = 'code',
  ...props
}: IEnvAccentBadge) => {
  return (
    <Badge
      size={size}
      variant={variant}
      theme={ENV_ACCENT_THEME[accent.color]}
      title={`${accent.labelKey}: ${accent.value}`}
      {...props}
    >
      <span
        aria-hidden
        className="inline-block w-1.5 h-1.5 rounded-full bg-current opacity-80"
      />
      {showLabel ? accent.value : null}
    </Badge>
  )
}
