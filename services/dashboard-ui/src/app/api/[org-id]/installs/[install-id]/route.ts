import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import type { TInstall } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, installId] = req.url.split('/').slice(4, 7)
  
  let install = {}
  try {
    const data = await fetch(
      `${API_URL}/v1/installs/${installId}`,
      await getFetchOpts(orgId)
    )
    install = await data.json() as TInstall

  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(install)
})
