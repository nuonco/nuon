import { type NextRequest, NextResponse } from 'next/server'
import { createRunnerBootstrapToken } from '@/lib'
import type { TRouteProps } from '@/types'

export async function POST(
  _: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId'>
) {
  const { installId, orgId } = await params
  const response = await createRunnerBootstrapToken({ installId, orgId })
  return NextResponse.json(response)
}
