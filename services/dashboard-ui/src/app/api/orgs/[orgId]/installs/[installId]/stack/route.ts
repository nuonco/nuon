import { type NextRequest, NextResponse } from 'next/server'
import { getInstallStack } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  _: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId'>
) {
  const { installId, orgId } = await params
  const response = await getInstallStack({ installId, orgId })
  return NextResponse.json(response)
}
