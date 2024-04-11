import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

// This function can be marked `async` if using `await` inside
export function middleware(request: NextRequest) {
  let response

  // if authed

  // if orgs.length === 0
  // redirect to onboarding
  // else
  response = NextResponse.next()

  // else
  // redirect to login

  return response
}

// See "Matching Paths" below to learn more
export const config = {
  // matcher: '*',
}
