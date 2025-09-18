import { NextRequest, NextResponse } from 'next/server'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'install-id'>
) => {
  return NextResponse.json({
    deprecated: 'use /api/orgs/:orgId/installs/:installId',
  })
}
