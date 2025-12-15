import { type NextRequest, NextResponse } from 'next/server'
import { getLogStreamLogs } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'logStreamId'>
) {
  const { logStreamId, orgId } = await params
  const { searchParams } = new URL(request.url)
  const order = searchParams.get('order') as "asc" || "asc"
  const offset = request.headers.get('X-Nuon-API-Offset') || undefined
  const response = await getLogStreamLogs({
    logStreamId,
    orgId,
    offset,
    order,
  })
  return NextResponse.json(response)
}
