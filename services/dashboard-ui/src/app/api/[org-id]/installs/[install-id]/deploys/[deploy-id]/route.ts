import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getInstallDeploy } from '@/lib'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, installId, __, installDeployId] = req.url
    .split('/')
    .slice(4, 9)

  let installComponent = {}
  try {
    installComponent = await getInstallDeploy({
      orgId,
      installId,
      installDeployId,
    })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(installComponent)
})
