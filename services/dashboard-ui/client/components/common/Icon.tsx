import { ComponentProps, ElementType } from 'react'
import { FaAws, FaGithub } from 'react-icons/fa'
import {
  SiAwslambda,
  SiDocker,
  SiGooglecloud,
  SiHelm,
  SiKubernetes,
  SiOpencontainersinitiative,
  SiPulumi,
  SiTerraform,
} from 'react-icons/si'
import { VscAzure } from 'react-icons/vsc'
import * as PhosphorIcons from '@phosphor-icons/react'
import { Loading } from './Loading'
import { AWSColor, AzureColor, GCPColor } from './CloudPlatformColorIcons'
import type { TTheme } from '@/types'
import { cn } from '@/utils/classnames'

const customIcons = {
  AWS: FaAws,
  AWSColor,
  AWSLambda: SiAwslambda,
  Azure: VscAzure,
  AzureColor,
  Docker: SiDocker,
  GCP: SiGooglecloud,
  GCPColor,
  GitHub: FaGithub,
  Helm: SiHelm,
  Kubernetes: SiKubernetes,
  Loading: Loading,
  OCI: SiOpencontainersinitiative,
  Pulumi: SiPulumi,
  Terraform: SiTerraform,
} as const

type CustomIconVariant = keyof typeof customIcons

type PhosphorIconVariant = keyof Omit<
  typeof PhosphorIcons,
  'Icon' | 'IconContext' | 'IconBase' | 'createComponent' | 'IconProps'
>

export type TIconVariant = PhosphorIconVariant | CustomIconVariant

type PhosphorIconProps = ComponentProps<typeof PhosphorIcons.HouseIcon>

const THEME_CLASSES: Record<TTheme, string> = {
  default: '',
  neutral: 'text-cool-grey-600 dark:text-white/70',
  info: 'text-blue-800 dark:text-blue-600',
  warn: 'text-orange-800 dark:text-orange-600',
  error: 'text-red-800 dark:text-red-500',
  success: 'text-green-800 dark:text-green-500',
  brand: 'text-primary-600 dark:text-primary-500',
}

interface IconProps extends Omit<PhosphorIconProps, 'ref'> {
  variant: TIconVariant
  theme?: TTheme
}

export const Icon = ({
  variant,
  size = 16,
  weight = 'regular',
  theme = 'default',
  className,
  ...props
}: IconProps) => {
  const themeClass = cn(THEME_CLASSES[theme], className)

  if (variant in customIcons) {
    const CustomIcon = customIcons[variant as CustomIconVariant]
    return <CustomIcon size={size} className={themeClass} />
  }

  const IconComponent = PhosphorIcons[variant as PhosphorIconVariant] as
    | ElementType
    | undefined

  if (!IconComponent) {
    return null
  }

  return <IconComponent size={size} weight={weight} className={themeClass} {...props} />
}
