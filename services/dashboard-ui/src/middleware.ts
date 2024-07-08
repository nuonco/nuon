import { getSession } from '@auth0/nextjs-auth0/edge'
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

// This function can be marked `async` if using `await` inside
export default async function middleware(request: NextRequest) {
  const session = await getSession()

  // console.log('============================= middleware url =======================', new URL('/dashboard', request.nextUrl.origin))

  if (session && new URL(request.url).pathname === '/') {
    return NextResponse.redirect(new URL('/dashboard', request.url))
  }

  return NextResponse.next()
}
