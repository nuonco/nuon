import { type NextRequest, NextResponse } from 'next/server'
import { getLogsByLogStreamId } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'logStreamId'>
) {
  const { logStreamId, orgId } = await params
  const offset = request.headers.get('X-Nuon-API-Offset') || undefined
  const response = await getLogsByLogStreamId({
    logStreamId,
    orgId,
    offset,
  })
  return NextResponse.json({
    ...response,
    headers: Object.fromEntries(response.headers?.entries()),
  })
}
