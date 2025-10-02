import { type NextRequest, NextResponse } from 'next/server'
import { getInstallComponents } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId'>
) {
  const { orgId, installId } = await params
  const { searchParams } = new URL(request.url)
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('offset') || undefined
  const q = searchParams.get('q') || undefined
  const types = searchParams.get('types') || undefined

  const response = await getInstallComponents({
    installId,
    orgId,
    limit,
    offset,
    q,
    types,
  })

  return NextResponse.json(response)
}
