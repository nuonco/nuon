import { NextRequest, NextResponse } from 'next/server'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'runner-id'>
) => {
  return NextResponse.json({
    depricated: 'use the new /api/orgs/:id/runners/:id/jobs endpoint',
  })
}
