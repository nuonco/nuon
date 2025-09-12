import { type NextRequest, NextResponse } from 'next/server'
import { getInstallDeployById } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId' | 'deployId'>
) {
  const { installId, deployId, orgId } = await params

  const response = await getInstallDeployById({
    installId,
    deployId,
    orgId,
  })
  return NextResponse.json(response)
}
