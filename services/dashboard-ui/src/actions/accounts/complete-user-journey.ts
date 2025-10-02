'use server'

import { executeServerAction } from '@/actions/execute-server-action'
import { completeUserJourney as complete } from '@/lib'

export async function completeUserJourney({
  journeyName,
  path,
}: {
  journeyName: string
  path?: string
}) {
  return executeServerAction({
    action: complete,
    args: { journeyName },
    path,
  })
}
