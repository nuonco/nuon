import { getSession } from '@auth0/nextjs-auth0/edge'
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { API_URL } from '@/utils/configs'
import type { TOrg } from '@/types'

export default async function middleware(request: NextRequest) {
  const session = await getSession()

  if (session) {
    if (
      new URL(request.url).pathname === '/' ||
      new URL(request.url).pathname.split('/')[1] === 'beta'
    ) {
      let redirectPath = '/getting-started'
      const orgSession = request.cookies.get('org-session')

      const orgs: Array<TOrg> = await (
        await fetch(`${API_URL}/v1/orgs`, {
          cache: 'no-store',
          headers: {
            Authorization: `Bearer ${session?.accessToken}`,
            'Content-Type': 'application/json',
            'X-Nuon-Org-ID': '',
          },
        })
      ).json()

      if (
        orgSession &&
        orgs.length > 0 &&
        orgs.some((org) => org.id === orgSession?.value)
      ) {
        redirectPath = `/${orgSession?.value}/apps`
      } else if (orgs.length > 0) {
        redirectPath = `/${orgs[0].id}/apps`
      }

      return NextResponse.redirect(new URL(redirectPath, request.url))
    }
  }

  return NextResponse.next()
}
