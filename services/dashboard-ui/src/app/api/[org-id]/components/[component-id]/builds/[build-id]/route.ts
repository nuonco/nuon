import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getComponentBuild } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (req: NextRequest, { params }: TRouteRes<'org-id' | 'build-id'>) => {
    const orgId = params?.['org-id']
    const buildId = params?.['build-id']

    let build = {}
    try {
      build = await getComponentBuild({ orgId, buildId })
    } catch (error) {
      console.error(error)
    }

    return NextResponse.json(build)
  }
)
