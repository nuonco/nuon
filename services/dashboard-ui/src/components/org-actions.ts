'use server'

import { revalidatePath } from 'next/cache'
import { cookies } from 'next/headers'
import { auth0 } from '@/lib/auth'
import { joinWaitlist, type IJoinWaitlist } from '@/lib'
import type { TInvite, TWaitlist, TOrg } from '@/types'
import { mutateData, nueMutateData, SF_TRIAL_ACCESS_ENDPOINT } from '@/utils'
import { getFetchOpts } from '@/utils/get-fetch-opts'
import { API_URL } from '@/configs/api'

export async function requestWaitlistAccess(
  formData: FormData
): Promise<TWaitlist> {
  const session = await auth0.getSession()
  const data = Object.fromEntries(formData) as IJoinWaitlist

  if (SF_TRIAL_ACCESS_ENDPOINT) {
    await fetch(SF_TRIAL_ACCESS_ENDPOINT, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        firstName: session?.user?.given_name,
        lastName: session?.user?.family_name,
        email: session?.user?.email,
        companyName: data?.org_name,
        jobTitle: data?.job_title,
        description: data?.tell_us_more,
      }),
    }).catch((err) => {
      console.error('error posting to salesforce api')
    })
  }

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

export async function createOrganization(name: string) {
  return nueMutateData<TOrg>({
    path: `orgs`,
    body: {
      name: name,
      'sandbox-mode': false,
    },
  })
}

export async function createTrialOrganization() {
  const session = await auth0.getSession()

  if (!session?.user?.email) {
    return {
      data: null,
      error: { error: 'Unable to get user email for org creation' }
    }
  }

  // Generate org name from email (same pattern as backend was using)
  const orgName = `${session.user.email}-trial`

  return nueMutateData<TOrg>({
    path: `orgs`,
    body: {
      name: orgName,
      'sandbox-mode': false,
    },
  })
}

export async function setOrgSessionCookie(orgId: string) {
  const c = await cookies()
  c.set('org-session', orgId)
}

export async function completeUserJourney(journeyName: string) {
  return fetch(`${API_URL}/v1/account/user-journeys/${journeyName}/complete`, {
    ...(await getFetchOpts()), // ⚠️ CRITICAL: Global endpoint - do not pass orgId
    method: 'POST',
  }).then(async (response) => {
    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(`Failed to complete user journey: ${response.status} ${errorText}`)
    }
    return response.json()
  })
}
