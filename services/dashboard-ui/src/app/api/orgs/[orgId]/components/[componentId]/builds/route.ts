import { NextRequest, NextResponse } from 'next/server'
import { getComponentBuilds } from '@/lib'
import type { TRouteProps } from '@/types'

export const GET = async (
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'componentId'>
) => {
  const { orgId, componentId } = await params
  const { searchParams } = new URL(request.url)
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('offset') || undefined

  const response = await getComponentBuilds({
    orgId,
    componentId,
    offset,
    limit,
  })
  return NextResponse.json({
    ...response,
    headers: Object.fromEntries(response.headers.entries()),
  })
}
