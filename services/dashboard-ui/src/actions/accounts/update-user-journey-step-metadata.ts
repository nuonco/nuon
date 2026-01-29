'use server'

import { executeServerAction } from '@/actions/execute-server-action'
import { updateUserJourneyStepMetadata as update } from '@/lib'

export async function updateUserJourneyStepMetadata({
  journeyName,
  stepName,
  metadata,
  complete,
  path,
}: {
  journeyName: string
  stepName: string
  metadata: Record<string, string>
  complete?: boolean
  path?: string
}) {
  return executeServerAction({
    action: update,
    args: { journeyName, stepName, metadata, complete },
    path,
  })
}
