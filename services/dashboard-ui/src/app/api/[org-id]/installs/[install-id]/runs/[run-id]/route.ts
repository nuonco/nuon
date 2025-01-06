import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getInstallSandboxRun } from '@/lib'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, installId, __, installSandboxRunId] = req.url
    .split('/')
    .slice(4, 9)

  let run = {}
  try {
    run = await getInstallSandboxRun({ orgId, installId, installSandboxRunId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(run)
})
