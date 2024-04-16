import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { API_URL, getFetchOpts } from '@/utils'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, installId, _, componentId] = req.url.split('/').slice(4, 8)
  const data = await fetch(
    `${API_URL}/v1/installs/${installId}/components/${componentId}/deploys`,
    await getFetchOpts(orgId)
  )
  const deploy = await data
    .json()
    .then((d) => d?.[0] || { error: 'no deploys' })
  let status = {
    status: 'failed',
    status_description: `No active deploy on install ${installId}`,
  }
  if (!deploy.error) {
    status = {
      status: deploy?.status,
      status_description: deploy?.status_description,
    }
  }

  return NextResponse.json(status)
})
