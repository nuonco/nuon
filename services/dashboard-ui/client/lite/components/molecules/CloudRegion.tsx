import type { TCloudPlatform } from '@/types'
import { cn } from '@/utils/classnames'
import { getFlagEmoji } from '@/utils/string-utils'
import { cloudRegionsFor } from '../../utils/cloud-regions'
import { Text, type IText } from '../atoms/Text'

export interface ICloudRegion extends Omit<IText, 'children'> {
  platform: TCloudPlatform
  location?: string
  region?: string
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
  const cloudRegion = cloudRegionsFor(platform).find(
    (item) => item.value === value
  )
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
