import { NextRequest, NextResponse } from 'next/server'
import { getComponentBuild } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'build-id'>
) => {
  const { ['org-id']: orgId, ['build-id']: buildId } = await params

  let build = {}
  try {
    build = await getComponentBuild({ orgId, buildId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(build)
}
