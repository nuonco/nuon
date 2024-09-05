import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getComponentBuilds }  from '@/lib';

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, componentId, __] = req.url.split('/').slice(4, 9)

  let builds = []
  try {
    builds = await getComponentBuilds({ orgId, componentId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(builds)
})
