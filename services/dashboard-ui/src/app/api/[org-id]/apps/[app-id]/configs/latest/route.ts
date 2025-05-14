import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { nueQueryData } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (
    req: NextRequest,
    { params }: TRouteRes<'org-id' | 'app-id'>
  ) => {
    const orgId = params?.['org-id']
    const appId = params?.['app-id']

    const res = await nueQueryData({
      orgId,
      path: `apps/${appId}/latest-config`,
    })

    return NextResponse.json(res)
  }
)
