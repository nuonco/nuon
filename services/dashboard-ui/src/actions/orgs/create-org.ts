'use server'

import { getSession } from '@/lib/auth-server'
import { executeServerAction } from '@/actions/execute-server-action'
import { SF_TRIAL_ACCESS_ENDPOINT } from '@/configs/app'
import { createOrg as create, type TCreateOrgBody } from '@/lib'

export async function createOrg({
  body,
  path,
}: {
  body: TCreateOrgBody & {
    companyName?: string
    jobTitle?: string
    notes?: string
  }
  path?: string
}) {
  const session = await getSession()

  if (SF_TRIAL_ACCESS_ENDPOINT) {
    const firstName = session?.user?.given_name || session?.user?.name
    const lastName = session?.user?.family_name || ''
    const requestBody = JSON.stringify({
      firstName,
      lastName,
      email: session?.user?.email,
      companyName: `${body?.companyName || 'N/A'} | ${body?.name}`,
      jobTitle: body?.jobTitle || 'N/A',
      notes: body?.notes || 'N/A',
      subject: 'trial-signup',
    })

    await fetch(SF_TRIAL_ACCESS_ENDPOINT, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: requestBody,
    }).catch((err) => {
      console.error('error posting to salesforce api:', err)
    })
  }

  return executeServerAction({
    action: create,
    args: {
      body: { name: body?.name, use_sandbox_mode: body?.use_sandbox_mode },
    },
    path,
  })
}
