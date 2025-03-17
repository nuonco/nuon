import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getInstallComponentDeploys } from '@/lib'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, installId, __, componentId] = req.url
    .split('/')
    .slice(4, 9)
  
  let deploys = []
  try {
    deploys = await getInstallComponentDeploys({
      orgId,
      installId,
      componentId,
    })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(deploys)
})
