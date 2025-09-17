import { NextRequest, NextResponse } from 'next/server'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'build-id'>
) => {
  return NextResponse.json({
    deprecated: 'use /api/orgs/:orgId/components/:componentId/builds/:buildId',
  })
}
