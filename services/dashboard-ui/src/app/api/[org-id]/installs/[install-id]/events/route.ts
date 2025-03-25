import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getInstallEvents } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (req: NextRequest, { params }: TRouteRes<'org-id' | 'install-id'>) => {
    const orgId = params?.['org-id']
    const installId = params?.['install-id']

    let events = []
    try {
      events = await getInstallEvents({ orgId, installId })
    } catch (error) {
      console.error(error)
    }

    return NextResponse.json(events)
  }
)
