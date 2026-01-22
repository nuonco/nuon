import { type NextRequest, NextResponse } from 'next/server'
import { getTerraformWorkspaceLock } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  _request: NextRequest,
  { params }: TRouteProps<'orgId' | 'workspaceId'>
) {
  const { orgId, workspaceId } = await params

  const response = await getTerraformWorkspaceLock({ workspaceId, orgId })

  return NextResponse.json(response)
}
