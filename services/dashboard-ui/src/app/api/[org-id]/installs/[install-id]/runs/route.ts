import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getInstallSandboxRuns } from '@/lib'

export const GET = withApiAuthRequired(async (req: NextRequest, { params }) => {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string

  let sandboxRuns = []
  try {
    sandboxRuns = await getInstallSandboxRuns({ orgId, installId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(sandboxRuns)
})
