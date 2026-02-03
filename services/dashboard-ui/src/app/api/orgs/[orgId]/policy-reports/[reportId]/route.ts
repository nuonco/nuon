import { type NextRequest, NextResponse } from 'next/server'
import { getPolicyReport } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  _: NextRequest,
  { params }: TRouteProps<'orgId' | 'reportId'>
) {
  const { reportId, orgId } = await params
  const response = await getPolicyReport({
    reportId,
    orgId,
  })

  return NextResponse.json(response, { status: response.status })
}
