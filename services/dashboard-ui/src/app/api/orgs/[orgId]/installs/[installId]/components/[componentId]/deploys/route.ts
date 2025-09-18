import { type NextRequest, NextResponse } from 'next/server'
import { getDeploysByComponentId } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId' | 'componentId'>
) {
  const { installId, componentId, orgId } = await params
  const { searchParams } = new URL(request.url)
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('offset') || undefined
  const q = searchParams.get('q') || undefined

  const response = await getDeploysByComponentId({
    installId,
    componentId,
    orgId,
    limit,
    offset,
    q,
  })
  return NextResponse.json({
    ...response,
    headers: Object.fromEntries(response.headers?.entries()),
  })
}
