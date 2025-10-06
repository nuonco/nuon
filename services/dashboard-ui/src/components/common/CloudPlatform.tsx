import { Icon } from '@/components/common/Icon'
import { Text, type IText } from '@/components/common/Text'
import type { TCloudPlatform } from '@/types'

interface ICloudPlatform extends Omit<IText, 'children'> {
  displayVariant?: 'abbr' | 'name' | 'icon-only'
  platform: TCloudPlatform
}

interface ICloudPlatformConfig {
  abbr: string
  icon: React.ReactElement
  name: string
}

const CLOUD_PLATFORM_CONFIG: Record<
  TCloudPlatform | 'unknown',
  ICloudPlatformConfig
> = {
  aws: {
    abbr: 'AWS',
    icon: <Icon variant="AWS" aria-hidden="true" />,
    name: 'Amazone Web Services',
  },
  azure: {
    abbr: 'Azure',
    icon: <Icon variant="Azure" aria-hidden="true" />,
    name: 'Micosoft Azure',
  },
  gcp: {
    abbr: 'GCP',
    icon: <Icon variant="GCP" aria-hidden="true" />,
    name: 'Google Cloud',
  },
  unknown: {
    abbr: 'Unknown',
    icon: <Icon variant="Question" aria-hidden="true" />,
    name: 'Unknown',
  },
} as const

export const CloudPlatform = ({
  displayVariant = 'abbr',
  platform,
  ...props
}: ICloudPlatform) => {
  const config =
    CLOUD_PLATFORM_CONFIG[platform] || CLOUD_PLATFORM_CONFIG.unknown
  const isIconOnly = displayVariant === 'icon-only'

  return (
    <Text
      className="flex items-center gap-2 text-nowrap"
      {...props}
      title={isIconOnly ? config.name : undefined}
    >
      {config.icon}
      {!isIconOnly && (
        <span>{displayVariant === 'name' ? config.name : config.abbr}</span>
      )}
    </Text>
  )
}
