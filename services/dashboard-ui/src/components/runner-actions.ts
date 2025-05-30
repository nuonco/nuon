'use server'

import { revalidatePath } from 'next/cache'
import { cancelRunnerJob as cancelJob } from '@/lib'
import { nueMutateData, mutateData } from '@/utils'

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

interface IShutdownRunner {
  orgId: string
  runnerId: string
  path: string
  force?: boolean
}

export async function shutdownRunner({
  force = false,
  orgId,
  path,
  runnerId,
}: IShutdownRunner) {
  const reqPath = force
    ? `runners/${runnerId}/force-shutdown`
    : `runners/${runnerId}/graceful-shutdown`

  return nueMutateData({
    orgId,
    path: reqPath,
    body: {},
  }).then((r) => {
    revalidatePath(path)
    return r
  })
}

export interface IUnlockWorkspace {
  workspaceId: string,
  orgId: string,
}

export async function unlockWorkspace({
  workspaceId,
  orgId,
}: IUnlockWorkspace) {
  return mutateData<any>({
    errorMessage: 'Unable to lock workspace state.',
    orgId,
    path: `terraform-workspaces/${workspaceId}/unlock`,
    data: {}
  })
}
