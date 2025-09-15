import { type NextRequest, NextResponse } from 'next/server'
import { getComponentBuildById } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'componentId' | 'buildId'>
) {
  const { orgId, componentId, buildId } = await params

  const response = await getComponentBuildById({ orgId, componentId, buildId })
  return NextResponse.json(response)
}
