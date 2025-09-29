'use server'

import { revalidatePath } from 'next/cache'
import { cancelRunnerJob as cancelJob } from '@/lib'
import { nueMutateData, mutateData } from '@/utils'

// interface ICancelRunnerJob {
//   orgId: string
//   path: string
//   runnerJobId: string
// }

// export async function cancelRunnerJob({
//   orgId,
//   path,
//   runnerJobId,
// }: ICancelRunnerJob) {
//   try {
//     await cancelJob({
//       runnerJobId,
//       orgId,
//     })

//     revalidatePath(path)
//   } catch (error) {
//     console.error(error)
//     throw new Error(
//       `${error?.message || 'An error occured.'} Please refresh page and try again.`
//     )
//   }
// }

// interface IShutdownRunner {
//   orgId: string
//   runnerId: string
//   path: string
//   force?: boolean
// }

// export async function shutdownRunner({
//   force = false,
//   orgId,
//   path,
//   runnerId,
// }: IShutdownRunner) {
//   const reqPath = force
//     ? `runners/${runnerId}/force-shutdown`
//     : `runners/${runnerId}/graceful-shutdown`

//   return nueMutateData({
//     orgId,
//     path: reqPath,
//     body: {},
//   }).then((r) => {
//     revalidatePath(path)
//     return r
//   })
// }

// interface IUpdateRunner {
//   orgId: string
//   runnerId: string
//   path: string
//   body: {
//     container_image_tag: string
//     container_image_url: string
//     org_awsiam_role_arn: string
//     org_k8s_service_account_name: string
//     runner_api_url: string
//   }
// }

// export async function updateRunner({
//   body,
//   orgId,
//   path,
//   runnerId,
// }: IUpdateRunner) {
//   return nueMutateData({
//     orgId,
//     method: 'PATCH',
//     path: `runners/${runnerId}/settings`,
//     body,
//   }).then((r) => {
//     revalidatePath(path)
//     return r
//   })
// }

export interface IUnlockWorkspace {
  workspaceId: string
  orgId: string
}

export async function unlockWorkspace({
  workspaceId,
  orgId,
}: IUnlockWorkspace) {
  return mutateData<any>({
    errorMessage: 'Unable to lock workspace state.',
    orgId,
    path: `terraform-workspaces/${workspaceId}/unlock`,
    data: {},
  })
}
