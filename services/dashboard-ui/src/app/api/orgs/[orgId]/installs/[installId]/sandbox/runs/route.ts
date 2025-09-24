import { type NextRequest, NextResponse } from 'next/server'
import { getInstallSandboxRuns } from '@/lib'
import type { TRouteProps } from '@/types'

export async function GET(
  request: NextRequest,
  { params }: TRouteProps<'orgId' | 'installId'>
) {
  const { installId, orgId } = await params
  const { searchParams } = new URL(request.url)

  const limit = searchParams.get('limit')
  const offset = searchParams.get('offset')

  const response = await getInstallSandboxRuns({
    installId,
    orgId,
    ...(limit && { limit }),
    ...(offset && { offset }),
  })

  return NextResponse.json(response)
}
