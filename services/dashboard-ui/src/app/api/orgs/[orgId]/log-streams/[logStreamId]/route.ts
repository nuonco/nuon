import { type NextRequest, NextResponse } from 'next/server'
import { getLogStreamById } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  _: NextRequest,
  { params }: TRouteProps<'orgId' | 'logStreamId'>
) {
  const { logStreamId, orgId } = await params
  const response = await getLogStreamById({ logStreamId, orgId })
  return NextResponse.json(response)
}
