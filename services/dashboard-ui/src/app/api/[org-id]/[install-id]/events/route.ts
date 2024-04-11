import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { API_URL, getFetchOpts } from '@/utils'

export const GET = async (req: NextRequest) => {
  const [orgId, installId] = req.url.split('/').slice(4, 6)
  const data = await fetch(
    `${API_URL}/v1/installs/${installId}/events`,
    await getFetchOpts(orgId)
  )

  return NextResponse.json(await data.json())
}
