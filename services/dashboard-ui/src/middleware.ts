import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { auth0 } from '@/utils/auth'
import { API_URL } from '@/utils/configs'
import type { TOrg } from '@/types'

// TODO(nnnnat): refactor this mess
export default async function middleware(request: NextRequest) {
  const authResponse = await auth0.middleware(request)
  const { pathname } = new URL(request.url)
  const headers = new Headers(request.headers)
  const reqCookieNames = request.cookies.getAll().map((cookie) => cookie.name)

  // set origin url encase of login redirect
  // TODO(nnnat): don't think we need this anymore
  headers.set('x-origin-path', pathname)

  if (request.nextUrl.pathname === '/api/auth/login') {
    // This is a workaround for this issue: https://github.com/auth0/nextjs-auth0/issues/1917
    // The auth0 middleware sets some transaction cookies that are not deleted after the login flow completes.
    // This causes stale cookies to be used in subsequent requests and eventually causes the request header to be rejected because it is too large.
    reqCookieNames.forEach((cookie) => {
      if (cookie.startsWith('__txn')) {
        authResponse.cookies.delete(cookie)
      }
    })
  }

  // if path starts with /auth, let the auth middleware handle it
  if (
    pathname.startsWith('/auth') ||
    pathname.startsWith('/api/auth') ||
    pathname.startsWith('/v2/logout')
  ) {
    return authResponse
  }

  const session = await auth0.getSession(request)

  if (!session && pathname !== '/') {
    const { origin } = new URL(request.url)
    return NextResponse.redirect(
      `${origin}/api/auth/login?returnTo=${pathname}`
    )
  }

  if (session && pathname !== '/favicon.ico') {
    let redirectPath = '/request-access'

    if (
      pathname === '/admin/temporal' &&
      !session?.user?.email?.endsWith('@nuon.co')
    ) {
      return NextResponse.redirect(new URL('/', request.url))
    }

    // set active org
    // TODO(nnnat): move this to the app/page file
    if (pathname === '/' || pathname.split('/')[1] === 'beta') {
      const orgSession = request.cookies.get('org-session')

      const orgs: Array<TOrg> = await (
        await fetch(`${API_URL}/v1/orgs`, {
          cache: 'no-store',
          headers: {
            Authorization: `Bearer ${session?.tokenSet?.accessToken}`,
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

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|livez|readyz|\\.js|\\.css$|api/ddp|api/ctl-api|_app|admin/temporal-codec/decode).*)',
  ],
}
