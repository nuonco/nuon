import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getInstallSandboxRun } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = withApiAuthRequired(
  async (
    req: NextRequest,
    { params }: TRouteRes<'org-id' | 'install-id' | 'run-id'>
  ) => {
    const orgId = params?.['org-id']
    const installId = params?.['install-id']
    const installSandboxRunId = params?.['run-id']

    let run = {}
    try {
      run = await getInstallSandboxRun({
        orgId,
        installId,
        installSandboxRunId,
      })
    } catch (error) {
      console.error(error)
    }

    return NextResponse.json(run)
  }
)
