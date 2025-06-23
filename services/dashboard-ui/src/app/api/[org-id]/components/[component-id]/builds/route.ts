import { NextRequest, NextResponse } from 'next/server'
import { getComponentBuilds } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'component-id'>
) => {
  const { ['org-id']: orgId, ['component-id']: componentId } = await params

  let builds = []
  try {
    builds = await getComponentBuilds({ orgId, componentId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(builds)
}
