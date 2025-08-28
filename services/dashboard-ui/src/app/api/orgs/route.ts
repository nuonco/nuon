import { NextResponse, type NextRequest } from 'next/server'
import { buildQueryParams } from '@/utils/build-query-params'
import { nueQueryData } from '@/utils'

export const GET = async (request: NextRequest) => {
  const { searchParams } = new URL(request.url)
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('offset') || undefined
  const q = searchParams.get('q') || undefined

  const res = await nueQueryData({
    path: `orgs${buildQueryParams({ limit, offset, q })}`,
    headers: {
      'x-nuon-pagination-enabled': true,
    },
  })

  return NextResponse.json({
    ...res,
    headers: Object.fromEntries(res.headers.entries()),
  })
}
