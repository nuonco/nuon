import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getRunner } from '@/lib'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, runnerId] = req.url.split('/').slice(4, 7)

  let runner = {}
  try {
    runner = await getRunner({ orgId, runnerId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(runner)
})
