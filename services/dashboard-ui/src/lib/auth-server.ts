'use server'

import { cookies } from 'next/headers'
import { auth0 } from '@/lib/auth'
import { USE_AUTH_SERVICE } from '@/configs/auth'
import { API_URL } from '@/configs/api'
import type { IUser } from '@/types/dashboard.types'
import type { TMe, TAPIResponse } from '@/types'

interface ISession {
  user: IUser
  accessToken?: string
  accessTokenExpiresAt?: number
  [key: string]: any
}

// Helper function to get auth token from X-Nuon-Auth cookie
async function getAuthTokenFromCookie(): Promise<string | null> {
  const cookieStore = await cookies()
  return cookieStore.get('X-Nuon-Auth')?.value || null
}

// Direct API call function to avoid circular dependency with main api() function
async function callAuthServiceAPI<T>(path: string): Promise<TAPIResponse<T>> {
  let response: Response | undefined
  try {
    const token = await getAuthTokenFromCookie()
    
    const fetchOpts: RequestInit = {
      cache: 'no-store',
      method: 'GET',
      headers: {
        Accept: 'application/json',
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      signal: AbortSignal.timeout(10000),
    }

    response = await fetch(`${API_URL}/v1/${path}`, fetchOpts)

    // Convert headers to a plain object for serialization
    const headersObj = Object.fromEntries(response.headers.entries())

    let data = null
    const contentType = response.headers.get('content-type')
    const contentLength = response.headers.get('content-length')

    // Only try to parse JSON if there's actually content
    if (contentLength !== '0' && contentType?.includes('application/json')) {
      const text = await response.text()
      if (text) {
        try {
          data = JSON.parse(text)
        } catch (parseError) {
          console.warn('Failed to parse response as JSON:', parseError)
          data = text
        }
      }
    }

    if (response.ok) {
      return {
        data,
        error: null,
        status: response.status,
        headers: headersObj,
      }
    } else {
      return {
        data: null,
        error: data || {
          error: 'Unknown error',
          description: 'No error details provided',
        },
        status: response.status,
        headers: headersObj,
      }
    }
  } catch (error) {
    const errorHeadersObj = response
      ? Object.fromEntries(response.headers.entries())
      : {}

    return {
      data: null,
      error: {
        description: 'An unexpected error occurred while fetching data.',
        error: error instanceof Error ? error.message : 'Unknown Error',
        user_error: false,
      },
      status: 500,
      headers: errorHeadersObj,
    }
  }
}

// Transform TMe response from auth service to IUser format
function transformMeToUser(me: TMe): IUser {
  const firstIdentity = me.identities?.[0]
  return {
    sub: me.id,
    email: me.email,
    name: firstIdentity?.name,
    picture: firstIdentity?.picture,
  }
}

export async function getSession(): Promise<ISession | null | undefined> {
  if (USE_AUTH_SERVICE) {
    try {
      const { data: me, error } = await callAuthServiceAPI<TMe>('auth/me')
      if (error || !me) {
        return null // Not authenticated or API error
      }

      const accessToken = await getAuthTokenFromCookie()
      const user = transformMeToUser(me)

      return {
        user,
        accessToken,
      }
    } catch (error) {
      console.error('Error fetching user session from auth service:', error)
      return null
    }
  }
  
  // Use Auth0 for now
  return await auth0.getSession()
}

export async function getAccessToken(): Promise<string | null> {
  if (USE_AUTH_SERVICE) {
    return await getAuthTokenFromCookie()
  }

  // Extract token from Auth0 session
  const session = await auth0.getSession()
  return session?.accessToken || null
}

export async function getUserProfile(): Promise<IUser | null> {
  if (USE_AUTH_SERVICE) {
    try {
      const { data: me, error } = await callAuthServiceAPI<TMe>('auth/me')
      if (error || !me) {
        return null // Not authenticated or API error
      }

      return transformMeToUser(me)
    } catch (error) {
      console.error('Error fetching user profile from auth service:', error)
      return null
    }
  }

  // Extract user from Auth0 session
  const session = await auth0.getSession()
  return session?.user || null
}

export async function isAuthenticated(): Promise<boolean> {
  if (USE_AUTH_SERVICE) {
    const token = await getAuthTokenFromCookie()
    return !!token // User is authenticated if token exists in cookie
  }

  // Check Auth0 session exists
  const session = await auth0.getSession()
  return !!session?.user
}

export async function hasAdminAccess(): Promise<boolean> {
  if (USE_AUTH_SERVICE) {
    try {
      const { data: me, error } = await callAuthServiceAPI<TMe>('auth/me')
      if (error || !me) {
        return false // Not authenticated or API error
      }

      // Check if user has @nuon.co email
      return me.email?.endsWith('@nuon.co') ?? false
    } catch (error) {
      console.error('Error checking admin access from auth service:', error)
      return false
    }
  }

  // Check if Auth0 user has @nuon.co email
  const session = await auth0.getSession()
  return session?.user?.email?.endsWith('@nuon.co') ?? false
}