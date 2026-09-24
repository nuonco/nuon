import { useCopy } from '../../hooks/use-copy'
import type { TPopoverSide } from '../../hooks/use-popover'
import { Button, type IButton } from './Button'
import { Icon } from './Icon'

export interface ICopyButton
  extends Omit<IButton, 'children' | 'iconOnly' | 'tooltip' | 'value'> {
  value: string
  label?: string
  side?: TPopoverSide
}

const MESSAGE = {
  idle: (label: string) => label,
  copied: () => 'Copied',
  error: () => 'Press ⌘C to copy',
}

export const CopyButton = ({
  value,
  label = 'Copy',
  side = 'top',
  variant = 'ghost',
  size = 'sm',
  onClick,
  ...props
}: ICopyButton) => {
  const { copy, status } = useCopy()

  return (
    <Button
      variant={variant}
      size={size}
      iconOnly
      aria-label={label}
      tooltip={MESSAGE[status](label)}
      tooltipSide={side}
      onClick={(event) => {
        onClick?.(event)
        void copy(value)
      }}
      {...props}
    >
      <Icon
        variant={status === 'copied' ? 'CheckIcon' : 'CopyIcon'}
        size={size === 'sm' ? 14 : 16}
      />
    </Button>
  )
}
