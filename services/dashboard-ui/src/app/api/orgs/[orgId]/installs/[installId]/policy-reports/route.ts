import { type NextRequest, NextResponse } from 'next/server'
import { getInstallPolicyReports } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId'>
) {
  const { installId, orgId } = await params
  const { searchParams } = new URL(request.url)
  const ownerType = searchParams.get('owner_type') || undefined
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('offset') || undefined

  const response = await getInstallPolicyReports({
    installId,
    orgId,
    ownerType,
    limit,
    offset,
  })
  return NextResponse.json(response)
}
