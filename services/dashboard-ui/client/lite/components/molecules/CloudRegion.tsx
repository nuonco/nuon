import { AWS_REGIONS, AZURE_REGIONS, GCP_REGIONS } from '@/configs/cloud-regions'
import type { TCloudPlatform } from '@/types'
import { cn } from '@/utils/classnames'
import { getFlagEmoji } from '@/utils/string-utils'
import { Text, type IText } from '../atoms/Text'

export interface ICloudRegion extends Omit<IText, 'children'> {
  platform: TCloudPlatform
  location?: string
  region?: string
}

const regionsFor = (platform: TCloudPlatform) => {
  if (platform === 'azure') return AZURE_REGIONS
  if (platform === 'gcp') return GCP_REGIONS
  return AWS_REGIONS
}

export const CloudRegion = ({
  platform,
  location,
  region,
  variant = 'caption',
  className,
  ...props
}: ICloudRegion) => {
  const value = platform === 'azure' ? location : region
  const cloudRegion =
    platform === 'unknown'
      ? undefined
      : regionsFor(platform).find((item) => item.value === value)
  const countryCode = cloudRegion?.iconVariant?.replace(/^flag-/, '')

  return (
    <Text
      variant={variant}
      className={cn(
        'inline-flex min-w-0 items-center gap-1.5 whitespace-nowrap',
        className
      )}
      {...props}
    >
      {cloudRegion ? (
        <>
          {countryCode ? (
            <span aria-hidden className="shrink-0">
              {getFlagEmoji(countryCode)}
            </span>
          ) : null}
          <span className="truncate">{cloudRegion.text}</span>
        </>
      ) : (
        'Unknown'
      )}
    </Text>
  )
}
