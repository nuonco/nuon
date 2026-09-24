import {
  ArrowClockwiseIcon,
  ArrowDownIcon,
  ArrowUpIcon,
  ArrowElbowDownLeftIcon,
  ArrowLineRightIcon,
  ArrowsHorizontalIcon,
  ArrowsInLineVerticalIcon,
  ArrowsOutLineVerticalIcon,
  ArrowSquareOutIcon,
  AppWindowIcon,
  BookOpenTextIcon,
  CaretDownIcon,
  CaretLeftIcon,
  CaretRightIcon,
  CaretUpIcon,
  CardsIcon,
  CheckCircleIcon,
  CheckIcon,
  CircleHalfIcon,
  CopyIcon,
  CornersInIcon,
  CornersOutIcon,
  ClockCountdownIcon,
  DesktopIcon,
  DotsThreeIcon,
  GearIcon,
  FileDashedIcon,
  FileIcon,
  FunnelIcon,
  GitBranchIcon,
  GitCommitIcon,
  HeartbeatIcon,
  HouseIcon,
  InfoIcon,
  KeyIcon,
  LightningIcon,
  MagnifyingGlassIcon,
  MinusCircleIcon,
  MinusIcon,
  MoonIcon,
  PlugsConnectedIcon,
  PlusIcon,
  PuzzlePieceIcon,
  ProhibitIcon,
  QuestionIcon,
  RepeatIcon,
  RobotIcon,
  ShieldCheckIcon,
  ShippingContainerIcon,
  SlidersHorizontalIcon,
  SneakerMoveIcon,
  SparkleIcon,
  SignOutIcon,
  SidebarSimpleIcon,
  SquaresFourIcon,
  SquareSplitHorizontalIcon,
  SquareSplitVerticalIcon,
  SunIcon,
  TableIcon,
  TrashIcon,
  CubeIcon,
  UserIcon,
  UsersThreeIcon,
  WarningIcon,
  WebhooksLogoIcon,
  XCircleIcon,
  XIcon,
  type IconProps as PhosphorIconProps,
} from '@phosphor-icons/react'
import { FaGithub } from 'react-icons/fa'

const ICONS = {
  ArrowClockwiseIcon,
  ArrowsHorizontalIcon,
  ArrowsInLineVerticalIcon,
  ArrowsOutLineVerticalIcon,
  ArrowElbowDownLeftIcon,
  ArrowLineRightIcon,
  ArrowDownIcon,
  ArrowUpIcon,
  ArrowSquareOutIcon,
  AppWindowIcon,
  BookOpenTextIcon,
  CaretDownIcon,
  CaretLeftIcon,
  CaretRightIcon,
  CaretUpIcon,
  CardsIcon,
  CheckCircleIcon,
  CheckIcon,
  CircleHalfIcon,
  CopyIcon,
  CornersInIcon,
  CornersOutIcon,
  ClockCountdownIcon,
  DesktopIcon,
  DotsThreeIcon,
  GearIcon,
  FileDashedIcon,
  FileIcon,
  FunnelIcon,
  GitBranchIcon,
  GitCommitIcon,
  HeartbeatIcon,
  HouseIcon,
  InfoIcon,
  KeyIcon,
  LightningIcon,
  MagnifyingGlassIcon,
  MinusCircleIcon,
  MinusIcon,
  MoonIcon,
  PlugsConnectedIcon,
  PlusIcon,
  PuzzlePieceIcon,
  ProhibitIcon,
  QuestionIcon,
  RepeatIcon,
  RobotIcon,
  ShieldCheckIcon,
  ShippingContainerIcon,
  SlidersHorizontalIcon,
  SneakerMoveIcon,
  SparkleIcon,
  SignOutIcon,
  SidebarSimpleIcon,
  SquaresFourIcon,
  SquareSplitHorizontalIcon,
  SquareSplitVerticalIcon,
  SunIcon,
  TableIcon,
  TrashIcon,
  CubeIcon,
  UserIcon,
  UsersThreeIcon,
  WarningIcon,
  WebhooksLogoIcon,
  XCircleIcon,
  XIcon,
} as const

const CUSTOM_ICONS = {
  GitHub: FaGithub,
} as const

export type TIconVariant = keyof typeof ICONS | keyof typeof CUSTOM_ICONS

export interface IIcon extends Omit<PhosphorIconProps, 'ref' | 'color'> {
  variant: TIconVariant
}

export const Icon = ({
  variant,
  size = 16,
  weight = 'regular',
  ...props
}: IIcon) => {
  const Custom = CUSTOM_ICONS[variant as keyof typeof CUSTOM_ICONS]
  if (Custom) {
    return <Custom size={size} aria-hidden {...(props as object)} />
  }

  const Component = ICONS[variant as keyof typeof ICONS]

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
