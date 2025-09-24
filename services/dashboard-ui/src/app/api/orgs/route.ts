import { NextResponse, type NextRequest } from 'next/server'
import { getOrgs } from '@/lib'

export const GET = async (request: NextRequest) => {
  const { searchParams } = new URL(request.url)
  const limit = searchParams.get('limit') || undefined
  const offset = searchParams.get('offset') || undefined
  const q = searchParams.get('q') || undefined

  const response = await getOrgs({ limit, offset, q })

  return NextResponse.json(response)
}
