import { api } from '@/lib/api'
import type { TAccount } from '@/types'

export const updateUserJourneyStepMetadata = ({
  journeyName,
  stepName,
  metadata,
  complete,
}: {
  journeyName: string
  stepName: string
  metadata: Record<string, string>
  complete?: boolean
}) =>
  api<TAccount>({
    method: 'PATCH',
    path: `account/user-journeys/${journeyName}/steps/${stepName}`,
    body: { metadata, ...(complete !== undefined && { complete }) },
  })
