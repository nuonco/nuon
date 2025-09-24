import { type NextRequest, NextResponse } from 'next/server'
import { getApps } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId'>
) {
  const { orgId } = await params
  const { searchParams } = new URL(request.url)
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('offset') || undefined
  const q = searchParams.get('q') || undefined

  const response = await getApps({ orgId, limit, offset, q })
  return NextResponse.json(response)
}
