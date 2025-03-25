import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getInstallComponent } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (
    req: NextRequest,
    { params }: TRouteRes<'org-id' | 'install-id' | 'component-id'>
  ) => {
    const orgId = params?.['org-id']
    const installId = params?.['install-id']
    const componentId = params?.['component-id']

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
)
