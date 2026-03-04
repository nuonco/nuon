'use server'

import { getSession } from '@/lib/auth-server'
import { executeServerAction } from '@/actions/execute-server-action'
import { SF_TRIAL_ACCESS_ENDPOINT } from '@/configs/app'
import { createOrg as create, type TCreateOrgBody } from '@/lib'
import { statsd } from '@/lib/metrics'

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
    const lastName = session?.user?.family_name || 'LNU'
    const requestBody = JSON.stringify({
      firstName,
      lastName,
      email: session?.user?.email,
      companyName: `${body?.companyName || 'N/A'} | ${body?.name}`,
      jobTitle: body?.jobTitle || 'N/A',
      notes: body?.notes || 'N/A',
      subject: 'trial-signup',
    })

    const sfMetric = 'ui.request.upstream.latency'
    const sfStart = Date.now()
    const sfTags = {
      upstream: 'salesforce',
      method: 'post',
      endpoint: new URL(SF_TRIAL_ACCESS_ENDPOINT).pathname,
      status: 'unknown',
      status_code_class: 'unknown',
    }
    await fetch(SF_TRIAL_ACCESS_ENDPOINT, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: requestBody,
    })
      .then((res) => {
        sfTags.status = res.status < 400 ? 'success' : 'error'
        sfTags.status_code_class = Math.floor(res.status / 100) + 'xx'
        statsd?.timing(sfMetric, Date.now() - sfStart, sfTags)
      })
      .catch((err) => {
        console.error('error posting to salesforce api:', err)
        sfTags.status = 'error'
        statsd?.timing(sfMetric, Date.now() - sfStart, sfTags)
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
