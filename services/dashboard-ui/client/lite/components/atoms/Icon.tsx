import {
  ArrowClockwiseIcon,
  ArrowSquareOutIcon,
  CaretDownIcon,
  CaretRightIcon,
  CheckCircleIcon,
  CheckIcon,
  CircleHalfIcon,
  CopyIcon,
  ClockCountdownIcon,
  DesktopIcon,
  MoonIcon,
  PlusIcon,
  SparkleIcon,
  SunIcon,
  TrashIcon,
  WarningIcon,
  XCircleIcon,
  XIcon,
  type IconProps as PhosphorIconProps,
} from '@phosphor-icons/react'

const ICONS = {
  ArrowClockwiseIcon,
  ArrowSquareOutIcon,
  CaretDownIcon,
  CaretRightIcon,
  CheckCircleIcon,
  CheckIcon,
  CircleHalfIcon,
  CopyIcon,
  ClockCountdownIcon,
  DesktopIcon,
  MoonIcon,
  PlusIcon,
  SparkleIcon,
  SunIcon,
  TrashIcon,
  WarningIcon,
  XCircleIcon,
  XIcon,
} as const

export type TIconVariant = keyof typeof ICONS

export interface IIcon extends Omit<PhosphorIconProps, 'ref' | 'color'> {
  variant: TIconVariant
}

export const Icon = ({
  variant,
  size = 16,
  weight = 'regular',
  ...props
}: IIcon) => {
  const Component = ICONS[variant]

  if (!Component) {
    if (process.env.NODE_ENV === 'development') {
      console.warn(
        `Icon variant "${variant}" is missing. Import it from @phosphor-icons/react and add it to the ICONS map in client/lite/components/atoms/Icon.tsx.`
      )
    }
    return null
  }

  return <Component size={size} weight={weight} aria-hidden {...props} />
}
