import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getOrg } from '@/lib'
import type { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (req: NextRequest, { params }: TRouteRes<'org-id'>) => {
    const orgId = params?.['org-id']

    let org = {}
    try {
      org = await getOrg({ orgId })
    } catch (error) {
      console.error(error)
    }

    return NextResponse.json(org)
  }
)
