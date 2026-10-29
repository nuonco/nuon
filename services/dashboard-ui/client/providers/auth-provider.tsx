import { createContext, useState, useCallback, useEffect } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { getMe } from '@/lib/ctl-api/auth/get-me'
import { syncMarketingConsent } from '@/lib/marketing-consent'
import { isNuonSession } from '@/utils/session-utils'
import type { IUser, TMe } from '@/types'

const DEMO_MODE_KEY = 'nuon_demo_mode'

interface IAuthContext {
  user: IUser | null
  isAuthenticated: boolean
  isAdmin: boolean
  isNuonEmployee: boolean
  isLoading: boolean
  error: unknown
  demoMode: boolean
  toggleDemoMode: () => void
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
    placeholderData: keepPreviousData,
    queryKey: ['auth', 'me'],
    queryFn: getMe,
    staleTime: Infinity,
    refetchOnWindowFocus: false,
    retry: false,
  })

  const [demoMode, setDemoMode] = useState(
    () => localStorage.getItem(DEMO_MODE_KEY) === 'true'
  )

  const toggleDemoMode = useCallback(() => {
    setDemoMode((prev) => {
      const next = !prev
      localStorage.setItem(DEMO_MODE_KEY, String(next))
      return next
    })
  }, [])

  useEffect(() => {
    if (me?.id) syncMarketingConsent()
  }, [me?.id])

  const user = me ? meToUser(me) : null
  const isNuonEmployee = !!user && isNuonSession(user)

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isNuonEmployee,
        isAdmin: isNuonEmployee && !demoMode,
        isLoading,
        error,
        demoMode,
        toggleDemoMode,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}
