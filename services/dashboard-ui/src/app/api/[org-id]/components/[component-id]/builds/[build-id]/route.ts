import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getBuild }  from '@/lib';

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, __, ___, buildId] = req.url.split('/').slice(4, 9)

  let build = {}
  try {
    build = await getBuild({ orgId, buildId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(build)
})
