import { NextRequest, NextResponse } from 'next/server'
import { getInstalls } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id'>
) => {
  const { ['org-id']: orgId } = await params

  let installs = []
  try {
    installs = await getInstalls({ orgId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(installs)
}
