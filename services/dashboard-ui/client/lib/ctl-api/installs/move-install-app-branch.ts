import { api } from '@/lib/api'
import type { TInstall } from '@/types'

export type TMoveInstallAppBranchBody = {
  app_branch_id: string
}

export async function moveInstallAppBranch({
  installId,
  orgId,
  body,
}: {
  installId: string
  orgId: string
  body: TMoveInstallAppBranchBody
}) {
  return api<TInstall>({
    body,
    method: 'PATCH',
    orgId,
    path: `installs/${installId}/app-branch`,
  })
}
