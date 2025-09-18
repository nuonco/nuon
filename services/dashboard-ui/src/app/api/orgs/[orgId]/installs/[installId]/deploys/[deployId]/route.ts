import { type NextRequest, NextResponse } from 'next/server'
import { getDeployById } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId' | 'deployId'>
) {
  const { installId, deployId, orgId } = await params

  const response = await getDeployById({
    installId,
    deployId,
    orgId,
  })
  return NextResponse.json(response)
}
