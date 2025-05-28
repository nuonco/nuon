import { getSession } from '@auth0/nextjs-auth0/edge'
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { API_URL } from '@/utils/configs'
import type { TOrg } from '@/types'

// TODO(nnnnat): refactor this mess
export default async function middleware(request: NextRequest) {
  const pathname = new URL(request.url).pathname
  const headers = new Headers(request.headers)
  const session = await getSession()
  // set origin url encase of login redirect
  headers.set('x-origin-path', pathname)

  if (session && pathname !== '/favicon.ico') {
    let redirectPath = '/request-access'

    if (
      pathname === '/admin/temporal' &&
      !session?.user?.email?.endsWith('@nuon.co')
    ) {
      return NextResponse.redirect(new URL('/', request.url))
    }

    if (pathname === '/' || pathname.split('/')[1] === 'beta') {
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

      return NextResponse.redirect(new URL(redirectPath, request.url), {
        headers,
      })
    }
  }

  return NextResponse.next({
    request: { headers },
  })
}
