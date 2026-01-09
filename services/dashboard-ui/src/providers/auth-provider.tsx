'use client'

import { createContext } from 'react'
import { useUser } from '@auth0/nextjs-auth0/client'
import type { IUser } from '@/types/dashboard.types'

interface IAuthContext {
  user: IUser | null | undefined
  error?: Error
  isLoading: boolean
  isAdmin: boolean
  useAuthService: boolean
  authServiceUrl?: string
}

export const AuthContext = createContext<IAuthContext | undefined>(undefined)

// Auth0-based auth provider
function Auth0AuthProvider({ 
  children,
  authServiceUrl,
}: { 
  children: React.ReactNode
  authServiceUrl?: string
}) {
  const { user, error, isLoading } = useUser()
  const isAdmin = user?.email?.endsWith('@nuon.co') ?? false

  return (
    <AuthContext.Provider
      value={{
        user,
        error,
        isLoading,
        isAdmin,
        useAuthService: false,
        authServiceUrl,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

// Auth service-based auth provider  
function AuthServiceAuthProvider({ 
  children,
  initialUser,
  authServiceUrl,
}: { 
  children: React.ReactNode
  initialUser: IUser | null
  authServiceUrl?: string
}) {
  const isAdmin = initialUser?.email?.endsWith('@nuon.co') ?? false

  return (
    <AuthContext.Provider
      value={{
        user: initialUser,
        error: null,
        isLoading: false,
        isAdmin,
        useAuthService: true,
        authServiceUrl,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

// Main auth provider that chooses the right implementation
export function AuthProvider({ 
  children,
  useAuthService = false,
  initialUser = null,
  authServiceUrl
}: { 
  children: React.ReactNode
  useAuthService?: boolean
  initialUser?: IUser | null
  authServiceUrl?: string
}) {
  if (useAuthService) {
    return (
      <AuthServiceAuthProvider 
        initialUser={initialUser} 
        authServiceUrl={authServiceUrl}
      >
        {children}
      </AuthServiceAuthProvider>
    )
  } else {
    return (
      <Auth0AuthProvider authServiceUrl={authServiceUrl}>
        {children}
      </Auth0AuthProvider>
    )
  }
}