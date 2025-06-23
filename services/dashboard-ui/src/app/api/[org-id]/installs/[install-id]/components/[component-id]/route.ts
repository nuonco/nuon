import { NextRequest, NextResponse } from 'next/server'
import { getInstallComponent } from '@/lib'
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

  let installComponent = {}
  try {
    installComponent = await getInstallComponent({
      orgId,
      installId,
      componentId,
    })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(installComponent)
}
