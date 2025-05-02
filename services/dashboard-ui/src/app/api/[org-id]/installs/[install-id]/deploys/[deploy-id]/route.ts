import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { nueQueryData } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (
    req: NextRequest,
    { params }: TRouteRes<'org-id' | 'install-id' | 'deploy-id'>
  ) => {
    const orgId = params?.['org-id']
    const installId = params?.['install-id']
    const installDeployId = params?.['deploy-id']

    const res = await nueQueryData({
      orgId,
      path: `installs/${installId}/deploys/${installDeployId}`,
    })

    return NextResponse.json(res)
  }
)
