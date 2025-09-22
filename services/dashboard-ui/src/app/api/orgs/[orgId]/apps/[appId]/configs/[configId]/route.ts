import { type NextRequest, NextResponse } from 'next/server'
import { getAppConfigById } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'appId' | 'configId'>
) {
  const { appId, configId, orgId } = await params
  const { searchParams } = new URL(request.url)
  const recurse = searchParams.get('recurse') === 'true'

  const response = await getAppConfigById({
    appId,
    appConfigId: configId,
    orgId,
    recurse,
  })
  return NextResponse.json(response)
}
