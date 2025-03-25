import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getInstallDeploy } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (
    req: NextRequest,
    { params }: TRouteRes<'org-id' | 'install-id' | 'deploy-id'>
  ) => {
    const orgId = params?.['org-id']
    const installId = params?.['install-id']
    const installDeployId = params?.['deploy-id']

    let installComponent = {}
    try {
      installComponent = await getInstallDeploy({
        orgId,
        installId,
        installDeployId,
      })
    } catch (error) {
      console.error(error)
    }

    return NextResponse.json(installComponent)
  }
)
