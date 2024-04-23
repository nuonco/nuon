import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { API_URL, getFetchOpts } from '@/utils'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, __, componentId] = req.url.split('/').slice(4, 8)

  let config = {}
  try {
    const data = await fetch(
      `${API_URL}/v1/components/${componentId}/configs`,
      await getFetchOpts(orgId)
    )
    const configs = await data.json()
    config = configs?.[0] || config
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(config)
})
