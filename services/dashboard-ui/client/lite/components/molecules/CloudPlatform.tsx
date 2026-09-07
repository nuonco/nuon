import type { TCloudPlatform } from '@/types'
import { cn } from '@/utils/classnames'
import { Brand, type TBrandVariant } from '../atoms/Brand'
import { Icon } from '../atoms/Icon'
import { Text, type IText } from '../atoms/Text'
import { Tooltip } from '../atoms/Tooltip'

export type TCloudPlatformDisplay = 'abbr' | 'name' | 'icon'

export interface ICloudPlatform extends Omit<IText, 'children'> {
  platform: TCloudPlatform
  display?: TCloudPlatformDisplay
  tone?: 'color' | 'mono'
  iconSize?: number | string
}

const PLATFORMS: Record<
  TCloudPlatform,
  { abbr: string; name: string; brand?: TBrandVariant }
> = {
  aws: {
    abbr: 'AWS',
    name: 'Amazon Web Services',
    brand: 'AWS',
  },
  azure: {
    abbr: 'Azure',
    name: 'Microsoft Azure',
    brand: 'Azure',
  },
  gcp: {
    abbr: 'GCP',
    name: 'Google Cloud',
    brand: 'GCP',
  },
  unknown: {
    abbr: 'Unknown',
    name: 'Unknown',
  },
}

const PlatformMark = ({
  brand,
  size,
  tone,
}: {
  brand?: TBrandVariant
  size: number | string
  tone: 'color' | 'mono'
}) =>
  brand ? (
    <Brand variant={brand} size={size} tone={tone} />
  ) : (
    <Icon variant="QuestionIcon" size={size} />
  )

export const CloudPlatform = ({
  platform,
  display = 'abbr',
  tone = 'color',
  iconSize = '1em',
  className,
  ...props
}: ICloudPlatform) => {
  const config = PLATFORMS[platform] ?? PLATFORMS.unknown

  if (display === 'icon') {
    return (
      <Tooltip content={config.name} tabIndex={0} aria-label={config.name}>
        <PlatformMark brand={config.brand} size={iconSize} tone={tone} />
      </Tooltip>
    )
  }

  return (
    <Text
      className={cn(
        'inline-flex items-center gap-2 whitespace-nowrap',
        className
      )}
      {...props}
    >
      <PlatformMark brand={config.brand} size={iconSize} tone={tone} />
      <span>{display === 'name' ? config.name : config.abbr}</span>
    </Text>
  )
}
