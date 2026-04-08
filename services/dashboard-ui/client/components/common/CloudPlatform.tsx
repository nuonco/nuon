import { Icon } from '@/components/common/Icon'
import { Text, type IText } from '@/components/common/Text'
import type { TCloudPlatform } from '@/types'

interface ICloudPlatform extends Omit<IText, 'children'> {
  colorVariant?: 'mono' | 'color'
  displayVariant?: 'abbr' | 'name' | 'icon-only'
  iconSize?: string
  platform: TCloudPlatform
}

interface ICloudPlatformConfig {
  abbr: string
  iconVariant: 'AWS' | 'Azure' | 'GCP' | 'Question'
  iconVariantColor: 'AWSColor' | 'AzureColor' | 'GCPColor' | 'Question'
  name: string
}

const CLOUD_PLATFORM_CONFIG: Record<
  TCloudPlatform | 'unknown',
  ICloudPlatformConfig
> = {
  aws: {
    abbr: 'AWS',
    iconVariant: 'AWS',
    iconVariantColor: 'AWSColor',
    name: 'Amazone Web Services',
  },
  azure: {
    abbr: 'Azure',
    iconVariant: 'Azure',
    iconVariantColor: 'AzureColor',
    name: 'Micosoft Azure',
  },
  gcp: {
    abbr: 'GCP',
    iconVariant: 'GCP',
    iconVariantColor: 'GCPColor',
    name: 'Google Cloud',
  },
  unknown: {
    abbr: 'Unknown',
    iconVariant: 'Question',
    iconVariantColor: 'Question',
    name: 'Unknown',
  },
} as const

export const CloudPlatform = ({
  colorVariant = 'mono',
  displayVariant = 'abbr',
  iconSize,
  platform,
  ...props
}: ICloudPlatform) => {
  const config =
    CLOUD_PLATFORM_CONFIG[platform] || CLOUD_PLATFORM_CONFIG.unknown
  const isIconOnly = displayVariant === 'icon-only'
  const iconVariant =
    colorVariant === 'color' ? config.iconVariantColor : config.iconVariant

  return (
    <Text
      flex
      className="gap-2 text-nowrap"
      {...props}
      title={isIconOnly ? config.name : undefined}
    >
      <Icon variant={iconVariant} aria-hidden="true" size={iconSize} />
      {!isIconOnly && (
        <span>{displayVariant === 'name' ? config.name : config.abbr}</span>
      )}
    </Text>
  )
}
