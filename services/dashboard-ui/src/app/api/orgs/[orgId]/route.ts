import { type NextRequest, NextResponse } from 'next/server'
import { getOrg } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(_: NextRequest, { params }: TRouteProps<'orgId'>) {
  const { orgId } = await params
  const response = await getOrg({ orgId })
  return NextResponse.json(response)
}
