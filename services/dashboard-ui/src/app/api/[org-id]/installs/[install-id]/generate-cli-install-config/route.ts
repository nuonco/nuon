import { NextRequest } from 'next/server'
import { API_URL } from '@/configs/api'
import { TRouteRes } from '@/app/api/[org-id]/types'
import { getFetchOpts } from '@/utils/get-fetch-opts'

export const GET = async (
  req: NextRequest,
  { params }: TRouteRes<'org-id' | 'install-id'>
) => {
  const { ['org-id']: orgId, ['install-id']: installId } = await params

  var resp = await fetch(
    `${API_URL}/v1/installs/${installId}/generate-cli-install-config`,
    await getFetchOpts(orgId, {}, 30000)
  )

  return resp
}
