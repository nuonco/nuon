import { type NextRequest, NextResponse } from 'next/server'
import { api } from '@/lib/api'
import type { TRouteProps } from '@/types'

export async function GET(
  _: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId'>
) {
  const { installId, orgId } = await params
  const response = await api({
    orgId,
    path: `installs/${installId}/generate-cli-install-config`
  })
  return NextResponse.json(response)
}
