import { Icon } from '@/components/common/Icon'
import { Text, type IText } from '@/components/common/Text'
import type { TCloudPlatform } from '@/types'

interface ICloudPlatform extends Omit<IText, 'children'> {
  displayVariant?: 'abbr' | 'name' | 'icon-only'
  iconSize?: string
  platform: TCloudPlatform
}

interface ICloudPlatformConfig {
  abbr: string
  iconVariant: 'AWS' | 'Azure' | 'GCP' | 'Question'
  name: string
}

const CLOUD_PLATFORM_CONFIG: Record<
  TCloudPlatform | 'unknown',
  ICloudPlatformConfig
> = {
  aws: {
    abbr: 'AWS',
    iconVariant: 'AWS',
    name: 'Amazone Web Services',
  },
  azure: {
    abbr: 'Azure',
    iconVariant: 'Azure',
    name: 'Micosoft Azure',
  },
  gcp: {
    abbr: 'GCP',
    iconVariant: 'GCP',
    name: 'Google Cloud',
  },
  unknown: {
    abbr: 'Unknown',
    iconVariant: 'Question',
    name: 'Unknown',
  },
} as const

export const CloudPlatform = ({
  displayVariant = 'abbr',
  iconSize,
  platform,
  ...props
}: ICloudPlatform) => {
  const config =
    CLOUD_PLATFORM_CONFIG[platform] || CLOUD_PLATFORM_CONFIG.unknown
  const isIconOnly = displayVariant === 'icon-only'

  return (
    <Text
      className="!flex items-center gap-2 text-nowrap"
      {...props}
      title={isIconOnly ? config.name : undefined}
    >
      <Icon variant={config?.iconVariant} aria-hidden="true" size={iconSize} />
      {!isIconOnly && (
        <span>{displayVariant === 'name' ? config.name : config.abbr}</span>
      )}
    </Text>
  )
}
