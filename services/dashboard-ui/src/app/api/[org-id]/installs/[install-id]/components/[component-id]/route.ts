import { NextRequest, NextResponse } from 'next/server'
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { getInstallComponent } from '@/lib'

export const GET = withApiAuthRequired(async (req: NextRequest) => {
  const [orgId, _, installId, __, installComponentId] = req.url
    .split('/')
    .slice(4, 9)

  let installComponent = {}
  try {
    installComponent = await getInstallComponent({
      orgId,
      installId,
      installComponentId,
    })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(installComponent)
})
