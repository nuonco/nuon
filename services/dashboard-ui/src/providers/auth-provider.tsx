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
}

export const AuthContext = createContext<IAuthContext | undefined>(undefined)

export function AuthProvider({ 
  children,
  useAuthService = false
}: { 
  children: React.ReactNode
  useAuthService?: boolean
}) {
  // For now, we only use Auth0 - auth service implementation will come later
  const { user, error, isLoading } = useUser()

  // Check if user is Nuon admin
  const isAdmin = user?.email?.endsWith('@nuon.co') ?? false

  return (
    <AuthContext.Provider
      value={{
        user,
        error,
        isLoading,
        isAdmin,
        useAuthService,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}