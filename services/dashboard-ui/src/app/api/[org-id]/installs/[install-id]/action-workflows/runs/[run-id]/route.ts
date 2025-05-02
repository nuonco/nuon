import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { nueQueryData } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (
    req: NextRequest,
    { params }: TRouteRes<'org-id' | 'install-id' | 'run-id'>
  ) => {
    const orgId = params?.['org-id']
    const installId = params?.['install-id']
    const runId = params?.['run-id']

    const res = await nueQueryData({
      orgId,
      path: `installs/${installId}/action-workflows/runs/${runId}`
    })

    return NextResponse.json(res)
  }
)
