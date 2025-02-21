'use server'

import { revalidatePath } from 'next/cache'
import { cancelRunnerJob as cancelJob } from '@/lib'

interface ICancelRunnerJob {
  orgId: string
  path: string
  runnerJobId: string
}

export async function cancelRunnerJob({
  orgId,
  path,
  runnerJobId,
}: ICancelRunnerJob) {
  try {
    await cancelJob({
      runnerJobId,
      orgId,
    })

    revalidatePath(path)
  } catch (error) {
    console.error(error)
    throw new Error(
      `${error?.message || 'An error occured.'} Please refresh page and try again.`
    )
  }
}
