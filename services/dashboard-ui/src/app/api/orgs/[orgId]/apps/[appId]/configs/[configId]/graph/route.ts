import { type NextRequest, NextResponse } from 'next/server'
import { getAppConfigGraph } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'appId' | 'configId'>
) {
  const { appId, configId, orgId } = await params

  const response = await getAppConfigGraph({
    appId,
    appConfigId: configId,
    orgId,
  })
  return NextResponse.json(response)
}
