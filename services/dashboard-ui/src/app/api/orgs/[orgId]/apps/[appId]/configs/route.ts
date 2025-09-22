import { type NextRequest, NextResponse } from 'next/server'
import { getAppConfigs } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'appId'>
) {
  const { appId, orgId } = await params
  const { searchParams } = new URL(request.url)
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('offset') || undefined

  const response = await getAppConfigs({ appId, orgId, limit, offset })
  return NextResponse.json({
    ...response,
    headers: Object.fromEntries(response.headers.entries()),
  })
}
