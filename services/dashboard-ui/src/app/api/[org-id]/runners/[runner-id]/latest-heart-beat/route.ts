import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getRunnerLatestHeartbeat } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (req: NextRequest, { params }: TRouteRes<'org-id' | 'runner-id'>) => {
    const orgId = params?.['org-id']
    const runnerId = params?.['runner-id']

    let heartbeat = {}
    try {
      heartbeat = await getRunnerLatestHeartbeat({ orgId, runnerId })
    } catch (error) {
      console.error(error)
    }

    return NextResponse.json(heartbeat)
  }
)
