import { type NextRequest, NextResponse } from 'next/server'
import { getOrgById } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(_: NextRequest, { params }: TRouteProps<'orgId'>) {
  const { orgId } = await params
  const response = await getOrgById({ orgId })
  return NextResponse.json(response)
}
