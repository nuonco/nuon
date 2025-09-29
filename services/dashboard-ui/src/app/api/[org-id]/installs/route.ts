import { NextRequest, NextResponse } from 'next/server'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id'>
) => {
  return NextResponse.json({ deprecated: 'Use /api/orgs/:orgId/installs' })
}
