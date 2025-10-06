import { AWS_REGIONS, AZURE_REGIONS } from '@/configs/cloud-regions'
import { getFlagEmoji } from '@/utils/string-utils'
import { Text, IText } from './Text'

interface ICloudRegion extends Omit<IText, 'children'> {
  platform: 'aws' | 'azure'
  location?: string
  region?: string
}

export const CloudRegion = ({
  location,
  platform,
  region,
  ...props
}: ICloudRegion) => {
  const cloudRegion =
    platform === 'azure'
      ? AZURE_REGIONS.find((r) => r.value === location)
      : AWS_REGIONS.find((r) => r.value === region)

  return (
    <Text {...props}>
      {cloudRegion ? (
        <>
          {getFlagEmoji(cloudRegion.iconVariant?.substring(5))}{' '}
          {cloudRegion?.text}
        </>
      ) : (
        'Unknown'
      )}
    </Text>
  )
}
