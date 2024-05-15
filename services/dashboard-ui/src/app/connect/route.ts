'use server'

import { NextRequest, NextResponse } from 'next/server'
import { API_URL, getFetchOpts } from '@/utils'

export const GET = async (req: NextRequest) => {
  const github_install_id = req.nextUrl.searchParams.get('installation_id')
  const org_id = req.nextUrl.searchParams.get('state')

  await fetch(`${API_URL}/v1/vcs/connection-callback`, {
    ...(await getFetchOpts(org_id)),
    body: JSON.stringify({
      github_install_id,
      org_id,
    }),
    method: 'POST',
  }).catch(console.error)

  return NextResponse.redirect(new URL(`/dashboard/${org_id}`, req.url))
}
