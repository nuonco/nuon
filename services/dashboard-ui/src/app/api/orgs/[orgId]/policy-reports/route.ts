import { type NextRequest, NextResponse } from 'next/server'
import { getPolicyReports } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId'>
) {
  const { orgId } = await params
  const { searchParams } = new URL(request.url)

  const response = await getPolicyReports({
    orgId,
    ownerType: searchParams.get('owner_type') || undefined,
    ownerId: searchParams.get('owner_id') || undefined,
    appId: searchParams.get('app_id') || undefined,
    installId: searchParams.get('install_id') || undefined,
    format: searchParams.get('format') || undefined,
    status: searchParams.get('status') || undefined,
    limit: searchParams.get('limit')
      ? Number(searchParams.get('limit'))
      : undefined,
    offset: searchParams.get('offset')
      ? Number(searchParams.get('offset'))
      : undefined,
  })

  return NextResponse.json(response, { status: response.status })
}
