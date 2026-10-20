import type { ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { cn } from '@/utils/classnames'

interface IStepActions {
  onGoBack?: () => void
  onAdvance: () => void
  label: string
  disabled?: boolean
  disabledReason?: string
  showCaret?: boolean
  variant?: 'primary' | 'danger'
  cancelHref?: string
  secondary?: ReactNode
}

export const StepActions = ({
  onGoBack,
  onAdvance,
  label,
  disabled,
  disabledReason,
  showCaret = false,
  variant = 'primary',
  cancelHref,
  secondary,
}: IStepActions) => (
  <div className={cn('flex flex-wrap gap-3', onGoBack ? 'justify-between' : 'justify-end')}>
    {onGoBack ? (
      <Button type="button" variant="secondary" onClick={onGoBack}>
        <Icon variant="CaretLeftIcon" weight="bold" /> Back
      </Button>
    ) : null}
    <div className="flex flex-wrap items-center gap-3">
      {cancelHref ? (
        <Button href={cancelHref}>Cancel</Button>
      ) : null}
      {secondary}
      <Button
        type="button"
        variant={variant}
        disabled={disabled}
        onClick={onAdvance}
        tooltipProps={disabled && disabledReason ? { tipContent: disabledReason } : undefined}
      >
        {label}
        {showCaret ? <Icon variant="CaretRightIcon" weight="bold" /> : null}
      </Button>
    </div>
  </div>
)
