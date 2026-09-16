import {
  AWS_REGIONS,
  AZURE_REGIONS,
  GCP_REGIONS,
} from '@/configs/cloud-regions'
import type { TCloudPlatform } from '@/types'

export const cloudRegionsFor = (platform: TCloudPlatform) => {
  if (platform === 'azure') return AZURE_REGIONS
  if (platform === 'gcp') return GCP_REGIONS
  if (platform === 'aws') return AWS_REGIONS
  return []
}
