import { NextRequest, NextResponse } from 'next/server'
import { nueQueryData } from '@/utils'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'install-id'>
) => {
  const { searchParams } = new URL(req.url)
  const offset = searchParams.get('offset')
  const { ['org-id']: orgId, ['install-id']: installId } = await params

  const sp = new URLSearchParams({ offset, limit: '6' }).toString()
  const res = await nueQueryData({
    orgId,
    path: `installs/${installId}/sandbox-runs${sp ? '?' + sp : sp}`,
    headers: {
      'x-nuon-pagination-enabled': true,
    },
  })

  return NextResponse.json(res)
}
