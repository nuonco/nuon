import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import type {
  TInstall,
  TInstallComponent,
  TInstallDeploy,
  TSandboxRun,
} from '@/types'
import { API_URL, getFetchOpts, getFullInstallStatus } from '@/utils'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, installId] = req.url.split('/').slice(4, 6)
  const data = await fetch(
    `${API_URL}/v1/installs/${installId}`,
    await getFetchOpts(orgId)
  )
  const install = (await data.json()) as TInstall

  return NextResponse.json(getFullInstallStatus(install))
})
