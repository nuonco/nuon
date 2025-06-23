import { NextRequest, NextResponse } from 'next/server'
import { getInstallSandboxRuns } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'install-id'>
) => {
  const { ['org-id']: orgId, ['install-id']: installId } = await params

  let sandboxRuns = []
  try {
    sandboxRuns = await getInstallSandboxRuns({ orgId, installId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(sandboxRuns)
}
