import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { API_URL, getFetchOpts } from '@/utils'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, installId] = req.url.split('/').slice(4, 7)

  let events = []
  try {
    const data = await fetch(
      `${API_URL}/v1/installs/${installId}/events`,
      await getFetchOpts(orgId)
    )
    events = await data.json()
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(events)
})
