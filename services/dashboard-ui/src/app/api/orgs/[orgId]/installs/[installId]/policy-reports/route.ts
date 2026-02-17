import { type NextRequest, NextResponse } from 'next/server'
import { getInstallPolicyReports } from '@/lib'
import type {
  TPolicyReportOwnerType,
  TPolicyReportStatus,
} from '@/lib/ctl-api/installs/get-install-policy-reports'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId'>
) {
  const { installId, orgId } = await params
  const { searchParams } = new URL(request.url)
  const ownerType = (searchParams.get('owner_type') || undefined) as
    | TPolicyReportOwnerType
    | undefined
  const status = (searchParams.get('status') || undefined) as
    | TPolicyReportStatus
    | undefined
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('offset') || undefined

  const response = await getInstallPolicyReports({
    installId,
    orgId,
    ownerType,
    status,
    limit,
    offset,
  })
  return NextResponse.json(response)
}
