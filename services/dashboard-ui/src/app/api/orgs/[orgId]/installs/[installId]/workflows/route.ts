import { type NextRequest, NextResponse } from 'next/server'
import { getInstallWorkflows } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId'>
) {
  const { orgId, installId } = await params
  const { searchParams } = new URL(request.url)
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('offset') || undefined
  const type = searchParams.get('type') || undefined
    const planonly = searchParams.get('planonly') || undefined

  const response = await getInstallWorkflows({
    orgId,
    installId,
    limit,
    offset,
    planonly: planonly === "true",
    type,
  })
  return NextResponse.json(response)
}
