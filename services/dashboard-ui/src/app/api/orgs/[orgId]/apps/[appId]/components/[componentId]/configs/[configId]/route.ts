import { type NextRequest, NextResponse } from 'next/server'
import { getComponentConfig } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'appId' | 'componentId' | 'configId'>
) {
  const { appId, componentId, configId, orgId } = await params

  const response = await getComponentConfig({
    appId,
    componentId,
    configId,
    orgId,
  })
  return NextResponse.json(response)
}