import { type NextRequest, NextResponse } from 'next/server'
import { getVCSConnectionRepos } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  _: NextRequest,
  { params }: TRouteProps<'orgId' | 'connectionId'>
) {
  const { orgId, connectionId } = await params
  const response = await getVCSConnectionRepos({ orgId, connectionId })
  return NextResponse.json(response)
}
