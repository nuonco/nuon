import { NextRequest, NextResponse } from 'next/server'
import { getOrg } from '@/lib'
import type { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id'>
) => {
  const { ['org-id']: orgId } = await params

  let org = {}
  try {
    org = await getOrg({ orgId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(org)
}
