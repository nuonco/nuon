import { type NextRequest, NextResponse } from 'next/server'
import {
  getPolicyReports,
  type TPolicyReportOwnerType,
  type TPolicyReportStatus,
} from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId'>
) {
  const { orgId } = await params
  const { searchParams } = new URL(request.url)

  const ownerType = searchParams.get('owner_type') as
    | TPolicyReportOwnerType
    | undefined
  const status = searchParams.get('status') as TPolicyReportStatus | undefined

  const response = await getPolicyReports({
    orgId,
    ownerType: ownerType || undefined,
    ownerId: searchParams.get('owner_id') || undefined,
    appId: searchParams.get('app_id') || undefined,
    installId: searchParams.get('install_id') || undefined,
    status: status || undefined,
    limit: searchParams.get('limit')
      ? Number(searchParams.get('limit'))
      : undefined,
    offset: searchParams.get('offset')
      ? Number(searchParams.get('offset'))
      : undefined,
  })

  return NextResponse.json(response, { status: response.status })
}
