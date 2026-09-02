import {
  ArrowClockwiseIcon,
  ArrowUpIcon,
  ArrowElbowDownLeftIcon,
  ArrowsHorizontalIcon,
  ArrowsInLineVerticalIcon,
  ArrowsOutLineVerticalIcon,
  ArrowSquareOutIcon,
  CaretDownIcon,
  CaretRightIcon,
  CaretUpIcon,
  CheckCircleIcon,
  CheckIcon,
  CircleHalfIcon,
  CopyIcon,
  ClockCountdownIcon,
  DesktopIcon,
  DotsThreeIcon,
  FunnelIcon,
  MagnifyingGlassIcon,
  MoonIcon,
  PlusIcon,
  SlidersHorizontalIcon,
  SparkleIcon,
  SquareSplitHorizontalIcon,
  SquareSplitVerticalIcon,
  SunIcon,
  TrashIcon,
  WarningIcon,
  XCircleIcon,
  XIcon,
  type IconProps as PhosphorIconProps,
} from '@phosphor-icons/react'

const ICONS = {
  ArrowClockwiseIcon,
  ArrowsHorizontalIcon,
  ArrowsInLineVerticalIcon,
  ArrowsOutLineVerticalIcon,
  ArrowElbowDownLeftIcon,
  ArrowUpIcon,
  ArrowSquareOutIcon,
  CaretDownIcon,
  CaretRightIcon,
  CaretUpIcon,
  CheckCircleIcon,
  CheckIcon,
  CircleHalfIcon,
  CopyIcon,
  ClockCountdownIcon,
  DesktopIcon,
  DotsThreeIcon,
  FunnelIcon,
  MagnifyingGlassIcon,
  MoonIcon,
  PlusIcon,
  SlidersHorizontalIcon,
  SparkleIcon,
  SquareSplitHorizontalIcon,
  SquareSplitVerticalIcon,
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
