import { NextResponse, type NextRequest } from 'next/server'
import { USE_AUTH_SERVICE } from '@/configs/auth'
import { auth0 } from '@/lib/auth'

export default async function middleware(request: NextRequest) {
  const { pathname } = new URL(request.url)

  if (USE_AUTH_SERVICE) {
    // eslint-disable-next-line no-console
    console.log('using nuon auth service')
  } else {
    // eslint-disable-next-line no-console

    const authResponse = await auth0.middleware(request)
    const reqCookieNames = request.cookies.getAll().map((cookie) => cookie.name)

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

    if (session) {
      if (
        pathname === '/admin/temporal' &&
        !session?.user?.email?.endsWith('@nuon.co')
      ) {
        return NextResponse.redirect(new URL('/', request.url))
      }
    }
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|livez|readyz|\\.js|\\.css$|api/ddp|api/ctl-api|_app|admin/temporal-codec/decode).*)',
  ],
}
