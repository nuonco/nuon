'use server'

import { revalidatePath } from 'next/cache'
import { joinWaitlist, type IJoinWaitlist } from '@/lib'
import type { TInvite, TWaitlist } from '@/types'
import { mutateData } from '@/utils'

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
