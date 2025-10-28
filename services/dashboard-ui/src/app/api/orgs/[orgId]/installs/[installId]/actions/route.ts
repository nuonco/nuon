import { type NextRequest, NextResponse } from 'next/server'
import { getInstallActionsLatestRuns } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId'>
) {
  const { installId, orgId } = await params
  const { searchParams } = new URL(request.url)
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('offset') || undefined
  const q = searchParams.get('q') || undefined
  const trigger_types = searchParams.get('trigger_types') || undefined

  const response = await getInstallActionsLatestRuns({
    installId,
    orgId,
    limit,
    offset,
    q,
    trigger_types,
  })
  return NextResponse.json(response)
}
