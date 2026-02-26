import { createContext, useEffect, useState } from 'react'
import { useConfig } from '@/hooks/use-config'
import type { TMe } from '@/types/ctl-api.types'
import type { IUser } from '@/types/dashboard.types'

interface IAuthContext {
  user: IUser | null
  isAuthenticated: boolean
  isLoading: boolean
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
  const { apiUrl } = useConfig()
  const [user, setUser] = useState<IUser | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    fetch(`${apiUrl}/v1/auth/me`, {
      credentials: 'include',
    })
      .then((res) => (res.ok ? res.json() : null))
      .then((me: TMe | null) => {
        setUser(me ? meToUser(me) : null)
      })
      .catch(() => {
        setUser(null)
      })
      .finally(() => {
        setIsLoading(false)
      })
  }, [apiUrl])

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isLoading,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}
