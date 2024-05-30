import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getInstallEvents } from '@/lib'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, installId] = req.url.split('/').slice(4, 7)

  let events = []
  try {
    events = await getInstallEvents({ orgId, installId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(events)
})
