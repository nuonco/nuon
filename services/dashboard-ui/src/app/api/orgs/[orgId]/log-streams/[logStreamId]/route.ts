import { type NextRequest, NextResponse } from 'next/server'
import { getLogStream } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  _: NextRequest,
  { params }: TRouteProps<'orgId' | 'logStreamId'>
) {
  const { logStreamId, orgId } = await params
  const response = await getLogStream({ logStreamId, orgId })
  return NextResponse.json(response)
}
