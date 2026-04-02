import { createContext } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getMe } from '@/lib/ctl-api/auth/get-me'
import { isNuonSession } from '@/utils/session-utils'
import type { IUser, TMe } from '@/types'

interface IAuthContext {
  user: IUser | null
  isAuthenticated: boolean
  isAdmin: boolean
  isLoading: boolean
  error: unknown
}

export const AuthContext = createContext<IAuthContext | undefined>(undefined)

function meToUser(me: TMe): IUser {
  const firstIdentity = me.identities?.[0]
  return {
    sub: me.id,
    email: me.email,
    name: firstIdentity?.name,
    picture: firstIdentity?.picture,
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { data: me, isLoading, error } = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: getMe,
    staleTime: Infinity,
    refetchOnWindowFocus: false,
    retry: false,
  })

  const user = me ? meToUser(me) : null

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isAdmin: !!user && isNuonSession(user),
        isLoading,
        error,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}
