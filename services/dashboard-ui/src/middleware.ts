import { NextResponse, type NextRequest } from 'next/server'
import { auth0 } from '@/lib/auth'

export default async function middleware(request: NextRequest) {
  const { pathname } = new URL(request.url)
  const reqCookieNames = request.cookies.getAll().map((cookie) => cookie.name)
  
  const txnCookies = reqCookieNames. filter((cookie) => cookie.startsWith('__txn'))

  // Block new login attempts if transaction already exists
  if (pathname === '/api/auth/login') {
    if (txnCookies. length > 0) {
      const response = NextResponse.redirect(new URL('/', request.url))
      // Set a flag to indicate we're waiting for auth
      response.cookies. set('__auth_waiting', 'true', { 
        maxAge: 60, // 1 minute
        httpOnly: true,
        sameSite: 'lax'
      })
      return response
    }
  }

  const authResponse = await auth0.middleware(request)

  if (pathname === '/api/auth/callback') {
    // Clear the waiting flag after successful callback
    authResponse.cookies.delete('__auth_waiting')
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
    // Check if we're waiting for another tab's auth to complete
    const isWaiting = request.cookies.get('__auth_waiting')
    
    if (txnCookies. length > 0 || isWaiting) {
      // Don't redirect to login, just let the request through
      // This prevents creating new login flows
      return NextResponse.next()
    }

    const { origin } = new URL(request. url)
    return NextResponse. redirect(
      `${origin}/api/auth/login?returnTo=${pathname}`
    )
  }

  if (session) {
    // Clear waiting flag if we have a session
    if (request.cookies.get('__auth_waiting')) {
      const response = NextResponse.next()
      response.cookies.delete('__auth_waiting')
      return response
    }
    
    if (
      pathname === '/admin/temporal' &&
      ! session?.user?.email?.endsWith('@nuon.co')
    ) {
      return NextResponse. redirect(new URL('/', request. url))
    }
  }

  return NextResponse.next()
}

export const config = {
  matcher:  [
    '/((?!_next/static|_next/image|favicon.ico|livez|readyz|\\.js|\\.css$|api/ddp|api/ctl-api|_app|admin/temporal-codec/decode).*)',
  ],
}
