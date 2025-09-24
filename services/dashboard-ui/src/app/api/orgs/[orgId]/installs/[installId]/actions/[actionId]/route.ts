import { type NextRequest, NextResponse } from 'next/server'
import { getInstallActionById } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId' | 'actionId'>
) {
  const { actionId, installId, orgId } = await params
  const { searchParams } = new URL(request.url)
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('offset') || undefined

  const response = await getInstallActionById({
    actionId,
    installId,
    orgId,
    limit,
    offset,
  })
  return NextResponse.json(response)
}
