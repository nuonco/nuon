import { NextRequest, NextResponse } from 'next/server'
import { getInstall } from '@/lib'
import { TRouteRes } from '@/app/api/[org-id]/types'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'install-id'>
) => {
  const { ['org-id']: orgId, ['install-id']: installId } = await params

  let install = {}
  try {
    install = await getInstall({ orgId, installId })
  } catch (error) {
    console.error(error)
  }

  return NextResponse.json(install)
}
