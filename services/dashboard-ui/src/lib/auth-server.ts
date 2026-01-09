import { auth0 } from '@/lib/auth'
import { USE_AUTH_SERVICE } from '@/configs/auth'
import type { IUser } from '@/types/dashboard.types'

interface ISession {
  user: IUser
  accessToken?: string
  accessTokenExpiresAt?: number
  [key: string]: any
}

export async function getSession(): Promise<ISession | null | undefined> {
  if (USE_AUTH_SERVICE) {
    // TODO: Implement auth service session retrieval
    throw new Error('Auth service not yet implemented')
  }
  
  // Use Auth0 for now
  return await auth0.getSession()
}

export async function getAccessToken(): Promise<string | null> {
  if (USE_AUTH_SERVICE) {
    // TODO: Implement auth service token retrieval
    throw new Error('Auth service not yet implemented')
  }

  // Extract token from Auth0 session
  const session = await auth0.getSession()
  return session?.accessToken || null
}

export async function getUserProfile(): Promise<IUser | null> {
  if (USE_AUTH_SERVICE) {
    // TODO: Implement auth service user profile retrieval
    throw new Error('Auth service not yet implemented')  
  }

  // Extract user from Auth0 session
  const session = await auth0.getSession()
  return session?.user || null
}

export async function isAuthenticated(): Promise<boolean> {
  if (USE_AUTH_SERVICE) {
    // TODO: Implement auth service authentication check
    throw new Error('Auth service not yet implemented')
  }

  // Check Auth0 session exists
  const session = await auth0.getSession()
  return !!session?.user
}

export async function hasAdminAccess(): Promise<boolean> {
  if (USE_AUTH_SERVICE) {
    // TODO: Implement auth service admin check
    throw new Error('Auth service not yet implemented')
  }

  // Check if Auth0 user has @nuon.co email
  const session = await auth0.getSession()
  return session?.user?.email?.endsWith('@nuon.co') ?? false
}