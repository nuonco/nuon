import type { CSSProperties, ReactNode } from 'react'
import type { TToastTheme } from '../../../providers/toast-provider'
import {
  STATUS_THEME_COLOR,
  STATUS_THEME_ICON,
} from '../../../utils/status-theme'
import { Button } from '../../atoms/Button'
import { Card } from '../../atoms/Card'
import { Icon } from '../../atoms/Icon'
import { Text } from '../../atoms/Text'

export interface IToast {
  heading: string
  description?: ReactNode
  theme?: TToastTheme
  actionLabel?: string
  onAction?: () => void
  onDismiss: () => void
}

export const Toast = ({
  heading,
  description,
  theme = 'default',
  actionLabel,
  onAction,
  onDismiss,
}: IToast) => {
  const urgent = theme === 'error' || theme === 'warn'
  const style = {
    '--toast-status': STATUS_THEME_COLOR[theme],
    '--card-shadow-floating': 'var(--toast-shadow)',
  } as CSSProperties

  return (
    <Card
      role={urgent ? 'alert' : 'status'}
      data-toast-theme={theme}
      aria-live={urgent ? 'assertive' : 'polite'}
      aria-atomic="true"
      padding="none"
      blur="lg"
      opacity="strong"
      shadow="floating"
      style={style}
      className="relative flex min-h-20 w-full overflow-hidden"
    >
      <span
        aria-hidden
        className="absolute inset-y-0 left-0 w-1.5 bg-[var(--toast-status)]"
      />
      <span
        aria-hidden
        className="flex shrink-0 pl-4 pt-4 text-[var(--toast-status)]"
      >
        <Icon variant={STATUS_THEME_ICON[theme]} size={20} />
      </span>
      <div className="flex min-w-0 flex-1 flex-col gap-1 px-3 py-4">
        <Text as="p" weight="medium">
          {heading}
        </Text>
        {description ? (
          typeof description === 'string' ? (
            <Text as="p" variant="caption" color="secondary">
              {description}
            </Text>
          ) : (
            <div className="text-caption text-secondary">{description}</div>
          )
        ) : null}
        {actionLabel && onAction ? (
          <div className="pt-2">
            <Button size="sm" variant="secondary" onClick={onAction}>
              {actionLabel}
            </Button>
          </div>
        ) : null}
      </div>
      <div className="shrink-0 pr-2 pt-2">
        <Button
          size="sm"
          variant="ghost"
          iconOnly
          aria-label={`Dismiss ${heading} notification`}
          onClick={onDismiss}
        >
          <Icon variant="XIcon" size={16} />
        </Button>
      </div>
    </Card>
  )
}
