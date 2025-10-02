'use server'

import { auth0 } from '@/lib/auth'
import { executeServerAction } from '@/actions/execute-server-action'
import { SF_TRIAL_ACCESS_ENDPOINT } from '@/configs/app'
import { createOrg as create, type TCreateOrgBody } from '@/lib'

export async function createOrg({
  body,
  path,
}: {
  body: TCreateOrgBody
  path?: string
}) {
  const session = await auth0.getSession()

  if (SF_TRIAL_ACCESS_ENDPOINT) {
    await fetch(SF_TRIAL_ACCESS_ENDPOINT, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        firstName: session?.user?.given_name,
        lastName: session?.user?.family_name,
        email: session?.user?.email,
        companyName: body?.name,
        jobTitle: '',
        description: '',
      }),
    }).catch((err) => {
      console.error('error posting to salesforce api:', err)
    })
  }

  return executeServerAction({
    action: create,
    args: { body },
    path,
  })
}
