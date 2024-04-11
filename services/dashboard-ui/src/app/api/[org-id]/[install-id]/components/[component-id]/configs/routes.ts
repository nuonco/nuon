import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { API_URL, getFetchOpts } from '@/utils'

export const GET = async (req: NextRequest) => {
  const [orgId, installId, _, componentId] = req.url.split('/').slice(4, 8)
  const data = await fetch(
    `${API_URL}/v1/components/${componentId}/configs`,
    await getFetchOpts(orgId)
  )
  const configs = await data.json()
  
  return NextResponse.json({})
}
