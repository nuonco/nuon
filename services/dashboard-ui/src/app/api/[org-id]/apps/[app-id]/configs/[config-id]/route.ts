import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { nueQueryData } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (
    req: NextRequest,
    { params }: TRouteRes<'org-id' | 'app-id' | 'config-id'>
  ) => {
    const orgId = params?.['org-id']
    const appId = params?.['app-id']
    const configId = params?.['config-id']

    const res = await nueQueryData({
      orgId,
      path: `apps/${appId}/config/${configId}?recurse=true`,
    })

    return NextResponse.json(res)
  }
)
