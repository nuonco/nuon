import { type NextRequest, NextResponse } from 'next/server'
import { getInstallAuditLog } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  req: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId'>
) {
  const { installId, orgId } = await params
  const start = req.nextUrl.searchParams.get('start')
  const end = req.nextUrl.searchParams.get('end')
  const response = await getInstallAuditLog({ installId, orgId, start, end })
  return NextResponse.json(response)
}
