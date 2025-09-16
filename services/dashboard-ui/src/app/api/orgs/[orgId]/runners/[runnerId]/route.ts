import { type NextRequest, NextResponse } from 'next/server'
import { getRunnerById } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'runnerId'>
) {
  const { runnerId, orgId } = await params

  const response = await getRunnerById({
    runnerId,
    orgId,
  })
  return NextResponse.json(response)
}
