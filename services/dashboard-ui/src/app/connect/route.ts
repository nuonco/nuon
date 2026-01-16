'use server'

import { NextRequest, NextResponse } from 'next/server'
import { api } from '@/lib/api'
import { USE_AUTH_SERVICE, APP_URL } from '@/configs/auth'
import type { TVCSConnection } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const GET = async (req: NextRequest) => {
  const github_install_id = req.nextUrl.searchParams.get('installation_id')
  const org_id = req.nextUrl.searchParams.get('state')

  const { data, error } = await api<TVCSConnection>({
    method: 'POST',
    path: 'vcs/connection-callback',
    body: {
      github_install_id,
      org_id,
    },
  })

  const params = buildQueryParams({ 'vcs-connected': data?.id })

  return NextResponse.redirect(
    new URL(
      `/${org_id}/apps${params}`,
      USE_AUTH_SERVICE ? APP_URL : process.env?.AUTH0_BASE_URL
    )
  )
}
