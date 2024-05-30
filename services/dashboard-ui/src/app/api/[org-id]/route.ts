import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getOrg  } from '@/lib'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId] = req.url.split('/').slice(4, 5)
  
  let org = {}
  try {
    org = await getOrg({ orgId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(org)
})
