import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getRunner } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (req: NextRequest, { params }: TRouteRes<'org-id' | 'runner-id'>) => {
    const orgId = params?.['org-id']
    const runnerId = params?.['runner-id']

    let runner = {}
    try {
      runner = await getRunner({ orgId, runnerId })
    } catch (error) {
      console.error(error)
    }

    return NextResponse.json(runner)
  }
)
