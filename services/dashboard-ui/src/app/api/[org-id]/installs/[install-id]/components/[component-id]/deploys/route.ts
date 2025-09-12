import { NextRequest, NextResponse } from 'next/server'
import { getInstallComponentDeploys } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'install-id' | 'component-id'>
) => {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['component-id']: componentId,
  } = await params

  const response = await getInstallComponentDeploys({
    orgId,
    installId,
    componentId,
  })

  return NextResponse.json(response)
}
