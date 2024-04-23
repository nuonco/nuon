import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import type { TInstall } from '@/types'
import { API_URL, getFetchOpts, getFullInstallStatus } from '@/utils'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, installId] = req.url.split('/').slice(4, 6)

  let status = {
    componentStatus: {
      status: 'error',
      status_description: 'Failed to get component statuses',
    },
    installStatus: {
      status: 'error',
      status_description: 'Failed to get install status',
    },
    sandboxStatus: {
      status: 'error',
      status_description: 'Failed to get sandbox status',
    },
  }
  try {
    const data = await fetch(
      `${API_URL}/v1/installs/${installId}`,
      await getFetchOpts(orgId)
    )
    const install = (await data.json()) as TInstall
    status = getFullInstallStatus(install)
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(status)
})
