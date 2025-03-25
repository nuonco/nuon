import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getInstalls } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (req: NextRequest, { params }: TRouteRes<'org-id'>) => {
    const orgId = params?.['org-id']

    let installs = []
    try {
      installs = await getInstalls({ orgId })
    } catch (error) {
      console.error(error)
    }

    return NextResponse.json(installs)
  }
)
