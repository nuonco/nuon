import { type NextRequest, NextResponse } from 'next/server'
import { getAccountsByOrgId } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId'>
) {
  const { orgId } = await params
  const { searchParams } = new URL(request.url)
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('offset') || undefined

  const response = await getAccountsByOrgId({ limit, offset, orgId })
  return NextResponse.json(response)
}
