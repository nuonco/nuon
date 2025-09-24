// TODO(NNNNAT): Can remove this endpoint once we've moved to just default install-components
import { type NextRequest, NextResponse } from 'next/server'
import { api } from '@/lib/api'
import type { TRouteProps, TInstallComponentSummary } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

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

  const response = await api<TInstallComponentSummary[]>({
    orgId,
    path: `installs/${installId}/components/summary${buildQueryParams({ offset, limit, q, types })}`,
    headers: {
      'x-nuon-pagination-enabled': true,
    },
  })

  return NextResponse.json(response)
}
