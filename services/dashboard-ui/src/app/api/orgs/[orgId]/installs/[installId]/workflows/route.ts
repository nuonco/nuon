import { type NextRequest, NextResponse } from 'next/server'
import { getInstallWorkflows } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId'>
) {
  const { orgId, installId } = await params
  const { searchParams } = new URL(request.url)
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('offset') || undefined

  const response = await getInstallWorkflows({
    orgId,
    installId,
    limit,
    offset,
  })
  return NextResponse.json({
    ...response,
    headers: Object.fromEntries(response.headers.entries()),
  })
}
