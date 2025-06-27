'use server'

import { revalidatePath } from 'next/cache'
import { joinWaitlist, type IJoinWaitlist } from '@/lib'
import type { TInvite, TWaitlist } from '@/types'
import { mutateData, nueMutateData } from '@/utils'

export async function requestWaitlistAccess(
  formData: FormData
): Promise<TWaitlist> {
  const data = Object.fromEntries(formData) as IJoinWaitlist

  return joinWaitlist(data)
}

export async function inviteUserToOrg(
  formData: FormData,
  orgId: string
): Promise<TInvite> {
  const data = Object.fromEntries(formData)

  return mutateData({
    errorMessage: 'Unable to invite user',
    data,
    path: `orgs/current/invites`,
    orgId,
  }).then((invite) => {
    revalidatePath(`/${orgId}/team`)
    return invite
  })
}

export async function connectGitHubToOrg({
  org_id,
  github_install_id,
}: {
  github_install_id: string
  org_id: string
}) {
  return nueMutateData({
    orgId: org_id,
    path: 'vcs/connection-callback',
    body: { github_install_id, org_id },
  }).then((res) => {
    revalidatePath(`/${org_id}`)
    return res
  })
}

export async function removeUserFromOrg(
  data: { user_id: string },
  orgId: string
): Promise<TInvite> {
  return mutateData({
    errorMessage: 'Unable to remove user',
    data,
    path: `orgs/current/remove-user`,
    orgId,
  }).then((invite) => {
    revalidatePath(`/${orgId}/team`)
    return invite
  })
}
