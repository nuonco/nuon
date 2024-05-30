import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getInstall } from '@/lib'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, installId] = req.url.split('/').slice(4, 7)
  
  let install = {}
  try {
    install = await getInstall({ orgId, installId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(install)
})
