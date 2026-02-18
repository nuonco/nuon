'use server'

import { revalidatePath } from 'next/cache'
import { updateRunner as update, type IUpdateRunnerBody } from '@/lib'
import { updateMngRunner } from '@/lib/ctl-api/runners/update-mng-runner'
import { reprovisionOrg } from '@/actions/admin/reprovision-org'

export async function updateRunnerWithProvisioning({
  body,
  orgId,
  path,
  runnerId,
  runnerType,
}: {
  body: IUpdateRunnerBody
  orgId: string
  path?: string
  runnerId: string
  runnerType: 'install' | 'org'
}) {
  const updateResult = await update({ body, orgId, runnerId })

  if (updateResult.error) {
    return updateResult
  }

  try {
    if (runnerType === 'org') {
      await reprovisionOrg(orgId)
    } else {
      await updateMngRunner({ orgId, runnerId })
    }
  } catch (postUpdateError) {
    return {
      ...updateResult,
      warning: `Runner updated successfully, but post-update operation failed: ${postUpdateError.message}`,
    }
  }

  if (path) {
    revalidatePath(path)
  }

  return updateResult
}
