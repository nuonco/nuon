import { NextRequest } from 'next/server'
import { API_URL } from '@/configs/api'
import { TRouteRes } from '@/app/api/[org-id]/types'
import { getFetchOpts } from '@/utils/get-fetch-opts'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'install-id'>
) => {
  const { ['org-id']: orgId, ['install-id']: installId } = await params

  const startTS = req.nextUrl.searchParams.get('start')
  const endTS = req.nextUrl.searchParams.get('end')

  var resp = await fetch(
    `${API_URL}/v1/installs/${installId}/audit_logs?start=${startTS}&end=${endTS}`,
    await getFetchOpts(orgId, {}, 30000)
  )

  return resp
}
