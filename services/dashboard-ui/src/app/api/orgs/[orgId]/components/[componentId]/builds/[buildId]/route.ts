import { type NextRequest, NextResponse } from 'next/server'
import { getComponentBuild } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'componentId' | 'buildId'>
) {
  const { orgId, componentId, buildId } = await params

  const response = await getComponentBuild({ orgId, componentId, buildId })
  return NextResponse.json(response)
}
