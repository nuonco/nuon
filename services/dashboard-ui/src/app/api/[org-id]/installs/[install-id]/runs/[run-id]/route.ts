import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getSandboxRun } from '@/lib';

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, installId, __, runId] = req.url.split('/').slice(4, 9)

  let run = {}
  try {    
    run = await getSandboxRun({ orgId, installId, runId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(run)
})
